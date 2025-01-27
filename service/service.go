/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

import (
	"context"
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service/cleansing"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/clean"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

// TITLE: サービス共通

// STRUCT: ダミーコンテキスト
var ctx context.Context = context.Background()

// STRUCT: printer(数値をカンマ区切りで出力するために利用)
var p = message.NewPrinter(language.Japanese)

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
	msg += cleansingResult(num, cs1.Result)
	detailMsg += cs1.ShowDetails()

	// PROCESS: 2.products
	num++
	cs2 := cleansing.NewProducts(conns, refData)
	msg += cleansingResult(num, cs2.Result)
	detailMsg += cs2.ShowDetails()

	// PROCESS: 3.orders
	num++
	cs3 := cleansing.NewOrders(conns, refData)
	msg += cleansingResult(num, cs3.Result)
	detailMsg += cs3.ShowDetails()

	// PROCESS: 3.orders
	num++
	cs4 := cleansing.NewOrderDetails(conns, refData)
	msg += cleansingResult(num, cs4.Result)
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

// FUNCTION: clensingResult件数
func cleansingResult(num int, result cleansing.Result) string {

	return fmt.Sprintf("  | %d. | %s | %s | %s | … | %s | %s | %s | … | %s | %3.1f%% |\n",
		num,
		result.TableName(),
		p.Sprintf("%d", result.EntryCount),
		p.Sprintf("%3.2fs", result.Elapsed()),
		p.Sprintf("%d", result.UnchangeCount),
		p.Sprintf("%d", result.ModifyCount),
		p.Sprintf("%d", result.RemoveCount),
		p.Sprintf("%d", result.AcceptCount()),
		result.AcceptRate(),
	)
}

// FUNCTION: cleanDBのテーブルを全てtruncate
func TruncateCleanDbAll(conns infra.DbConnection) {
	// INFO: テーブルリスト
	tables := []string{
		clean.TableNames.Operators,
		clean.TableNames.Products,
		clean.TableNames.Orders,
		clean.TableNames.OrderDetails,
	}
	for _, name := range tables {
		queries.Raw(fmt.Sprintf("truncate clean.%s CASCADE;", name)).ExecContext(ctx, conns.WorkDB)
	}
}
