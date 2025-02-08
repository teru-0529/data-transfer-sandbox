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
	controller := cleansing.New(conns)

	msg := NewMessage()
	msg.addHead(controller.Head())
	var inv *cleansing.Invoker

	// PROCESS: 1.operators
	inv = controller.CreateInvocer(cleansing.NewOperatorsCmd())
	msg.add(inv.Execute())

	// PROCESS: 2.products
	inv = controller.CreateInvocer(cleansing.NewProductsCmd())
	msg.add(inv.Execute())

	// PROCESS: 3.orders
	inv = controller.CreateInvocer(cleansing.NewOrdersCmd())
	msg.add(inv.Execute())

	// PROCESS: 4.orders
	inv = controller.CreateInvocer(cleansing.NewOrderDetailsCmd())
	msg.add(inv.Execute())

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
