/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"fmt"

	"github.com/teru-0529/data-transfer-sandbox/infra"
)

// TITLE: クレンジング共通

// STRUCT: クレンジング結果
type Status string

const NO_CHANGE Status = "⭕"
const MODIFY Status = "⚠<br>MODIFY"
const REMOVE Status = "⛔<br>REMOVE"

// STRUCT: 承認状況
type Approve string

const APPROVED Approve = "✅"
const STAY Approve = ""
const NOT_FINDED Approve = "🔰<br>CHECK!"

// STRUCT: クレンジング後のメッセージを管理
type Piece struct {
	status  Status
	approve Approve
	msg     string
}

// FUNCTION:
func NewPiece() *Piece {
	return &Piece{status: NO_CHANGE, approve: APPROVED}
}

// FUNCTION: 登録なし
func (p *Piece) isRemove() bool {
	return p.status == REMOVE
}

// FUNCTION: ワーニングあり
func (p *Piece) isWarn() bool {
	return p.status != NO_CHANGE
}

// FUNCTION: 登録なし
func (p *Piece) removed() *Piece {
	p.status = REMOVE
	return p
}

// FUNCTION: クレンジング
func (p *Piece) modified() *Piece {
	if p.status == NO_CHANGE {
		p.status = MODIFY
	}
	return p
}

// FUNCTION: DBエラー
func (p *Piece) dbError(err error) {
	p.removed()
	p.approve = NOT_FINDED
	p.addMessage(fmt.Sprintf("<span style=\"color:red;\">%v</span>", err), "")
}

// FUNCTION: 承認待ち
func (p *Piece) approveStay() *Piece {
	if p.approve == APPROVED {
		p.approve = STAY
	}
	return p
}

// FUNCTION: メッセージの追加
func (p *Piece) addMessage(msg string, id string) *Piece {
	br := ""
	_id := ""
	if len(p.msg) != 0 {
		br = "<BR>"
	}
	if id != "" {
		_id = fmt.Sprintf("[%s]", id)
	}
	p.msg += fmt.Sprintf("%s● %s %s", br, _id, msg)
	return p
}

// STRUCT: リファレンスデータ
type RefData struct {
	OperatorNameSet map[string]struct{} //担当者名
	ProductNameSet  map[string]struct{} //商品名
	OrderNoSet      map[int]struct{}    //受注番号
}

// FUNCTION: リファレンスデータの作成
func NewRefData() *RefData {
	return &RefData{
		OperatorNameSet: map[string]struct{}{},
		ProductNameSet:  map[string]struct{}{},
		OrderNoSet:      map[int]struct{}{},
	}
}

// STRUCT: コントローラー
type Controller struct {
	num     int
	ctx     infra.AppCtx
	conns   infra.DbConnection
	refData *RefData
}

// FUNCTION:
func New(conns infra.DbConnection) *Controller {
	return &Controller{
		num:     0,
		ctx:     infra.NewCtx(),
		conns:   conns,
		refData: NewRefData(),
	}
}

// FUNCTION: インボーカーの生成
func (c *Controller) CreateInvocer(cmd Command) *Invoker {
	c.num++
	return NewInvoker(c.num, c.ctx, c.conns, c.refData, cmd)
}

// FUNCTION: ヘッダーメッセージ
func (c *Controller) Head() string {
	msg := "\n## Legacy Data Check and Cleansing\n\n"
	msg += "  | # | TABLE | ENTRY | ELAPSED | … | UNCHANGE | MODIFY | REMOVE | … | ACCEPT | RATE |\n"
	msg += "  |--:|---|--:|--:|---|--:|--:|--:|---|--:|--:|\n"
	return msg
}
