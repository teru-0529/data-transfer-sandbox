/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cleansing

import (
	"context"
	"fmt"
	"math"

	"github.com/volatiletech/null/v8"
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

// STRUCT: クレンジング結果
type Status string

const NO_CHANGE Status = "⭕"
const MODIFY Status = "⚠<br>MODIFY"
const REMOVE Status = "⛔<br>REMOVE"

// STRUCT: クレンジング結果
type Result struct {
	TableNameJp   string
	TableNameEn   string
	EntryCount    int
	UnchangeCount int
	ModifyCount   int
	RemoveCount   int
	duration      float64
}

// FUNCTION:
func approveStr(apploved bool) string {
	if apploved {
		return "✅"
	} else {
		return ""
	}
}

// FUNCTION:
func judgeStatus(baseStatus, status Status) Status {
	if status == REMOVE || baseStatus == REMOVE {
		return REMOVE
	} else if status == MODIFY || baseStatus == MODIFY {
		return MODIFY
	} else {
		return NO_CHANGE
	}
}

// FUNCTION:
func genMessage(msg, appendMsg string, id string) string {
	br := ""
	_id := ""
	if len(msg) != 0 {
		br = "<BR>"
	}
	if id != "" {
		_id = fmt.Sprintf("[%s]", id)
	}
	return fmt.Sprintf("%s%s● %s %s", msg, br, _id, appendMsg)
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

// STRUCT: リファレンスデータ
type RefData struct {
	OperatorNameSet map[string]struct{} //担当者名
}

// FUNCTION: リファレンスデータの作成
func NewRefData() *RefData {
	return &RefData{
		OperatorNameSet: map[string]struct{}{},
	}
}
