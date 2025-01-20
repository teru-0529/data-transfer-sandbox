/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

import (
	"fmt"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service/cleansing"
)

// TITLE: サービス共通

// STRUCT: printer(数値をカンマ区切りで出力するために利用)
var p = message.NewPrinter(language.Japanese)

// FUNCTION: クレンジング
func Cleansing(file *os.File, conns infra.DbConnection) {

	var detailMessage string
	var num int

	file.WriteString("\n## Legacy Data Check and Cleansing\n\n")

	file.WriteString("  | # | TABLE | ENTRY | ELAPSED | … | UNCHANGE | MODIFY | REMOVE | … | ACCEPT | RATE |\n")
	file.WriteString("  |--:|---|--:|--:|---|--:|--:|--:|---|--:|--:|\n")

	// PROCESS: 1.operators
	num++
	cs1 := cleansing.NewOperators(conns)
	file.WriteString(cleansingResult(num, cs1.Result))
	detailMessage += cs1.ShowDetails()

	// PROCESS: 2.products
	num++
	cs2 := cleansing.NewProducts(conns)
	file.WriteString(cleansingResult(num, cs2.Result))
	detailMessage += cs2.ShowDetails()

	// PROCESS: 3.orders
	num++
	cs3 := cleansing.NewOrders(conns)
	file.WriteString(cleansingResult(num, cs3.Result))
	detailMessage += cs3.ShowDetails()

	// PROCESS: 3.orders
	num++
	cs4 := cleansing.NewOrderDetails(conns)
	file.WriteString(cleansingResult(num, cs4.Result))
	detailMessage += cs4.ShowDetails()

	// PROCESS: 詳細メッセージ
	if len(detailMessage) > 0 {
		file.WriteString("\n<details><summary>(open) modify and remove detail info</summary>\n")
		file.WriteString(detailMessage)
		file.WriteString("\n</details>\n")
	}
	file.WriteString("\n-----\n")
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
