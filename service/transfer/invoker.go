/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package transfer

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

// TITLE: データ変換インボーカー

// STRUCT: レコードインターフェース
type CleanRecord interface {
	applyChanges(ctx infra.AppCtx, db *sql.DB) int
}

// STRUCT: レコード(ラッパー)
type Record struct {
	rec CleanRecord
}

// FUNCTION:
func (r *Record) save(ctx infra.AppCtx, db *sql.DB) int {
	return r.rec.applyChanges(ctx, db)
}

// STRUCT: コマンドインターフェース
type Command interface {
	getTableInfo() TableInfo
	entryCount(ctx infra.AppCtx, db *sql.DB) int
	operationCount(ctx infra.AppCtx, db *sql.DB) int
	fetchRecords(ctx infra.AppCtx, db *sql.DB, qmArray []qm.QueryMod) []Record
	resultCount(ctx infra.AppCtx, db *sql.DB) int
	showDetails(ctx infra.AppCtx, tableName string) string
}

// STRUCT: インボーカー
type Invoker struct {
	num   int
	ctx   infra.AppCtx
	conns infra.DbConnection
	cmd   Command
}

// FUNCTION:
func NewInvoker(num int, ctx infra.AppCtx, conns infra.DbConnection, cmd Command) *Invoker {
	return &Invoker{
		num:   num,
		ctx:   ctx,
		conns: conns,
		cmd:   cmd,
	}
}

// FUNCTION: 実行
func (inv *Invoker) Execute() (string, string) {
	s := time.Now()
	result := ResultCount{}

	// PROCESS: テーブル名称取得
	table := inv.cmd.getTableInfo()
	log.Printf("[%s] table transfer ...", table.tableEn)

	// PROCESS: 入力データ量
	result.entryCount = inv.cmd.entryCount(inv.ctx, inv.conns.WorkDB)

	// PROCESS: 移行先のtruncate
	_, err := queries.Raw(table.truncateSql()).ExecContext(inv.ctx.Ctx, inv.conns.ProductDB)
	if err != nil {
		log.Fatalln(err)
	}

	// PROCESS: データ取得/登録
	result.changeCount = inv.iterate(inv.cmd.operationCount(inv.ctx, inv.conns.WorkDB))

	// PROCESS: 結果データ量
	result.resultCount = inv.cmd.resultCount(inv.ctx, inv.conns.ProductDB)

	// PROCESS: 後処理
	duration := time.Since(s).Seconds()
	log.Printf("transfer completed … %3.2fs\n", duration)
	return inv.showRecord(table, result, duration), inv.cmd.showDetails(inv.ctx, table.Name())
}

// FUNCTION: データ取得/登録
func (inv *Invoker) iterate(count int) int {
	changeCount := 0
	bar := pb.Default.Start(count)
	bar.SetMaxWidth(80)

	// PROCESS: SQLによる取得を分割する
	for lap := 0; lap < inv.ctx.LapNumber(count); lap++ {
		qmArray := []qm.QueryMod{qm.Limit(inv.ctx.Limit), qm.Offset(lap * inv.ctx.Limit)}

		records := inv.cmd.fetchRecords(inv.ctx, inv.conns.WorkDB, qmArray)

		for _, record := range records {
			// PROCESS: レコード毎のデータ登録
			changeCount += record.save(inv.ctx, inv.conns.ProductDB)
			bar.Increment()
		}
	}
	bar.Finish()
	return changeCount
}

// FUNCTION: メッセージの出力
func (inv *Invoker) showRecord(t TableInfo, r ResultCount, duration float64) string {
	return fmt.Sprintf("  | %d. | %s | %s | %s | %s | … | %s | … | %s | %s |\n",
		inv.num,
		t.schema,
		t.Name(),
		inv.ctx.Printer.Sprintf("%d", r.entryCount),
		inv.ctx.Printer.Sprintf("%3.2fs", duration),
		inv.ctx.Printer.Sprintf("%+d", r.changeCount),
		inv.ctx.Printer.Sprintf("%d", r.resultCount),
		r.checkRecord(),
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
	return fmt.Sprintf("TRUNCATE %s.%s CASCADE;", t.schema, t.tableEn)
}

// STRUCT: 結果件数
type ResultCount struct {
	entryCount  int
	changeCount int
	resultCount int
}

// FUNCTION: 件数推移チェック
func (r ResultCount) checkRecord() string {
	if r.entryCount+r.changeCount == r.resultCount {
		return ""
	} else {
		return "❎"
	}
}
