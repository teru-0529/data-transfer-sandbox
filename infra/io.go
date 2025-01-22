/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package infra

// TITLE:ファイル書き込み設定

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// FUNCTION: ファイルへの書き込み（フォルダが無ければ作成する）
func WriteText(filePath string, msg string) error {
	dir := filepath.Dir(filePath)

	// PROCESS: フォルダが存在しない場合作成する
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

// FUNCTION: ダンプファイルを出力する（フォルダが無ければ作成する）
func WriteDump(fileName string, containerName string, config DbConfig, extArgs []string) error {
	s := time.Now()
	dir := filepath.Dir(fileName)

	// PROCESS: フォルダが存在しない場合作成する
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("cannot create directory: %s", err.Error())
		}
	}

	// PROCESS: パラメータの構築
	cmdArgs := []string{
		"exec", "-e", fmt.Sprintf("PGPASSWORD=%s", config.Password),
		"-i", containerName,
		"pg_dump",
		"-U", config.User, // ユーザー
		"-d", config.Database, // データベース名
		"-f", "/tmp/dump.sql",
	}
	cmdArgs = append(cmdArgs, extArgs...)

	// PROCESS: pg_dumpコマンドを構築
	dumpCmd := exec.Command("docker", cmdArgs...)
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr

	// PROCESS: コマンドを実行
	if err := dumpCmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %v", err)
	}

	// PROCESS: コピーコマンドを構築
	copyArgs := []string{"cp", fmt.Sprintf("%s:/tmp/dump.sql", containerName), fileName}
	copyCmd := exec.Command("docker", copyArgs...)
	copyCmd.Stdout = os.Stdout
	copyCmd.Stderr = os.Stderr

	// PROCESS: コマンドを実行
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy dump file: %v", err)
	}

	duration := time.Since(s).Seconds()
	log.Printf("pg_dump completed [%s] … %3.2fs\n", filepath.Base(fileName), duration)
	return nil
}

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
