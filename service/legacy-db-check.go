/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

// TITLE: 移行元DBの情報取得

import (
	"context"
	"fmt"
	"os"

	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/spec/source/legacy"
)

// STRUCT: ダミーコンテキスト
var ctx context.Context = context.Background()

// FUNCTION: 移行元情報の書き込み
func LegacyInfo(file *os.File, conns infra.DbConnection) {
	var num int64

	file.WriteString("\n<details><summary>(open) input Legacy database info</summary>\n\n")
	file.WriteString("## Legacy DB Data Count\n\n")

	file.WriteString("  | # | TABLE | COUNT |\n")
	file.WriteString("  |--:|---|--:|\n")

	// PROCESS: ORDERS
	num, _ = legacy.Orders().Count(ctx, conns.LegacyDB)
	writeCount(file, 1, "受注", "orders", num)

	// PROCESS: ORDER_DETAILS
	num, _ = legacy.OrderDetails().Count(ctx, conns.LegacyDB)
	writeCount(file, 2, "受注明細", "order_details", num)

	// PROCESS: PRODUCTS
	num, _ = legacy.Products().Count(ctx, conns.LegacyDB)
	writeCount(file, 3, "商品", "products", num)

	// PROCESS: OPERATORS
	num, _ = legacy.Operators().Count(ctx, conns.LegacyDB)
	writeCount(file, 4, "担当者", "operators", num)

	file.WriteString("\n</details>\n")
	file.WriteString("\n-----\n")

}

func writeCount(file *os.File, no int, jpName string, enNmae string, num int64) {
	file.WriteString(fmt.Sprintf("  | %d. | %s(%s) | %d |\n", no, jpName, enNmae, num))

}
