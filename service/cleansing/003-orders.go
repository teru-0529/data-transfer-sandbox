/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: 詳細メッセージ
type OrderMsg struct {
	OrderNo int
	bp      *Piece
}

func NewOrderMsg(orderNo int) *OrderMsg {
	return &OrderMsg{
		OrderNo: orderNo,
		bp:      NewPiece(),
	}
}

// STRUCT: レコード
type OrderRecord struct {
	record legacy.Order
	msg    *OrderMsg
	setTo  *[]OrderMsg
}

// FUNCTION: 更新
func (r *OrderRecord) checkAndPersist(ctx infra.AppCtx, db *sql.DB, refData *RefData) Piece {

	// PROCESS: check #3-01:
	r.checkOrderDate(ctx)

	// PROCESS: check #3-02:
	r.checkOrderPic(refData.OperatorNameSet)

	// PROCESS: REMOVE判定時は登録なし
	if !r.msg.bp.isRemove() {
		r.persiste(ctx, db)
	}

	// PROCESS: REMOVE/MODIFY判定時は詳細情報の出力あり
	if r.msg.bp.isWarn() {
		*r.setTo = append(*r.setTo, *r.msg)
	}

	// PROCESS:INFO: 正常登録時に[受注番号]登録
	if !r.msg.bp.isRemove() {
		refData.OrderNoSet[r.record.OrderNo] = struct{}{}
	}

	return *r.msg.bp
}

// FUNCTION: データ登録
func (r *OrderRecord) persiste(ctx infra.AppCtx, db *sql.DB) {
	// INFO: 日付型変換
	orderDate, _ := time.Parse(ctx.DateLayout, r.record.OrderDate)

	// PROCESS: データ登録
	rec := clean.Order{
		OrderNo:      r.record.OrderNo,
		OrderDate:    orderDate,
		OrderPic:     r.record.OrderPic,
		CustomerName: r.record.CustomerName,
		CreatedBy:    ctx.OperationUser,
		UpdatedBy:    ctx.OperationUser,
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		r.msg.bp.dbError(err)
	}
}

// FUNCTION: #3-01(MODIFY): order_dateが日付のフォーマットに合致しない場合は、"20250101"にクレンジングする。
func (r *OrderRecord) checkOrderDate(ctx infra.AppCtx) {
	const ID = "#3-01"
	const ORDER_DATE = "20250101"
	orderDate := r.record.OrderDate
	_, err := time.Parse(ctx.DateLayout, orderDate)
	if err != nil {
		r.record.OrderDate = ORDER_DATE
		r.msg.bp.modified().addMessage(
			fmt.Sprintf("order_date(受注日付) が日付フォーマットではありません`%s`。<br>【クレンジング】`%s`(固定値) にクレンジング。", orderDate, ORDER_DATE), ID)
	}
}

// FUNCTION: #3-02(MODIFY): order_picが[担当者]に存在しない場合は、"N/A"にクレンジングする。
func (r *OrderRecord) checkOrderPic(OperatorNameSet map[string]struct{}) {
	const ID = "#3-02"
	const ORDER_PIC = "N/A"
	orderPic := r.record.OrderPic
	_, exist := OperatorNameSet[orderPic]
	if !exist {
		r.record.OrderPic = ORDER_PIC
		r.msg.bp.modified().addMessage(
			fmt.Sprintf("order_pic(受注担当者名) が[担当者]として存在しません`%s`。<br>【クレンジング】`%s`(固定値) にクレンジング。", orderPic, ORDER_PIC), ID)
	}
}

// STRUCT: コマンド
type OrdersCmd struct {
	details []OrderMsg
}

// FUNCTION: New
func NewOrdersCmd() *OrdersCmd {
	return &OrdersCmd{}
}

// FUNCTION: テーブル名設定
func (cmd *OrdersCmd) getTableInfo() TableInfo {
	return TableInfo{
		tableJp: "受注",
		tableEn: legacy.TableNames.Orders,
	}
}

// FUNCTION: 入力データ量
func (cmd *OrdersCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := legacy.Orders().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *OrdersCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	records, err := legacy.Orders(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &OrderRecord{
			record: *record,
			msg:    NewOrderMsg(record.OrderNo),
			setTo:  &cmd.details,
		}}
	}
	return results
}

// FUNCTION: 追加データ登録
func (r *OrdersCmd) extInsert(ctx infra.AppCtx, db *sql.DB, refData *RefData) {}

// FUNCTION: 詳細メッセージの出力
func (cmd *OrdersCmd) showDetails(ctx infra.AppCtx, tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | order_no | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|--:|---|:-:|:-:|---|\n"
	for i, piece := range cmd.details {
		msg += fmt.Sprintf("  | %d | %d | … | %s | %s | %s |\n",
			i+1,
			piece.OrderNo,
			piece.bp.status,
			piece.bp.approve,
			piece.bp.msg,
		)
	}

	return msg
}
