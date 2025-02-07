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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: 詳細メッセージ
type OrderDetailMsg struct {
	OrderNo        int
	OrderDetailNos string
	bp             *Piece
}

// STRUCT: レコード
type OrderDetailRecord struct {
	record  clean.WOrderDetail
	details *[]OrderDetailMsg
}

// FUNCTION: 更新
func (r *OrderDetailRecord) applyChanges(ctx infra.AppCtx, db *sql.DB, user null.String) int {

	// PROCESS: 明細を集約した場合
	if !r.record.Register.Bool {
		return r.setAggregated(int(r.record.DetailCount.Int64))
	}

	// PROCESS: ステータスの判定
	var status string
	if r.record.IsRemaining.Bool {
		status = orders.OrderStatusWORK_IN_PROGRESS
	} else if r.record.IsShipped.Bool {
		status = orders.OrderStatusCOMPLETED
	} else {
		status = orders.OrderStatusCANCELED
	}

	// PROCESS: データ登録
	rec := orders.OrderDetail{
		OrderNo:           r.record.WOrderNo.String,
		ProductID:         r.record.WProductID.String,
		ReceivingQuantity: int(r.record.ReceivingQuantity.Int64),
		ShippingQuantity:  int(r.record.WShippingQuantity.Int64),
		CancelQuantity:    int(r.record.WCancelQuantity.Int64),
		RemainingQuantity: int(r.record.WRemainingQuantity.Int64),
		CostPrice:         r.record.CostPrice.Int,
		SelllingPrice:     r.record.SellingPrice.Int,
		OrderStatus:       status,
		CreatedBy:         user,
		UpdatedBy:         user,
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		return r.setError(err)
	}

	return 0
}

// FUNCTION: 更新(エラー)
func (r *OrderDetailRecord) setError(err error) int {
	msg := OrderDetailMsg{
		OrderNo:        r.record.OrderNo.Int,
		OrderDetailNos: r.record.AggregatedDetails.String,
		bp:             errorPiece(err),
	}
	*r.details = append(*r.details, msg)
	return msg.bp.count
}

// FUNCTION: 更新(明細統合)
func (r *OrderDetailRecord) setAggregated(no int) int {
	msg := OrderDetailMsg{
		OrderNo:        r.record.OrderNo.Int,
		OrderDetailNos: r.record.AggregatedDetails.String,
		bp:             modifiedPiece(fmt.Sprintf("受注明細 [%d] 件を集約ししました。", no), int(1-no)),
	}
	*r.details = append(*r.details, msg)
	return msg.bp.count
}

// STRUCT: コマンド
type OrderDetailsCmd struct {
	details []OrderDetailMsg
}

// FUNCTION: New
func NewOrderDetailsCmd() *OrderDetailsCmd {
	return &OrderDetailsCmd{details: []OrderDetailMsg{}}
}

// FUNCTION: テーブル名設定
func (cmd *OrderDetailsCmd) getTableInfo() TableInfo {
	return TableInfo{
		schema:  "orders",
		tableJp: "受注明細",
		tableEn: orders.TableNames.OrderDetails,
	}
}

// FUNCTION: 入力データ量
func (cmd *OrderDetailsCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := clean.OrderDetails().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理データ量(OrderDetailView)
func (cmd *OrderDetailsCmd) operationCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := clean.WOrderDetails().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *OrderDetailsCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	records, err := clean.WOrderDetails(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &OrderDetailRecord{record: *record, details: &cmd.details}}
	}
	return results
}

// FUNCTION: 結果データ量
func (cmd *OrderDetailsCmd) resultCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := orders.OrderDetails().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 詳細メッセージの出力
func (cmd *OrderDetailsCmd) showDetails(tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | order_no | order_detail_nos | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|---|:-:|:-:|---|\n"
	for i, m := range cmd.details {
		msg += fmt.Sprintf("  | %d | %d | %s | … | %s | %+d | %s |\n",
			i+1,
			m.OrderNo,
			m.OrderDetailNos,
			m.bp.status,
			m.bp.count,
			m.bp.msg,
		)
	}
	return msg
}
