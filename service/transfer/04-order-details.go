/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package transfer

import (
	"fmt"
	"log"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/product/orders"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT:
type OrderDetailPiece struct {
	OrderNo       int
	OrderDetailNo string
	bp            *Piece
}

// STRUCT:
type OrderDetailsTransfer struct {
	conns   infra.DbConnection
	Result  Result
	Details []*OrderDetailPiece
}

// FUNCTION:
func NewOrderDetails(conns infra.DbConnection) OrderDetailsTransfer {
	s := time.Now()

	// INFO: 固定値設定
	ts := OrderDetailsTransfer{
		conns:  conns,
		Result: Result{Schema: "orders", TableNameJp: "受注明細", TableNameEn: "order_details"},
	}
	log.Printf("[%s] table transfer ...", ts.Result.TableNameEn)

	// PROCESS: 入力データ量
	ts.setEntryCount()

	// PROCESS: 移行先のtruncate
	_, err := queries.Raw(ts.Result.truncateSql()).ExecContext(ctx, ts.conns.ProductDB)
	if err != nil {
		log.Fatalln(err)
	}

	// PROCESS: 1行ごと処理
	ts.iterate()

	// PROCESS: 結果データ量
	ts.setResultCount()

	duration := time.Since(s).Seconds()
	ts.Result.duration = duration
	log.Printf("cleansing completed … %3.2fs\n", duration)
	return ts
}

// FUNCTION: 入力データ量
func (ts *OrderDetailsTransfer) setEntryCount() {
	// INFO: Workテーブル名
	num, err := clean.OrderDetails().Count(ctx, ts.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}
	ts.Result.EntryCount = int(num)
}

// FUNCTION: 結果データ量
func (ts *OrderDetailsTransfer) setResultCount() {
	// INFO: Productionテーブル名
	num, err := orders.OrderDetails().Count(ctx, ts.conns.ProductDB)
	if err != nil {
		log.Fatalln(err)
	}
	ts.Result.ResultCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (ts *OrderDetailsTransfer) iterate() {

	// INFO: 実体の登録件数取得(EXT)
	num, err := clean.WOrderDetails().Count(ctx, ts.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}

	bar := pb.Default.Start(int(num))
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < sectionCount(int(num)); section++ {
		// INFO: Workテーブル名
		records, err := clean.WOrderDetails(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, ts.conns.WorkDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎のデータ登録
			ts.saveData(record)

			bar.Increment()
		}
	}
	bar.Finish()
}

// FUNCTION: レコード毎のデータ登録
func (ts *OrderDetailsTransfer) saveData(record *clean.WOrderDetail) {

	if record.Register.Bool {
		// PROCESS: 登録条件を満たしたとき

		var status string
		if record.IsRemaining.Bool {
			status = orders.OrderStatusWORK_IN_PROGRESS
		} else if record.IsShipped.Bool {
			status = orders.OrderStatusCOMPLETED
		} else {
			status = orders.OrderStatusCANCELED
		}

		// PROCESS: データ登録
		// INFO: productionテーブル
		rec := orders.OrderDetail{
			OrderNo:           record.WOrderNo.String,
			ProductID:         record.WProductID.String,
			ReceivingQuantity: int(record.ReceivingQuantity.Int64),
			ShippingQuantity:  int(record.WShippingQuantity.Int64),
			CancelQuantity:    int(record.WCancelQuantity.Int64),
			RemainingQuantity: int(record.WRemainingQuantity.Int64),
			CostPrice:         record.CostPrice.Int,
			SelllingPrice:     record.SellingPrice.Int,
			OrderStatus:       status,
			CreatedBy:         OPERATION_USER,
			UpdatedBy:         OPERATION_USER,
		}
		err := rec.Insert(ctx, ts.conns.ProductDB, boil.Infer())

		// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
		if err != nil {
			// INFO: piece
			piece := OrderDetailPiece{
				OrderNo:       record.OrderNo.Int,
				OrderDetailNo: record.AggregatedDetails.String,
				bp:            removedPiece(fmt.Sprintf("%v", err)),
			}
			ts.Result.setResult(piece.bp)
			ts.Details = append(ts.Details, &piece)
		}

	} else {
		// PROCESS: 登録条件を満たさないとき
		no := record.DetailCount.Int64
		// INFO: piece
		piece := OrderDetailPiece{
			OrderNo:       record.OrderNo.Int,
			OrderDetailNo: record.AggregatedDetails.String,
			bp:            modifiedPiece(fmt.Sprintf("受注明細 [%d] 件を集約ししました。", no), int(1-no)),
		}
		ts.Result.setResult(piece.bp)
		ts.Details = append(ts.Details, &piece)

	}
}

// FUNCTION: 詳細情報の出力
func (ts *OrderDetailsTransfer) ShowDetails() string {
	if len(ts.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", ts.Result.TableName())

	msg += "  | # | order_no | order_detail_no | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|---|:-:|:-:|---|\n"
	for i, piece := range ts.Details {
		msg += fmt.Sprintf("  | %d | %d | %s | … | %s | %+d | %s |\n",
			i+1,
			piece.OrderNo,
			piece.OrderDetailNo,
			piece.bp.status,
			piece.bp.count,
			piece.bp.msg,
		)
	}

	return msg
}
