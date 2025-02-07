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

	var num int                       //TODO:
	refData := cleansing.NewRefData() //TODO:
	msg := NewMessage()

	msg.addHead("\n## Legacy Data Check and Cleansing\n\n")
	msg.addHead("  | # | TABLE | ENTRY | ELAPSED | … | UNCHANGE | MODIFY | REMOVE | … | ACCEPT | RATE |\n")
	msg.addHead("  |--:|---|--:|--:|---|--:|--:|--:|---|--:|--:|\n")

	var inv *cleansing.Invoker
	creater := cleansing.NewCreater(conns)

	// PROCESS: 1.operators
	inv = creater.Create(cleansing.NewOperatorsCmd())
	msg.add(inv.Execute())

	// TODO:
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
	// TODO:

	return msg.str()
}

// FUNCTION: 移行
func Transfer(conns infra.DbConnection) string {
	controller := transfer.New(conns)

	msg := NewMessage()
	msg.addHead(controller.Head())
	var inv *transfer.Invoker

	// PROCESS: 1.operators
	inv = controller.CreateInvocer(transfer.NewOperatorsCmd())
	msg.add(inv.Execute())

	// PROCESS: 2.products
	inv = controller.CreateInvocer(transfer.NewProductsCmd())
	msg.add(inv.Execute())

	// PROCESS: 3.orders
	inv = controller.CreateInvocer(transfer.NewOrdersCmd())
	msg.add(inv.Execute())

	// PROCESS: 4.order_details
	inv = controller.CreateInvocer(transfer.NewOrderDetailsCmd())
	msg.add(inv.Execute())

	return msg.str()
}
