/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// STRUCT: 詳細メッセージ
type OperatorMsg struct {
	OperatorId string
	bp         *Piece
}

func NewOperatorMsg(operatorId string) *OperatorMsg {
	return &OperatorMsg{
		OperatorId: operatorId,
		bp:         NewPiece(),
	}
}

// STRUCT: レコード
type OperatorRecord struct {
	record legacy.Operator
	msg    *OperatorMsg
	setTo  *[]OperatorMsg
}

// FUNCTION: 更新
func (r *OperatorRecord) checkAndPersist(ctx infra.AppCtx, db *sql.DB, refData *RefData) Piece {

	// PROCESS: check #1-01:
	r.checkOperatorName(refData.OperatorNameSet)

	// PROCESS:TODO: check #1-02:
	r.checkOperatorId()

	// PROCESS: REMOVE判定時は登録なし
	if !r.msg.bp.isRemove() {
		r.persiste(ctx, db)
	}

	// PROCESS: REMOVE/MODIFY判定時は詳細情報の出力あり
	if r.msg.bp.isWarn() {
		*r.setTo = append(*r.setTo, *r.msg)
	}

	// PROCESS:INFO: 正常登録時に[担当者名]登録
	if !r.msg.bp.isRemove() {
		refData.OperatorNameSet[r.record.OperatorName] = struct{}{}
	}

	return *r.msg.bp
}

// FUNCTION: データ登録
func (r *OperatorRecord) persiste(ctx infra.AppCtx, db *sql.DB) {
	// PROCESS: データ登録
	rec := clean.Operator{
		OperatorID:   r.record.OperatorID,
		OperatorName: r.record.OperatorName,
		CreatedBy:    ctx.OperationUser,
		UpdatedBy:    ctx.OperationUser,
	}
	err := rec.Insert(ctx.Ctx, db, boil.Infer())

	// PROCESS: 登録に失敗した場合は、削除(エラーログを格納)
	if err != nil {
		r.msg.bp.dbError(err)
	}
}

// FUNCTION: #1-01(REMOVE): 担当者のユニークチェック、すでに担当者名が存在する場合、移行対象から除外する。
func (r *OperatorRecord) checkOperatorName(OperatorNameSet map[string]struct{}) {
	const ID = "#1-01"
	_, exist := OperatorNameSet[r.record.OperatorName]
	if exist {
		r.msg.bp.removed().addMessage(
			fmt.Sprintf("operator_name(担当者名) がユニーク制約に違反しています`%s`。【除外】", r.record.OperatorName), ID)
	}
}

// FUNCTION: #1-02(MODIFY): 担当者IDが5桁ではない場合、末尾に「X」を埋めてクレンジングする。
func (r *OperatorRecord) checkOperatorId() {
	const ID = "#1-02"
	const LENGTH int = 5
	const CHAR string = "X"
	operatorId := r.record.OperatorID
	if len(operatorId) < LENGTH {
		r.record.OperatorID = operatorId + strings.Repeat(CHAR, LENGTH-len(operatorId))
		r.msg.bp.approveStay() //TODO: 承認確認中
		r.msg.bp.modified().addMessage(
			fmt.Sprintf("operator_id(担当者ID) の桁数が5桁未満(%d桁)です。<br>【クレンジング】末尾に`X`を追加", len(operatorId)), ID)
	}
}

// STRUCT: コマンド
type OperatorsCmd struct {
	details []OperatorMsg
}

// FUNCTION: New
func NewOperatorsCmd() *OperatorsCmd {
	return &OperatorsCmd{}
}

// FUNCTION: テーブル名設定
func (cmd *OperatorsCmd) getTableInfo() TableInfo {
	return TableInfo{
		tableJp: "担当者",
		tableEn: legacy.TableNames.Operators,
	}
}

// FUNCTION: 入力データ量
func (cmd *OperatorsCmd) entryCount(ctx infra.AppCtx, con *sql.DB) int {
	num, err := legacy.Operators().Count(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}
	return int(num)
}

// FUNCTION: 処理対象レコードのフェッチ
func (cmd *OperatorsCmd) fetchRecords(ctx infra.AppCtx, con *sql.DB, qmArray []qm.QueryMod) []Record {
	records, err := legacy.Operators(qmArray...).All(ctx.Ctx, con)
	if err != nil {
		log.Fatalln(err)
	}

	results := make([]Record, len(records))
	for i, record := range records {
		results[i] = Record{rec: &OperatorRecord{
			record: *record,
			msg:    NewOperatorMsg(record.OperatorID),
			setTo:  &cmd.details,
		}}
	}
	return results
}

// FUNCTION: 追加データ登録
func (r *OperatorsCmd) extInsert(ctx infra.AppCtx, db *sql.DB, refData *RefData) {
	// INFO: ダミー担当者
	rec := clean.Operator{
		OperatorID:   "Z9999",
		OperatorName: "N/A",
		CreatedBy:    ctx.OperationUser,
		UpdatedBy:    ctx.OperationUser,
	}
	rec.Insert(ctx.Ctx, db, boil.Infer())
	refData.OperatorNameSet["N/A"] = struct{}{}
}

// FUNCTION: 詳細メッセージの出力
func (cmd *OperatorsCmd) showDetails(ctx infra.AppCtx, tableName string) string {
	if len(cmd.details) == 0 {
		return ""
	}

	var msg string
	msg += fmt.Sprintf("\n### %s\n\n", tableName)
	msg += "  | # | operator_id | … | RESULT | APPROVED | MESSAGE |\n"
	msg += "  |--:|---|---|:-:|:-:|---|\n"
	for i, piece := range cmd.details {
		msg += fmt.Sprintf("  | %d | %s | … | %s | %s | %s |\n",
			i+1,
			piece.OperatorId,
			piece.bp.status,
			piece.bp.approve,
			piece.bp.msg,
		)
	}

	return msg
}
