/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"context"
	"fmt"
	"math"

	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// TITLE: クレンジング共通

// STRUCT: 検索時最大件数
const LIMIT int = 1000

// STRUCT: 日付文字列フォーマット
const DATE_LAYOUT string = "20060102"

// STRUCT: OPERATION-USER
var OPERATION_USER = null.StringFrom("DATA_TRANSFER")

// STRUCT: ダミーコンテキスト
var ctx context.Context = context.Background()

// STRUCT: printer(数値をカンマ区切りで出力するために利用)
var p = message.NewPrinter(language.Japanese)

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
	p.addMessage(redFont(fmt.Sprintf("%v", err)), "")
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

// STRUCT: クレンジング結果
type Result struct {
	TableNameJp   string
	TableNameEn   string
	EntryCount    int
	UnchangeCount int
	ModifyCount   int
	RemoveCount   int
	DbCheckCount  int
	duration      float64
}

// FUNCTION:
func (r Result) TableName() string {
	return fmt.Sprintf("%s(%s)", r.TableNameEn, r.TableNameJp)
}

// FUNCTION:
func (r Result) AcceptCount() int {
	return r.UnchangeCount + r.ModifyCount
}

// FUNCTION:
func (r Result) AcceptRate() float64 {
	var rate float64
	if r.EntryCount == 0 {
		rate = 0.0
	} else {
		rate = float64(r.AcceptCount()) / float64(r.EntryCount)
	}
	// 小数点2位で四捨五入
	return math.Round(rate*1000) / 10
}

// FUNCTION: trauncate文の生成
func (r Result) truncateSql() string {
	return fmt.Sprintf("truncate clean.%s CASCADE;", r.TableNameEn)
}

// FUNCTION: Limit単位の呼び出し回数
func (r Result) sectionCount() int {
	// LIMIT=5
	// EntryCount=0 ・・・(0 -1 +5) / 5 = 4 / 5 = 0
	// EntryCount=1 ・・・(1 -1 +5) / 5 = 5 / 5 = 1
	// EntryCount=4 ・・・(4 -1 +5) / 5 = 8 / 5 = 1
	// EntryCount=5 ・・・(5 -1 +5) / 5 = 9 / 5 = 1
	// EntryCount=6 ・・・(6 -1 +5) / 5 = 10 / 5 = 2
	return (r.EntryCount - 1 + LIMIT) / LIMIT
}

// FUNCTION: trauncate文の生成
func (r Result) Elapsed() float64 {
	// 小数点3位で四捨五入
	return math.Round(r.duration*100) / 100
}

// FUNCTION: クレンジング結果の登録
func (r *Result) setResult(bp *Piece) {
	switch bp.status {
	case NO_CHANGE:
		r.UnchangeCount++
	case MODIFY:
		r.ModifyCount++
	case REMOVE:
		r.RemoveCount++
	}
	if bp.approve == NOT_FINDED {
		r.DbCheckCount++
	}
}

// FUNCTION: clensingResult件数
func (r *Result) ShowRecord(num int) string {
	var removeCountStr string
	if r.DbCheckCount > 0 {
		redmsg := redFont(p.Sprintf("※%d", r.DbCheckCount))
		removeCountStr = p.Sprintf("%d(%s)", r.RemoveCount, redmsg)
	} else {
		removeCountStr = p.Sprintf("%d", r.RemoveCount)
	}

	return fmt.Sprintf("  | %d. | %s | %s | %s | … | %s | %s | %s | … | %s | %3.1f%% |\n",
		num,
		r.TableName(),
		p.Sprintf("%d", r.EntryCount),
		p.Sprintf("%3.2fs", r.Elapsed()),
		p.Sprintf("%d", r.UnchangeCount),
		p.Sprintf("%d", r.ModifyCount),
		removeCountStr,
		// p.Sprintf("%d", r.RemoveCount),
		p.Sprintf("%d", r.AcceptCount()),
		r.AcceptRate(),
	)
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

// FUNCTION: MDで赤字にする
func redFont(str string) string {
	return fmt.Sprintf("<span style=\"color:red;\">%s</span>", str)
}
