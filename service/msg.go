/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package service

// TITLE: サービス共通

// STRUCT: レスポンスメッセージ
type Message struct {
	header string
	main   string
	detail string
}

// FUNCTION: 新規作成
func NewMessage() *Message {
	return &Message{}
}

// FUNCTION: ヘッダーメッセージの追加
func (m *Message) addHead(header string) {
	m.header += header
}

// FUNCTION: メインメッセージの追加
func (m *Message) add(main, detail string) {
	m.main += main
	m.detail += detail
}

// FUNCTION: メッセージの出力
func (m *Message) str() string {
	msg := m.header
	msg += m.main

	// PROCESS: 詳細メッセージが存在する場合には追記する
	if len(m.detail) > 0 {
		msg += "\n<details><summary>(open) modify and remove detail info</summary>\n"
		msg += m.detail
		msg += "\n</details>\n"
	}
	msg += "\n-----\n"
	return msg
}
