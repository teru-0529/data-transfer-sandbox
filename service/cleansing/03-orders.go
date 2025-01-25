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
type OrdersClensing struct {
	conns   infra.DbConnection
	Result  Result
	Details []*OrderPiece
}

type OrderPiece struct {
	OrderNo  int
	status   Status
	approved bool
	message  string
}

// FUNCTION:
func NewOrders(conns infra.DbConnection) OrdersClensing {
	s := time.Now()

	// INFO: 固定値設定
	cs := OrdersClensing{conns: conns, Result: Result{
		TableNameJp: "受注",
		TableNameEn: "orders",
	}}

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
	log.Printf("cleansing completed [%s] … %3.2fs\n", cs.Result.TableNameEn, duration)
	return cs
}

// FUNCTION: 入力データ量
func (cs *OrdersClensing) setEntryCount() {
	// INFO: Legacyテーブル名
	num, err := legacy.Orders().Count(ctx, cs.conns.LegacyDB)
	if err != nil {
		log.Fatalln(err)
	}
	cs.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (cs *OrdersClensing) iterate() {
	bar := pb.Default.Start(cs.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < cs.Result.sectionCount(); section++ {
		// INFO: Legacyテーブル名
		records, err := legacy.Orders(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, cs.conns.LegacyDB)
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
func (cs *OrdersClensing) checkAndClensing(record *legacy.Order) *OrderPiece {
	// INFO: piece
	piece := OrderPiece{
		OrderNo: record.OrderNo,
		status:  NO_CHANGE,
	}

	// PROCESS: order_dateが日付のフォーマットに合致しない場合は、"20250101"にクレンジングする。
	orderDate := record.OrderDate
	var defOrderDate = "20250101"

	_, err := time.Parse(DATE_LAYOUT, orderDate)
	if err != nil {
		record.OrderDate = defOrderDate

		piece.status = MODIFY
		piece.approved = true
		if len(piece.message) != 0 {
			piece.message += "<BR>"
		}
		piece.message += fmt.Sprintf("● order_date(受注日付) が日付フォーマットではない。`%s` → `%s`(固定値) にクレンジング。", orderDate, defOrderDate)
	}

	// PROCESS: order_picが[担当者]に存在しない場合は、"N/A"にクレンジングする。
	orderPic := record.OrderPic
	var defOrderPic = "N/A"

	// PK検索ではないので、legacy.OperatorExists()は使えない。
	ok1, _ := legacy.Operators(legacy.OperatorWhere.OperatorName.EQ(orderPic)).Exists(ctx, cs.conns.LegacyDB)
	if !ok1 {
		record.OrderPic = defOrderPic

		piece.status = MODIFY
		piece.approved = true
		if len(piece.message) != 0 {
			piece.message += "<BR>"
		}
		piece.message += fmt.Sprintf("● order_pic(受注担当者名) が[担当者]に存在しない。`%s` → `%s`(固定値) にクレンジング。", orderPic, defOrderPic)
	}

	return &piece
}

// FUNCTION: レコード毎のクレンジング後データ登録
func (cs *OrdersClensing) saveData(record *legacy.Order, piece *OrderPiece) {
	// PROCESS: REMOVEDの場合はDBに登録しない
	if piece.status == REMOVE {
		cs.setResult(piece)
		return
	}

	orderDate, _ := time.Parse(DATE_LAYOUT, record.OrderDate)

	// PROCESS: データ登録
	// INFO: cleanテーブル
	rec := clean.Order{
		OrderNo:      record.OrderNo,
		OrderDate:    orderDate,
		OrderPic:     record.OrderPic,
		CustomerName: record.CustomerName,
		CreatedBy:    OPERATION_USER,
		UpdatedBy:    OPERATION_USER,
	}
	err := rec.Insert(ctx, cs.conns.WorkDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納、未承認扱い)
	if err != nil {
		piece.status = REMOVE
		piece.approved = false
		if len(piece.message) != 0 {
			piece.message += "<BR>"
		}
		piece.message += fmt.Sprintf("%v", err)
	}
	cs.setResult(piece)
}

// FUNCTION: クレンジング結果の登録
func (cs *OrdersClensing) setResult(piece *OrderPiece) {
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
func (cs *OrdersClensing) ShowDetails() string {
	if len(cs.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", cs.Result.TableName())

	msg += "  | # | order_no | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|--:|---|:-:|:-:|---|\n"
	for i, piece := range cs.Details {
		msg += fmt.Sprintf("  | %d | %d | … | %s | %s | %s |\n",
			i+1,
			piece.OrderNo,
			piece.status,
			approveStr(piece.approved),
			piece.message,
		)
	}

	return msg
}
