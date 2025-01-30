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
type OperatorPiece struct {
	OperatorId string
	bp         *Piece
}

// STRUCT:
type OperatorTransfer struct {
	conns   infra.DbConnection
	Result  Result
	Details []*OperatorPiece
}

// FUNCTION:
func NewOperators(conns infra.DbConnection) OperatorTransfer {
	s := time.Now()

	// INFO: 固定値設定
	ts := OperatorTransfer{
		conns:  conns,
		Result: Result{Schema: "orders", TableNameJp: "担当者", TableNameEn: "operators"},
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
func (ts *OperatorTransfer) setEntryCount() {
	// INFO: Workテーブル名
	num, err := clean.Operators().Count(ctx, ts.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}
	ts.Result.EntryCount = int(num)
}

// FUNCTION: 結果データ量
func (ts *OperatorTransfer) setResultCount() {
	// INFO: Productionテーブル名
	num, err := orders.Operators().Count(ctx, ts.conns.ProductDB)
	if err != nil {
		log.Fatalln(err)
	}
	ts.Result.ResultCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (ts *OperatorTransfer) iterate() {
	bar := pb.Default.Start(ts.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < ts.Result.sectionCount(); section++ {
		// INFO: Workテーブル名
		records, err := clean.Operators(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, ts.conns.WorkDB)
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
func (ts *OperatorTransfer) saveData(record *clean.Operator) {

	// PROCESS: データ登録
	// INFO: productionテーブル
	rec := orders.Operator{
		OperatorID:   record.OperatorID,
		OperatorName: record.OperatorName,
		CreatedBy:    OPERATION_USER,
		UpdatedBy:    OPERATION_USER,
	}
	err := rec.Insert(ctx, ts.conns.ProductDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		// INFO: piece
		piece := OperatorPiece{
			OperatorId: record.OperatorID,
			bp:         removedPiece(fmt.Sprintf("%v", err)),
		}
		ts.Result.setResult(piece.bp)
		ts.Details = append(ts.Details, &piece)
	}
}

// FUNCTION: 詳細情報の出力
func (ts *OperatorTransfer) ShowDetails() string {
	if len(ts.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", ts.Result.TableName())

	msg += "  | # | operator_id | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range ts.Details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %+d | %s |\n",
			i+1,
			piece.OperatorId,
			piece.bp.status,
			piece.bp.count,
			piece.bp.msg,
		)
	}

	return msg
}
