/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

import (
	"context"
	"fmt"

	"github.com/volatiletech/null/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// TITLE:コンテキスト

// STRUCT: アプリコンテキスト
type AppCtx struct {
	Ctx           context.Context
	DateLayout    string           //日付文字列フォーマット
	Limit         int              //検索時最大件数
	OperationUser null.String      //更新ユーザー名
	Printer       *message.Printer //printer(数値をカンマ区切りで出力するために利用)
}

// FUNCTION: context setting
func NewCtx() AppCtx {
	return AppCtx{
		Ctx:           context.Background(),
		DateLayout:    "20060102",
		Limit:         10000,
		OperationUser: null.StringFrom("DATA_TRANSFER"),
		Printer:       message.NewPrinter(language.Japanese),
	}
}

// FUNCTION: Limit単位の呼び出し回数
func (ctx AppCtx) LapNumber(val int) int {
	// LIMIT=5
	// val=0 ・・・(0 -1 +5) / 5 = 4 / 5 = 0
	// val=1 ・・・(1 -1 +5) / 5 = 5 / 5 = 1
	// val=4 ・・・(4 -1 +5) / 5 = 8 / 5 = 1
	// val=5 ・・・(5 -1 +5) / 5 = 9 / 5 = 1
	// val=6 ・・・(6 -1 +5) / 5 = 10 / 5 = 2
	return (val - 1 + ctx.Limit) / ctx.Limit
}

// FUNCTION: MDで赤字にする
func (ctx AppCtx) emphasizedStr(str string) string {
	return fmt.Sprintf("<span style=\"color:red;\">%s</span>", str)
}
