/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

import (
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service/cleansing"
	"github.com/teru-0529/data-transfer-sandbox/service/transfer"
)

// TITLE: サービス共通

// FUNCTION: クレンジング
func Cleansing(conns infra.DbConnection) string {

	var num int
	refData := cleansing.NewRefData()
	msg := NewMessage()

	msg.addHead("\n## Legacy Data Check and Cleansing\n\n")
	msg.addHead("  | # | TABLE | ENTRY | ELAPSED | … | UNCHANGE | MODIFY | REMOVE | … | ACCEPT | RATE |\n")
	msg.addHead("  |--:|---|--:|--:|---|--:|--:|--:|---|--:|--:|\n")

	// PROCESS: 1.operators
	num++
	cs1 := cleansing.NewOperators(conns, refData)
	msg.add(cs1.Result.ShowRecord(num), cs1.ShowDetails())

	// PROCESS: 2.products
	num++
	cs2 := cleansing.NewProducts(conns, refData)
	msg.add(cs2.Result.ShowRecord(num), cs2.ShowDetails())

	// PROCESS: 3.orders
	num++
	cs3 := cleansing.NewOrders(conns, refData)
	msg.add(cs3.Result.ShowRecord(num), cs3.ShowDetails())

	// PROCESS: 4.order_details
	num++
	cs4 := cleansing.NewOrderDetails(conns, refData)
	msg.add(cs4.Result.ShowRecord(num), cs4.ShowDetails())

	return msg.str()
}

// FUNCTION: 移行
func Transfer(conns infra.DbConnection) string {

	var num int
	msg := NewMessage()

	msg.addHead("\n## Data Transfer to Production DB\n\n")
	msg.addHead("  | # | SCHEMA | TABLE | ENTRY | ELAPSED | … | CHANGE | … | ACCEPT | CHECK |\n")
	msg.addHead("  |--:|---|---|--:|--:|---|--:|---|--:|:--:|\n")

	// PROCESS: 1.operators
	num++
	cs1 := transfer.NewOperators(conns)
	msg.add(cs1.Result.ShowRecord(num), cs1.ShowDetails())

	// PROCESS: 2.products
	num++
	cs2 := transfer.NewProducts(conns)
	msg.add(cs2.Result.ShowRecord(num), cs2.ShowDetails())

	// PROCESS: 3.orders
	num++
	cs3 := transfer.NewOrders(conns)
	msg.add(cs3.Result.ShowRecord(num), cs3.ShowDetails())

	// PROCESS: 4.order_details
	num++
	cs4 := transfer.NewOrderDetails(conns)
	msg.add(cs4.Result.ShowRecord(num), cs4.ShowDetails())

	return msg.str()
}
