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

// STRUCT:
type OrderDetailsClensing struct {
	conns   infra.DbConnection
	Result  Result
	Details []*OrderDetailPiece
}

type OrderDetailPiece struct {
	OrderNo       int
	OrderDetailNo int
	status        Status
	approved      bool
	message       string
}

// FUNCTION: ステータスの変更
func (p *OrderDetailPiece) setStatus(status Status, approved bool) {
	p.status = judgeStatus(p.status, status)
	p.approved = p.approved && approved
}

// FUNCTION: メッセージの追加
func (p *OrderDetailPiece) addMessage(msg string, id string) {
	p.message = genMessage(p.message, msg, id)
}

// FUNCTION:
func NewOrderDetails(conns infra.DbConnection) OrderDetailsClensing {
	s := time.Now()

	// INFO: 固定値設定
	cs := OrderDetailsClensing{conns: conns, Result: Result{
		TableNameJp: "受注明細",
		TableNameEn: "order_details",
	}}
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

			bar.Increment()
		}
	}
	bar.Finish()
}

// FUNCTION: レコード毎のチェック
func (cs *OrderDetailsClensing) checkAndClensing(record *legacy.OrderDetail) *OrderDetailPiece {
	// INFO: piece
	piece := OrderDetailPiece{
		OrderNo:       record.OrderNo,
		OrderDetailNo: record.OrderDetailNo,
		status:        NO_CHANGE,
		approved:      true,
	}

	// PROCESS: #4-01: shipping_flag/ cancel_flagの両方がtrueの場合、移行対象から除外する。
	if record.ShippingFlag && record.CanceledFlag {
		piece.setStatus(REMOVE, true)
		piece.addMessage("shipping_flag(出荷済フラグ)、canceled_flag(キャンセルフラグ)がいずれも `true`。移行対象から除外。", "#4-01")
	}

	// PROCESS: #4-02: order_noが[受注]に存在しない場合、移行対象から除外する。
	ok1, _ := legacy.OrderExists(ctx, cs.conns.LegacyDB, record.OrderNo)
	if !ok1 {
		piece.setStatus(REMOVE, false)
		piece.addMessage(fmt.Sprintf("order_no(受注番号) が[受注]に存在しない。移行対象から除外 `%d`。", record.OrderNo), "#4-02")
	}

	// PROCESS: #4-03: product_nameが[商品]に存在しない場合、移行対象から除外する。
	// TODO: クレンジング処理未記載の状況際限のためコメントアウト
	// ok2, _ := legacy.ProductExists(ctx, cs.conns.LegacyDB, record.ProductName)
	// if !ok2 {
	// 	piece.setStatus(REMOVE, true)
	// 	piece.addMessage(fmt.Sprintf("product_name(商品名) が[商品]に存在しない。移行対象から除外 `%s`。", record.ProductName), "#4-03")
	// }

	return &piece
}

// FUNCTION: レコード毎のクレンジング後データ登録
func (cs *OrderDetailsClensing) saveData(record *legacy.OrderDetail, piece *OrderDetailPiece) {
	// PROCESS: REMOVEDの場合はDBに登録しない
	if piece.status == REMOVE {
		cs.setResult(piece)
		return
	}

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
		WOrderNo:          "RO-9000001", //FIXME:
		CreatedBy:         OPERATION_USER,
		UpdatedBy:         OPERATION_USER,
	}
	err := rec.Insert(ctx, cs.conns.WorkDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納、未承認扱い)
	if err != nil {
		piece.setStatus(REMOVE, false)
		piece.addMessage(fmt.Sprintf("%v", err), "")
	}
	cs.setResult(piece)
}

// FUNCTION: クレンジング結果の登録
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
			piece.status,
			approveStr(piece.approved),
			piece.message,
		)
	}

	return msg
}
