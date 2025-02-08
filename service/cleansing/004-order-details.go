/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: 詳細メッセージ
type OrderDetailMsg struct {
	OrderNo       int
	OrderDetailNo int
	bp            *Piece
}

func NewOrderDetailMsg(orderNo, orderDetailNo int) *OrderDetailMsg {
	return &OrderDetailMsg{
		OrderNo:       orderNo,
		OrderDetailNo: orderDetailNo,
		bp:            NewPiece(),
	}
}

// STRUCT: レコード
type OrderDetailRecord struct {
	record     legacy.OrderDetail
	msg        *OrderDetailMsg
	setTo      *[]OrderDetailMsg
	orderNoGen *OrderNoGenerator
}

// FUNCTION: 更新
func (r *OrderDetailRecord) checkAndPersist(ctx infra.AppCtx, db *sql.DB, refData *RefData) Piece {

	// PROCESS: check #4-01:
	r.checkShippingAndCanceledFlag()

	// PROCESS:INFO: check #4-02:
	r.checkOrderNo(refData.OrderNoSet)

	// PROCESS: check #4-03:
	// TODO: クレンジング処理未記載の状況際限のためコメントアウト
	// r.checkProductName(refData.ProductNameSet)

	// PROCESS: REMOVE判定時は登録なし
	if !r.msg.bp.isRemove() {
		r.persiste(ctx, db)
	}

	// PROCESS: REMOVE/MODIFY判定時は詳細情報の出力あり
	if r.msg.bp.isWarn() {
		*r.setTo = append(*r.setTo, *r.msg)
	}

	return *r.msg.bp
}

// FUNCTION: データ登録
func (r *OrderDetailRecord) persiste(ctx infra.AppCtx, db *sql.DB) {
	// INFO: 受注番号の採番
	wOrderNo := r.orderNoGen.generate(r.record.OrderNo, r.record.ProductName, r.record.SellingPrice, r.record.CostPrice)

	// PROCESS: データ登録
	rec := clean.OrderDetail{
		OrderNo:           r.record.OrderNo,
		OrderDetailNo:     r.record.OrderDetailNo,
		ProductName:       r.record.ProductName,
		ReceivingQuantity: r.record.ReceivingQuantity,
		ShippingFlag:      r.record.ShippingFlag,
		CancelFlag:        r.record.CanceledFlag,
		SellingPrice:      r.record.SellingPrice,
		CostPrice:         r.record.CostPrice,
		WOrderNo:          wOrderNo,
		CreatedBy:         ctx.OperationUser,
		UpdatedBy:         ctx.OperationUser,
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		r.msg.bp.dbError(err)
	}
}

// FUNCTION: #4-01(REMOVE): shipping_flag/ cancel_flagの両方がtrueの場合、移行対象から除外する。
func (r *OrderDetailRecord) checkShippingAndCanceledFlag() {
	const ID = "#4-01"
	if r.record.ShippingFlag && r.record.CanceledFlag {
		r.msg.bp.removed().addMessage(
			"shipping_flag(出荷済フラグ)、canceled_flag(キャンセルフラグ)がいずれも `true`です。【除外】", ID)
	}
}

// FUNCTION: #4-02(REMOVE): order_noが[受注]に存在しない場合、移行対象から除外する。
func (r *OrderDetailRecord) checkOrderNo(OrderNoSet map[int]struct{}) {
	const ID = "#4-02"
	_, exist := OrderNoSet[r.record.OrderNo]
	if !exist {
		r.msg.bp.approveStay() //TODO: 承認確認中
		r.msg.bp.removed().addMessage("order_no(受注番号) が[受注]に存在しません。【除外】", ID)
	}
}

// FUNCTION: #4-03(REMOVE): product_nameが[商品]に存在しない場合、移行対象から除外する。
func (r *OrderDetailRecord) checkProductName(ProductNameSet map[string]struct{}) {
	const ID = "#4-03"
	productName := r.record.ProductName
	_, exist := ProductNameSet[productName]
	if !exist {
		r.msg.bp.removed().addMessage(fmt.Sprintf("product_name(商品名) が[商品]に存在しません`%s`。【除外】", productName), ID)
	}
}

// STRUCT: コマンド
type OrderDetailsCmd struct {
	details    []OrderDetailMsg
	orderNoGen *OrderNoGenerator
}

// FUNCTION: New
func NewOrderDetailsCmd() *OrderDetailsCmd {
	return &OrderDetailsCmd{orderNoGen: NewOrderNoGenerator()}
}

// FUNCTION: テーブル名設定
func (cmd *OrderDetailsCmd) getTableInfo() TableInfo {
	return TableInfo{
		tableJp: "受注明細",
		tableEn: legacy.TableNames.OrderDetails,
	}
}

// FUNCTION: 入力データ量
func (cmd *OrderDetailsCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := legacy.OrderDetails().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *OrderDetailsCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	records, err := legacy.OrderDetails(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &OrderDetailRecord{
			record:     *record,
			msg:        NewOrderDetailMsg(record.OrderNo, record.OrderDetailNo),
			setTo:      &cmd.details,
			orderNoGen: cmd.orderNoGen,
		}}
	}
	return results
}

// FUNCTION: 追加データ登録
func (r *OrderDetailsCmd) extInsert(ctx infra.AppCtx, db *sql.DB, refData *RefData) {}

// FUNCTION: 詳細メッセージの出力
func (cmd *OrderDetailsCmd) showDetails(ctx infra.AppCtx, tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | order_no | order_detail_no | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|--:|--:|---|:-:|:-:|---|\n"
	for i, piece := range cmd.details {
		msg += fmt.Sprintf("  | %d | %d | %d | … | %s | %s | %s |\n",
			i+1,
			piece.OrderNo,
			piece.OrderDetailNo,
			piece.bp.status,
			piece.bp.approve,
			piece.bp.msg,
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
