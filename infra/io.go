/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:ファイル書き込み設定

import (
	"fmt"
	"os"
	"path/filepath"
)

// FUNCTION: ファイルへの書き込み（フォルダが無ければ作成する）
func WriteText(filePath string, msg string) error {
	// PROCESS: フォルダが存在しない場合作成する
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("cannot create directory: %s", err.Error())
		}
	}

	// PROCESS: 出力用ファイルのオープン
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create file: %s", err.Error())
	}
	defer file.Close()

	// PROCESS: 書き込み
	file.WriteString(msg)

	return nil
}
