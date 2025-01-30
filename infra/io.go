/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:ファイル書き込み設定

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// FUNCTION: ファイルへの書き込み（sfxを指定した場合は、ファイル名の最後にsfxを付けた履歴も保存する）
func WriteLog(filePath string, msg string, timestamp *time.Time) error {

	// PROCESS: 最新ログを削除
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println(err)
	}

	// PROCESS: ログの書き込み
	if err := WriteText(filePath, msg); err != nil {
		return err
	}

	if timestamp != nil {
		// PROCESS: 履歴パスの構成
		ext := filepath.Ext(filePath)
		tsStr := timestamp.Format("060102-150405")
		histryFile := strings.Replace(filepath.Base(filePath), ext, fmt.Sprintf("-%s%s", tsStr, ext), 1)
		histryFilePath := filepath.Clean(path.Join(filepath.Dir(filePath), path.Join("history", histryFile)))

		// PROCESS: 履歴ログの書き込み
		if err := WriteText(histryFilePath, msg); err != nil {
			return err
		}
	}

	return nil
}

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

// FUNCTION: ファイルのコピー
func FileCopy(srcPath string, distPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(distPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}
