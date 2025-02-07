/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package transfer

import (
	"fmt"

	"github.com/teru-0529/data-transfer-sandbox/infra"
)

// TITLE: 移行共通

// STRUCT: クレンジング結果
type Status string

const MODIFY Status = "⚠<br>MODIFY"
const REMOVE Status = "⛔<br>REMOVE"

// STRUCT: クレンジング後のメッセージを管理
type Piece struct {
	status Status
	count  int
	msg    string
}

// FUNCTION: 登録なし(DBエラー)
func errorPiece(err error) *Piece {
	return removedPiece(fmt.Sprintf("<span style=\"color:red;\">%v</span>", err))
}

// FUNCTION: 登録なし
func removedPiece(msg string) *Piece {
	return &Piece{
		status: REMOVE,
		count:  -1,
		msg:    fmt.Sprintf("● %s", msg),
	}
}

// FUNCTION: 増幅等
func modifiedPiece(msg string, count int) *Piece {
	return &Piece{
		status: MODIFY,
		count:  count,
		msg:    fmt.Sprintf("● %s", msg),
	}
}

// STRUCT: インボーカークリエイター
type Creater struct {
	num   int
	ctx   infra.AppCtx
	conns infra.DbConnection
}

// FUNCTION:
func NewCreater(conns infra.DbConnection) *Creater {
	return &Creater{
		num:   0,
		ctx:   infra.NewCtx(),
		conns: conns,
	}
}

// FUNCTION: インボーカーの生成
func (c *Creater) Create(cmd Command) *Invoker {
	c.num++
	return NewInvoker(c.num, c.ctx, c.conns, cmd)
}
