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
type OperatorClensing struct {
	conns   infra.DbConnection
	Result  Result
	Details []*OperatorPiece
}

// FUNCTION:
func NewOperators(conns infra.DbConnection) OperatorClensing {
	s := time.Now()

	// INFO: 固定値設定
	cs := OperatorClensing{
		conns:  conns,
		Result: Result{Schema: "orders", TableNameJp: "担当者", TableNameEn: "operators"},
	}
	log.Printf("[%s] table transfer ...", cs.Result.TableNameEn)

	// PROCESS: 入力データ量
	cs.setEntryCount()

	// PROCESS: 移行先のtruncate
	_, err := queries.Raw(cs.Result.truncateSql()).ExecContext(ctx, cs.conns.ProductDB)
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
func (cs *OperatorClensing) setEntryCount() {
	// INFO: Workテーブル名
	num, err := clean.Operators().Count(ctx, cs.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}
	cs.Result.EntryCount = int(num)
}

// FUNCTION: データ繰り返し取得(1000件単位で分割)
func (cs *OperatorClensing) iterate() {
	bar := pb.Default.Start(cs.Result.EntryCount)
	bar.SetMaxWidth(80)

	// PROCESS: 1000件単位でのSQL実行に分割する
	for section := 0; section < cs.Result.sectionCount(); section++ {
		// INFO: Workテーブル名
		records, err := clean.Operators(qm.Limit(LIMIT), qm.Offset(section*LIMIT)).All(ctx, cs.conns.WorkDB)
		if err != nil {
			log.Fatalln(err)
		}

		for _, record := range records {
			// PROCESS: レコード毎のデータ登録
			cs.saveData(record)

			bar.Increment()
		}
	}
	bar.Finish()
}

// FUNCTION: レコード毎のデータ登録
func (cs *OperatorClensing) saveData(record *clean.Operator) {

	// PROCESS: データ登録
	// INFO: cleanテーブル
	rec := orders.Operator{
		OperatorID:   record.OperatorID,
		OperatorName: record.OperatorName,
		CreatedBy:    OPERATION_USER,
		UpdatedBy:    OPERATION_USER,
	}
	err := rec.Insert(ctx, cs.conns.ProductDB, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		piece := OperatorPiece{
			OperatorId: record.OperatorID,
			bp:         removedPiece(fmt.Sprintf("%v", err)),
		}
		cs.Result.setResult(piece.bp)
		cs.Details = append(cs.Details, &piece)
	}
}

// FUNCTION: 詳細情報の出力
func (cs *OperatorClensing) ShowDetails() string {
	if len(cs.Details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", cs.Result.TableName())

	msg += "  | # | operator_id | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range cs.Details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %d | %s |\n",
			i+1,
			piece.OperatorId,
			piece.bp.status,
			piece.bp.count,
			piece.bp.msg,
		)
	}

	return msg
}
