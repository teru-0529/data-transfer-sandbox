/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"fmt"
	"log"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: FIXME:
type OrderDetailPiece struct {
	OrderNo       int
	OrderDetailNo int
	bp            *Piece
	status        Status
	approved      bool
	message       string
}

// FUNCTION: ステータスの変更 FIXME:
func (p *OrderDetailPiece) setStatus(status Status, approved bool) {
	p.status = judgeStatus(p.status, status)
	p.approved = p.approved && approved
}

// FUNCTION: メッセージの追加 FIXME:
func (p *OrderDetailPiece) addMessage(msg string, id string) {
	p.message = genMessage(p.message, msg, id)
}

// STRUCT:
type OrderDetailsClensing struct {
	conns     infra.DbConnection
	refData   *RefData
	Result    Result
	Details   []*OrderDetailPiece
	generator *OrderNoGenerator
}

// FUNCTION:
func NewOrderDetails(conns infra.DbConnection, refData *RefData) OrderDetailsClensing {
	s := time.Now()

	// INFO: 固定値設定
	cs := OrderDetailsClensing{
		conns:     conns,
		refData:   refData,
		Result:    Result{TableNameJp: "受注明細", TableNameEn: "order_details"},
		generator: NewOrderNoGenerator(),
	}
	log.Printf("[%s] table cleansing ...", cs.Result.TableNameEn)

	// PROCESS: 入力データ量
	cs.setEntryCount()

	// PROCESS: クレンジング先のtruncate
	_, err := queries.Raw(cs.Result.truncateSql()).ExecContext(ctx, cs.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}

	// PROCESS: 1行ごと処理
	cs.iterate()

	duration := time.Since(s).Seconds()
	cs.Result.duration = duration
	log.Printf("cleansing completed … %3.2fs\n", duration)
	return cs
}

// FUNCTION: 入力データ量
func (cs *OrderDetailsClensing) setEntryCount() {
	// INFO: Legacyテーブル名
	num, err := legacy.OrderDetails().Count(ctx, cs.conns.LegacyDB)
	if err != nil {
		log.Fatalln(err)
	}
	cs.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (cs *OrderDetailsClensing) iterate() {
	bar := pb.Default.Start(cs.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < cs.Result.sectionCount(); section++ {
		// INFO: Legacyテーブル名
		records, err := legacy.OrderDetails(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, cs.conns.LegacyDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎のチェック
			piece := cs.checkAndClensing(record)
			// PROCESS: レコード毎のクレンジング後データ登録
			cs.saveData(record, piece)

			// FIXME:
			// PROCESS: 処理結果の登録
			cs.Result.setResult(piece.bp)
			if piece.bp.isWarn() {
				cs.Details = append(cs.Details, piece)
			}

			bar.Increment()
		}
	}
	bar.Finish()
}

// FUNCTION: レコード毎のチェック
func (cs *OrderDetailsClensing) checkAndClensing(record *legacy.OrderDetail) *OrderDetailPiece {
	// INFO: piece FIXME:
	piece := OrderDetailPiece{
		OrderNo:       record.OrderNo,
		OrderDetailNo: record.OrderDetailNo,
		bp:            NewPiece(),
		status:        NO_CHANGE,
		approved:      true,
	}

	// PROCESS: #4-01: shipping_flag/ cancel_flagの両方がtrueの場合、移行対象から除外する。
	if record.ShippingFlag && record.CanceledFlag {
		piece.bp.removed().addMessage(
			"shipping_flag(出荷済フラグ)、canceled_flag(キャンセルフラグ)がいずれも `true`。移行対象から除外。", "#4-01")

		// FIXME:
		piece.setStatus(REMOVE, true)
		piece.addMessage("shipping_flag(出荷済フラグ)、canceled_flag(キャンセルフラグ)がいずれも `true`。移行対象から除外。", "#4-01")
	}

	// PROCESS: #4-02: order_noが[受注]に存在しない場合、移行対象から除外する。
	_, exist1 := cs.refData.OrderNoSet[record.OrderNo]
	if !exist1 {
		piece.bp.approveStay() //TODO: 承認確認中
		piece.bp.removed().addMessage(
			fmt.Sprintf("order_no(受注番号) が[受注]に存在しない。移行対象から除外 `%d`。", record.OrderNo), "#4-02")

		// FIXME:
		piece.setStatus(REMOVE, false)
		piece.addMessage(fmt.Sprintf("order_no(受注番号) が[受注]に存在しない。移行対象から除外 `%d`。", record.OrderNo), "#4-02")
	}

	// PROCESS: #4-03: product_nameが[商品]に存在しない場合、移行対象から除外する。
	// TODO: クレンジング処理未記載の状況際限のためコメントアウト
	// _, exist2 := cs.refData.ProductNameSet[record.ProductName]
	// if !exist2 {
	// 	piece.bp.removed().addMessage(
	// 		fmt.Sprintf("product_name(商品名) が[商品]に存在しない。移行対象から除外 `%s`。", record.ProductName), "#4-03")

	// 	// FIXME:
	// 	piece.setStatus(REMOVE, true)
	// 	piece.addMessage(fmt.Sprintf("product_name(商品名) が[商品]に存在しない。移行対象から除外 `%s`。", record.ProductName), "#4-03")
	// }

	return &piece
}

// FUNCTION: レコード毎のクレンジング後データ登録
func (cs *OrderDetailsClensing) saveData(record *legacy.OrderDetail, piece *OrderDetailPiece) {
	// PROCESS: REMOVEDの場合はDBに登録しない
	// if piece.status == REMOVE { FIXME:
	if piece.bp.isRemove() {
		// cs.setResult(piece)
		return
	}

	// INFO: 受注番号の採番
	wOrderNo := cs.generator.generate(record.OrderNo, record.ProductName, record.SellingPrice, record.CostPrice)

	// PROCESS: データ登録
	// INFO: cleanテーブル
	rec := clean.OrderDetail{
		OrderNo:           record.OrderNo,
		OrderDetailNo:     record.OrderDetailNo,
		ProductName:       record.ProductName,
		ReceivingQuantity: record.ReceivingQuantity,
		ShippingFlag:      record.ShippingFlag,
		CancelFlag:        record.CanceledFlag,
		SellingPrice:      record.SellingPrice,
		CostPrice:         record.CostPrice,
		WOrderNo:          wOrderNo,
		CreatedBy:         OPERATION_USER,
		UpdatedBy:         OPERATION_USER,
	}
	err := rec.Insert(ctx, cs.conns.WorkDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納、未承認扱い)
	if err != nil {
		piece.bp.dbError(err)
		// FIXME:
		// piece.bp.setStatus(REMOVE).setApprove(NOT_FINDED).addMessage(redFont(fmt.Sprintf("%v", err)), "")

		// FIXME:
		piece.setStatus(REMOVE, false)
		piece.addMessage(fmt.Sprintf("%v", err), "")
	}
	// FIXME:
	// cs.setResult(piece)
}

// FUNCTION: クレンジング結果の登録 FIXME:
func (cs *OrderDetailsClensing) setResult(piece *OrderDetailPiece) {
	switch piece.status {
	case NO_CHANGE:
		cs.Result.UnchangeCount++
	case MODIFY:
		cs.Result.ModifyCount++
		cs.Details = append(cs.Details, piece)
	case REMOVE:
		cs.Result.RemoveCount++
		cs.Details = append(cs.Details, piece)
	}
}

// FUNCTION: 詳細情報の出力
func (cs *OrderDetailsClensing) ShowDetails() string {
	if len(cs.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", cs.Result.TableName())

	msg += "  | # | order_no | order_detail_no | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|--:|--:|---|:-:|:-:|---|\n"
	for i, piece := range cs.Details {
		msg += fmt.Sprintf("  | %d | %d | %d | … | %s | %s | %s |\n",
			i+1,
			piece.OrderNo,
			piece.OrderDetailNo,
			piece.bp.status,
			piece.bp.approve,
			piece.bp.msg,

			// FIXME:
			// piece.status,
			// approveStr(piece.approved),
			// piece.message,
		)
	}

	return msg
}

// STRUCT: 受注番号ジェネレータ
type OrderNoGenerator struct {
	OrderNoMap    map[OrderNoGenKey]string //受注番号
	OrderCountMap map[OrderNoCountKey]int  //商品ごとの受注番号数
}

type OrderNoGenKey struct {
	orderNo      int
	productName  string
	sellingPrice int
	costPrice    int
}
type OrderNoCountKey struct {
	orderNo     int
	productName string
}

// FUNCTION: generatorの生成
func NewOrderNoGenerator() *OrderNoGenerator {
	return &OrderNoGenerator{
		OrderNoMap:    map[OrderNoGenKey]string{},
		OrderCountMap: map[OrderNoCountKey]int{},
	}
}

// FUNCTION: 受注番号の採番
func (gen *OrderNoGenerator) generate(orderNo int, productName string, sellingPrice int, costPrice int) string {
	genKey := OrderNoGenKey{orderNo: orderNo, productName: productName, sellingPrice: sellingPrice, costPrice: costPrice}
	countKey := OrderNoCountKey{orderNo: orderNo, productName: productName}
	_ = countKey
	// PROCESS: すでに管理されている場合は該当する受注番号を返す
	result, exist := gen.OrderNoMap[genKey]
	if exist {
		return result
	}

	// PROCESS: シーケンス番号を取得(存在しない場合は0)
	no, exist := gen.OrderCountMap[countKey]
	if !exist {
		gen.OrderCountMap[countKey] = 0
		no = 0
	}

	// PROCESS: 受注番号を構成しMapに格納
	result = fmt.Sprintf("RO-9%05d%d", orderNo, no)
	gen.OrderNoMap[genKey] = result
	gen.OrderCountMap[countKey]++

	return result
}
