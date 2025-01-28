/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package transfer

import (
	"context"
	"fmt"
	"math"

	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// TITLE: 移行共通

// STRUCT: 検索時最大件数
const LIMIT int = 1000

// STRUCT: OPERATION-USER
var OPERATION_USER = null.StringFrom("DATA_TRANSFER")

// STRUCT: ダミーコンテキスト
var ctx context.Context = context.Background()

// STRUCT: printer(数値をカンマ区切りで出力するために利用)
var p = message.NewPrinter(language.Japanese)

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

// STRUCT: 移行結果
type Result struct {
	Schema      string
	TableNameJp string
	TableNameEn string
	EntryCount  int
	ChangeCount int
	ResultCount int
	duration    float64
}

// FUNCTION:
func (r Result) TableName() string {
	return fmt.Sprintf("%s(%s)", r.TableNameEn, r.TableNameJp)
}

// FUNCTION:
func (r Result) CheckRecord() string {
	if r.EntryCount+r.ChangeCount == r.ResultCount {
		return ""
	} else {
		return "❎"
	}
}

// FUNCTION: trauncate文の生成
func (r Result) truncateSql() string {
	return fmt.Sprintf("truncate %s.%s CASCADE;", r.Schema, r.TableNameEn)
}

// FUNCTION: Limit単位の呼び出し回数
func (r Result) sectionCount() int {
	return sectionCount(r.EntryCount)
}

// FUNCTION: ElapseTime
func (r Result) Elapsed() float64 {
	// 小数点3位で四捨五入
	return math.Round(r.duration*100) / 100
}

// FUNCTION: 移行結果の登録
func (r *Result) setResult(bp *Piece) {
	r.ChangeCount += bp.count
}

// FUNCTION: clensingResult件数
func (r *Result) ShowRecord(num int) string {
	return fmt.Sprintf("  | %d. | %s | %s | %s | %s | … | %s | … | %s | %s |\n",
		num,
		r.Schema,
		r.TableName(),
		p.Sprintf("%d", r.EntryCount),
		p.Sprintf("%3.2fs", r.Elapsed()),
		p.Sprintf("%+d", r.ChangeCount),
		p.Sprintf("%d", r.ResultCount),
		r.CheckRecord(),
	)
}

// FUNCTION: MDで赤字にする
func redFont(str string) string {
	return fmt.Sprintf("<span style=\"color:red;\">%s</span>", str)
}

// FUNCTION: Limit単位の呼び出し回数
func sectionCount(val int) int {
	// LIMIT=5
	// val=0 ・・・(0 -1 +5) / 5 = 4 / 5 = 0
	// val=1 ・・・(1 -1 +5) / 5 = 5 / 5 = 1
	// val=4 ・・・(4 -1 +5) / 5 = 8 / 5 = 1
	// val=5 ・・・(5 -1 +5) / 5 = 9 / 5 = 1
	// val=6 ・・・(6 -1 +5) / 5 = 10 / 5 = 2
	return (val - 1 + LIMIT) / LIMIT
}
