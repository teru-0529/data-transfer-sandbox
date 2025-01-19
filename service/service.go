/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

import (
	"fmt"
	"os"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service/cleansing"
)

// TITLE: サービス共通

// FUNCTION: クレンジング
func Cleansing(file *os.File, conns infra.DbConnection) {

	var detailMessage string
	var num int

	file.WriteString("\n## Legacy Data Check and Cleansing\n\n")

	file.WriteString("  | # | TABLE | ENTRY | … | UNCHANGE | MODIFY | REMOVE | … | ACCEPT | RATE |\n")
	file.WriteString("  |--:|---|--:|---|--:|--:|--:|---|--:|--:|\n")

	// PROCESS: 1.operators
	num++
	cs := cleansing.NewOperators(conns)
	file.WriteString(resultRecord(num, cs.Result))
	detailMessage += cs.ShowDetails()

	// PROCESS: 詳細メッセージ
	if len(detailMessage) > 0 {
		file.WriteString("\n<details><summary>(open) modify and remove detail info</summary>\n")
		file.WriteString(detailMessage)
		file.WriteString("\n</details>\n")
	}
	file.WriteString("\n-----\n")
}

// FUNCTION: result件数
func resultRecord(num int, result cleansing.Result) string {
	return fmt.Sprintf("  | %d. | %s | %d | … | %d | %d | %d | … | %d | %3.1f%% |\n",
		num,
		result.TableName(),
		result.EntryCount,
		result.UnchangeCount,
		result.ModifyCount,
		result.RemoveCount,
		result.AcceptCount(),
		result.AcceptRate(),
	)
}
