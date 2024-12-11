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

// FUNCTION: 書き込みファイルの準備（フォルダが無ければ作成する）
func NewFile(fileName string) (*os.File, func(), error) {
	dir := filepath.Dir(fileName)

	// PROCESS: フォルダが存在しない場合作成する
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return nil, nil, fmt.Errorf("cannot create directory: %s", err.Error())
		}
	}

	// PROCESS: 出力用ファイルのオープン
	file, err := os.Create(fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create file: %s", err.Error())
	}
	return file, func() { file.Close() }, nil
}
