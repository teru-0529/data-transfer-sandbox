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
type OrderPiece struct {
	OrderNo int
	bp      *Piece
}

// STRUCT:
type OrderTransfer struct {
	conns        infra.DbConnection
	Result       Result
	Details      []*OrderPiece
	OrderNoCount map[int]int //受注の分割数

}

// FUNCTION:
func NewOrders(conns infra.DbConnection) OrderTransfer {
	s := time.Now()

	// INFO: 固定値設定
	ts := OrderTransfer{
		conns:        conns,
		Result:       Result{Schema: "orders", TableNameJp: "受注", TableNameEn: "orders"},
		OrderNoCount: map[int]int{},
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

	duration := time.Since(s).Seconds()
	ts.Result.duration = duration
	log.Printf("cleansing completed … %3.2fs\n", duration)
	return ts
}

// FUNCTION: 入力データ量
func (ts *OrderTransfer) setEntryCount() {
	// INFO: Workテーブル名
	num, err := clean.Orders().Count(ctx, ts.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}
	ts.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (ts *OrderTransfer) iterate() {

	// INFO: 実態の登録件数取得(EXT)
	num, err := clean.WOrders().Count(ctx, ts.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}

	bar := pb.Default.Start(int(num))
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < ts.Result.sectionCount(); section++ {
		// INFO: Workテーブル名
		records, err := clean.WOrders(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, ts.conns.WorkDB)
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

	// INFO: 移行元の受注の削除/複製チェック
	// INFO: 1000件単位でのSQL実行に分割する
	for section := 0; section < ts.Result.sectionCount(); section++ {
		// INFO: Workテーブル名
		records, err := clean.Orders(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, ts.conns.WorkDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎の削除/複製チェック
			ts.check(record)
		}
	}
}

// FUNCTION: レコード毎のデータ登録
func (ts *OrderTransfer) saveData(record *clean.WOrder) {

	var status string
	if record.IsRemaining.Bool {
		status = orders.OrderStatusWORK_IN_PROGRESS
	} else if record.IsShipped.Bool {
		status = orders.OrderStatusCOMPLETED
	} else {
		status = orders.OrderStatusCANCELED
	}

	// PROCESS: データ登録
	// INFO: productテーブル
	rec := orders.Order{
		OrderNo:             record.WOrderNo.String,
		OrderDate:           record.OrderDate.Time,
		OrderPic:            record.OperatorID.String,
		CustomerName:        record.CustomerName.String,
		TotalOrderPrice:     int(record.WTotalOrderPrice.Int64),
		RemainingOrderPrice: int(record.WRemainingOrderPrice.Int64),
		OrderStatus:         status,
		CreatedBy:           OPERATION_USER,
		UpdatedBy:           OPERATION_USER,
	}
	err := rec.Insert(ctx, ts.conns.ProductDB, boil.Infer())

	orgOrderNo := record.OrderNo.Int
	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		// INFO: piece
		piece := OrderPiece{
			OrderNo: orgOrderNo,
			bp:      removedPiece(fmt.Sprintf("%v", err)),
		}
		ts.Result.setResult(piece.bp)
		ts.Details = append(ts.Details, &piece)
	}

	// INFO: 元の受注番号からいくつに分かれたかのカウント(EXT)
	_, exist := ts.OrderNoCount[orgOrderNo]
	if !exist {
		ts.OrderNoCount[orgOrderNo] = 1
	} else {
		ts.OrderNoCount[orgOrderNo]++
	}
}

// FUNCTION: 元レコードの状態確認 INFO:(EXT)
func (ts *OrderTransfer) check(record *clean.Order) {

	orderNo := record.OrderNo
	no, exist := ts.OrderNoCount[orderNo]

	// PROCESS: 登録していない場合
	if !exist {
		// INFO: piece
		piece := OrderPiece{
			OrderNo: orderNo,
			bp:      removedPiece("明細が存在しないため、登録しませんでした。"),
		}
		ts.Result.setResult(piece.bp)
		ts.Details = append(ts.Details, &piece)
		return
	}

	// PROCESS: 登録した件数が2以上の場合
	if no > 1 {
		// INFO: piece
		piece := OrderPiece{
			OrderNo: orderNo,
			bp: modifiedPiece(
				fmt.Sprintf("明細に`販売単価`もしくは`商品原価`が一致しない同一の商品が存在するため、受注を [%d] 件に分割しました。", no), no-1),
		}
		ts.Result.setResult(piece.bp)
		ts.Details = append(ts.Details, &piece)
		return
	}
}

// FUNCTION: 詳細情報の出力
func (ts *OrderTransfer) ShowDetails() string {
	if len(ts.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", ts.Result.TableName())

	msg += "  | # | order_no | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range ts.Details {
		msg += fmt.Sprintf("  | %d | %d | … | %s | %+d | %s |\n",
			i+1,
			piece.OrderNo,
			piece.bp.status,
			piece.bp.count,
			piece.bp.msg,
		)
	}

	return msg
}
