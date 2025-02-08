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
type OperatorMsg struct {
	OperatorId string
	bp         *Piece
}

// STRUCT: レコード
type OperatorRecord struct {
	record  clean.Operator
	details *[]OperatorMsg
}

// FUNCTION: 更新
func (r *OperatorRecord) persist(ctx infra.AppCtx, db *sql.DB) int {
	// PROCESS: データ登録
	rec := orders.Operator{
		OperatorID:   r.record.OperatorID,
		OperatorName: r.record.OperatorName,
		CreatedBy:    ctx.OperationUser,
		UpdatedBy:    ctx.OperationUser,
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		return r.setError(err)
	}
	return 0
}

// FUNCTION: 更新(エラー)
func (r *OperatorRecord) setError(err error) int {
	msg := OperatorMsg{
		OperatorId: r.record.OperatorID,
		bp:         errorPiece(err),
	}
	*r.details = append(*r.details, msg)
	return msg.bp.count
}

// STRUCT: コマンド
type OperatorsCmd struct {
	details []OperatorMsg
	entry   int
}

// FUNCTION: New
func NewOperatorsCmd() *OperatorsCmd {
	return &OperatorsCmd{details: []OperatorMsg{}}
}

// FUNCTION: テーブル名設定
func (cmd *OperatorsCmd) getTableInfo() TableInfo {
	return TableInfo{
		schema:  "orders",
		tableJp: "担当者",
		tableEn: orders.TableNames.Operators,
	}
}

// FUNCTION: 入力データ量
func (cmd *OperatorsCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := clean.Operators().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	cmd.entry = int(num) //INFO: 処理データ量=入力データ量
	return int(num)
}

// FUNCTION: 処理データ量(通常は、処理データ量=入力データ量)
func (cmd *OperatorsCmd) operationCount(ctx infra.AppCtx, con *sql.DB) int {
	return cmd.entry
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *OperatorsCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	records, err := clean.Operators(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &OperatorRecord{record: *record, details: &cmd.details}}
	}
	return results
}

// FUNCTION: 結果データ量
func (cmd *OperatorsCmd) resultCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := orders.Operators().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 詳細メッセージの出力
func (cmd *OperatorsCmd) showDetails(ctx infra.AppCtx, tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | operator_id | … | RESULT | CHANGE | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, m := range cmd.details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %s | %s |\n",
			i+1,
			m.OperatorId,
			m.bp.status,
			ctx.Printer.Sprintf("%+d", m.bp.count),
			m.bp.msg,
		)
	}

	return msg
}
