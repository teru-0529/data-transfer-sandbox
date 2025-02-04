/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

import (
	"context"
	"fmt"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service/cleansing"
	"github.com/teru-0529/data-transfer-sandbox/service/transfer"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

// TITLE: サービス共通

// STRUCT: ダミーコンテキスト
var ctx context.Context = context.Background()

// FUNCTION: クレンジング
func Cleansing(conns infra.DbConnection) string {

	var detailMsg string
	var num int
	refData := cleansing.NewRefData()

	msg := "\n## Legacy Data Check and Cleansing\n\n"
	msg += "  | # | TABLE | ENTRY | ELAPSED | … | UNCHANGE | MODIFY | REMOVE | … | ACCEPT | RATE |\n"
	msg += "  |--:|---|--:|--:|---|--:|--:|--:|---|--:|--:|\n"

	// PROCESS: 1.operators
	num++
	cs1 := cleansing.NewOperators(conns, refData)
	msg += cs1.Result.ShowRecord(num)
	detailMsg += cs1.ShowDetails()

	// PROCESS: 2.products
	num++
	cs2 := cleansing.NewProducts(conns, refData)
	msg += cs2.Result.ShowRecord(num)
	detailMsg += cs2.ShowDetails()

	// PROCESS: 3.orders
	num++
	cs3 := cleansing.NewOrders(conns, refData)
	msg += cs3.Result.ShowRecord(num)
	detailMsg += cs3.ShowDetails()

	// PROCESS: 4.order_details
	num++
	cs4 := cleansing.NewOrderDetails(conns, refData)
	msg += cs4.Result.ShowRecord(num)
	detailMsg += cs4.ShowDetails()

	// PROCESS: 詳細メッセージ
	if len(detailMsg) > 0 {
		msg += "\n<details><summary>(open) modify and remove detail info</summary>\n"
		msg += detailMsg
		msg += "\n</details>\n"
	}
	msg += "\n-----\n"
	return msg
}

// STRUCT: テーブル名
type TableNameOnly struct {
	Tablename string `boil:"tablename"`
}

// FUNCTION: cleanDBのテーブルを全てtruncate
func TruncateCleanDbAll(conns infra.DbConnection) {
	boil.DebugMode = true

	var tables []TableNameOnly
	queries.Raw("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'clean'").Bind(ctx, conns.WorkDB, &tables)
	for _, table := range tables {
		queries.Raw(fmt.Sprintf("truncate clean.%s CASCADE;", table.Tablename)).ExecContext(ctx, conns.WorkDB)
	}

	boil.DebugMode = false
}

// FUNCTION: 移行
func Transfer(conns infra.DbConnection) string {

	var detailMsg string
	var num int

	msg := "\n## Data Transfer to Production DB\n\n"
	msg += "  | # | SCHEMA | TABLE | ENTRY | ELAPSED | … | CHANGE | … | ACCEPT | CHECK |\n"
	msg += "  |--:|---|---|--:|--:|---|--:|---|--:|:--:|\n"

	// PROCESS: 1.operators
	num++
	cs1 := transfer.NewOperators(conns)
	msg += cs1.Result.ShowRecord(num)
	detailMsg += cs1.ShowDetails()

	// PROCESS: 2.products
	num++
	cs2 := transfer.NewProducts(conns)
	msg += cs2.Result.ShowRecord(num)
	detailMsg += cs2.ShowDetails()

	// PROCESS: 3.orders
	num++
	cs3 := transfer.NewOrders(conns)
	msg += cs3.Result.ShowRecord(num)
	detailMsg += cs3.ShowDetails()

	// PROCESS: 4.order_details
	num++
	cs4 := transfer.NewOrderDetails(conns)
	msg += cs4.Result.ShowRecord(num)
	detailMsg += cs4.ShowDetails()

	// PROCESS: 詳細メッセージ
	if len(detailMsg) > 0 {
		msg += "\n<details><summary>(open) modify and remove detail info</summary>\n"
		msg += detailMsg
		msg += "\n</details>\n"
	}
	msg += "\n-----\n"
	return msg
}
