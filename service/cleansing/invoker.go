/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// TITLE: クレンジングインボーカー

// STRUCT: レコードインターフェース

type LegacyRecord interface {
	checkAndPersist(ctx infra.AppCtx, db *sql.DB, refData *RefData) Piece
}

// STRUCT: レコード(ラッパー)
type Record struct {
	rec LegacyRecord
}

// FUNCTION:
func (r *Record) save(ctx infra.AppCtx, db *sql.DB, refData *RefData) Piece {
	return r.rec.checkAndPersist(ctx, db, refData)
}

// STRUCT: コマンドインターフェース
type Command interface {
	getTableInfo() TableInfo
	entryCount(ctx infra.AppCtx, con *sql.DB) int
	fetchRecords(ctx infra.AppCtx, db *sql.DB, qmArray []qm.QueryMod) []Record
	showDetails(ctx infra.AppCtx, tableName string) string
	extInsert(ctx infra.AppCtx, con *sql.DB, refData *RefData)
}

// STRUCT: インボーカー
type Invoker struct {
	num     int
	ctx     infra.AppCtx
	conns   infra.DbConnection
	cmd     Command
	refData *RefData
}

// FUNCTION:
func NewInvoker(num int, ctx infra.AppCtx, conns infra.DbConnection, refData *RefData, cmd Command) *Invoker {
	return &Invoker{
		num:     num,
		ctx:     ctx,
		conns:   conns,
		cmd:     cmd,
		refData: refData,
	}
}

// FUNCTION: 実行
func (inv *Invoker) Execute() (string, string) {
	s := time.Now()

	// PROCESS: テーブル名称取得
	table := inv.cmd.getTableInfo()
	log.Printf("[%s] table cleansing ...", table.tableEn)

	// PROCESS: 入力データ量
	count := inv.cmd.entryCount(inv.ctx, inv.conns.LegacyDB)

	// PROCESS: 移行先のtruncate
	_, err := queries.Raw(table.truncateSql()).ExecContext(inv.ctx.Ctx, inv.conns.WorkDB)
	if err != nil {
		log.Fatalln(err)
	}

	// PROCESS: データ取得/登録
	result := inv.iterate(count)

	// PROCESS: 追加データ登録
	inv.cmd.extInsert(inv.ctx, inv.conns.WorkDB, inv.refData)

	// PROCESS: 後処理
	duration := time.Since(s).Seconds()
	log.Printf("cleansing completed … %3.2fs\n", duration)
	return inv.showRecord(table, result, duration), inv.cmd.showDetails(inv.ctx, table.Name())
}

// FUNCTION: データ取得/登録
func (inv *Invoker) iterate(count int) ResultCount {
	result := ResultCount{EntryCount: count}
	bar := pb.Default.Start(count)
	bar.SetMaxWidth(80)

	// PROCESS: SQLによる取得を分割する
	for lap := 0; lap < inv.ctx.LapNumber(count); lap++ {
		qmArray := []qm.QueryMod{qm.Limit(inv.ctx.Limit), qm.Offset(lap * inv.ctx.Limit)}

		records := inv.cmd.fetchRecords(inv.ctx, inv.conns.LegacyDB, qmArray)

		for _, record := range records {
			// PROCESS: レコード毎のデータ登録
			result.add(record.save(inv.ctx, inv.conns.WorkDB, inv.refData))
			bar.Increment()
		}
	}
	bar.Finish()
	return result
}

// FUNCTION: メッセージの出力
func (inv *Invoker) showRecord(t TableInfo, r ResultCount, duration float64) string {
	return fmt.Sprintf("  | %d. | %s | %s | %s | … | %s | %s | %s%s | … | %s | %3.1f%% |\n",
		inv.num,
		t.Name(),
		inv.ctx.Printer.Sprintf("%d", r.EntryCount),
		inv.ctx.Printer.Sprintf("%3.2fs", duration),
		inv.ctx.Printer.Sprintf("%d", r.UnchangeCount),
		inv.ctx.Printer.Sprintf("%d", r.ModifyCount),
		inv.ctx.Printer.Sprintf("%d", r.RemoveCount),
		r.dbCheckCountStr(inv.ctx),
		inv.ctx.Printer.Sprintf("%d", r.AcceptCount()),
		r.AcceptRate(),
	)
}

// STRUCT: テーブル情報
type TableInfo struct {
	schema  string
	tableJp string
	tableEn string
}

// FUNCTION: テーブル名
func (t TableInfo) Name() string {
	return fmt.Sprintf("%s(%s)", t.tableEn, t.tableJp)
}

// FUNCTION: truuncate文
func (t TableInfo) truncateSql() string {
	return fmt.Sprintf("TRUNCATE clean.%s CASCADE;", t.tableEn)
}

// STRUCT: 結果件数
type ResultCount struct {
	EntryCount    int
	UnchangeCount int
	ModifyCount   int
	RemoveCount   int
	DbCheckCount  int
}

// FUNCTION: クレンジング結果の登録
func (r *ResultCount) add(bp Piece) {
	switch bp.status {
	case NO_CHANGE:
		r.UnchangeCount++
	case MODIFY:
		r.ModifyCount++
	case REMOVE:
		r.RemoveCount++
	}
	if bp.approve == NOT_FINDED {
		r.DbCheckCount++
	}
}

// FUNCTION:
func (r ResultCount) AcceptCount() int {
	return r.UnchangeCount + r.ModifyCount
}

// FUNCTION:
func (r ResultCount) AcceptRate() float64 {
	if r.EntryCount == 0 {
		return 0.0
	} else {
		return float64(r.AcceptCount()) / float64(r.EntryCount) * 100
	}
}

// FUNCTION:
func (r ResultCount) dbCheckCountStr(ctx infra.AppCtx) string {
	if r.DbCheckCount > 0 {
		return ctx.Printer.Sprintf("<span style=\"color:red;\">(※%d)</span>", r.DbCheckCount)
	} else {
		return ""
	}
}
