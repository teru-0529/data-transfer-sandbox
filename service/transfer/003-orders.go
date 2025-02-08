/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package transfer

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/product/orders"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: 詳細メッセージ
type OrderMsg struct {
	OrderNo int
	bp      *Piece
}

// STRUCT: レコード
type OrderRecord struct {
	record  clean.WOrder
	details *[]OrderMsg
}

// FUNCTION: 更新
func (r *OrderRecord) persist(ctx infra.AppCtx, db *sql.DB) int {

	// PROCESS: 明細が存在しない場合
	if !r.record.Register.Bool {
		return r.setNoDetail()
	}

	// PROCESS: データ登録
	rec := orders.Order{
		OrderNo:             r.record.WOrderNo.String,
		OrderDate:           r.record.OrderDate.Time,
		OrderPic:            r.record.OperatorID.String,
		CustomerName:        r.record.CustomerName.String,
		TotalOrderPrice:     int(r.record.WTotalOrderPrice.Int64),
		RemainingOrderPrice: int(r.record.WRemainingOrderPrice.Int64),
		OrderStatus:         r.orderStatus(),
		CreatedBy:           ctx.OperationUser,
		UpdatedBy:           ctx.OperationUser,
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		return r.setError(err)
	}

	// PROCESS: 分割した場合
	if r.record.Logging.Bool {
		return r.setDevided(int(r.record.ChangeCount.Int64))
	}

	return 0
}

// FUNCTION: 受注ステータス
func (r *OrderRecord) orderStatus() string {
	if r.record.IsRemaining.Bool {
		return orders.OrderStatusWORK_IN_PROGRESS
	} else if r.record.IsShipped.Bool {
		return orders.OrderStatusCOMPLETED
	} else {
		return orders.OrderStatusCANCELED
	}
}

// FUNCTION: 更新(エラー)
func (r *OrderRecord) setError(err error) int {
	msg := OrderMsg{
		OrderNo: r.record.OrderNo.Int,
		bp:      errorPiece(err),
	}
	*r.details = append(*r.details, msg)
	return msg.bp.count
}

// FUNCTION: 更新(分割)
func (r *OrderRecord) setDevided(no int) int {
	msg := OrderMsg{
		OrderNo: r.record.OrderNo.Int,
		bp: modifiedPiece(
			fmt.Sprintf("明細に`販売単価`もしくは`商品原価`が一致しない同一の商品が存在するため、受注を [%d] 件に分割しました。", no+1),
			int(no),
		),
	}
	*r.details = append(*r.details, msg)
	return msg.bp.count
}

// FUNCTION: 更新(明細なし)
func (r *OrderRecord) setNoDetail() int {
	msg := OrderMsg{
		OrderNo: r.record.OrderNo.Int,
		bp:      removedPiece("明細が存在しないため、登録しませんでした。"),
	}
	*r.details = append(*r.details, msg)
	return msg.bp.count
}

// STRUCT: コマンド
type OrdersCmd struct {
	details []OrderMsg
}

// FUNCTION: New
func NewOrdersCmd() *OrdersCmd {
	return &OrdersCmd{details: []OrderMsg{}}
}

// FUNCTION: テーブル名設定
func (cmd *OrdersCmd) getTableInfo() TableInfo {
	return TableInfo{
		schema:  "orders",
		tableJp: "受注",
		tableEn: orders.TableNames.Orders,
	}
}

// FUNCTION: 入力データ量
func (cmd *OrdersCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := clean.Orders().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理データ量(OrderView)
func (cmd *OrdersCmd) operationCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := clean.WOrders().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *OrdersCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	records, err := clean.WOrders(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &OrderRecord{record: *record, details: &cmd.details}}
	}
	return results
}

// FUNCTION: 結果データ量
func (cmd *OrdersCmd) resultCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := orders.Orders().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 詳細メッセージの出力
func (cmd *OrdersCmd) showDetails(ctx infra.AppCtx, tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | order_no | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, m := range cmd.details {
		msg += fmt.Sprintf("  | %d | %d | … | %s | %s | %s |\n",
			i+1,
			m.OrderNo,
			m.bp.status,
			ctx.Printer.Sprintf("%+d", m.bp.count),
			m.bp.msg,
		)
	}
	return msg
}
