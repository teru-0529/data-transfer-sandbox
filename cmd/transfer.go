/*
Copyright © 2025 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "data transfer service to product database.",
	Long:  "data transfer service to product database.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// PROCESS: 現在時刻(Elapse計測用)
		now := time.Now()

		// PROCESS: config, データベース(Sqlboiler)コネクションの取得
		config, conns, cleanUp := infra.LeadConfig(version)
		defer cleanUp()
		distDir := config.TransferDir()

		// PROCESS: データ移行実行
		transferMsg := service.Transfer(conns)

		// PROCESS: データダンプ(ローカル用DML)
		filePathLocal := path.Join(distDir, LOCAL_DML)
		if err := config.ProductDB.Dump(filePathLocal, dmlLocalArgs()); err != nil {
			return err
		}

		// PROCESS: データダンプ(AWS用DDL)
		filePathDdl := path.Join(distDir, AWS_DDL)
		if err := config.ProductDB.Dump(filePathDdl, ddlArgs()); err != nil {
			return err
		}

		// PROCESS: データダンプ(AWS用DML)
		filePathDml := path.Join(distDir, AWS_DML)
		if err := config.ProductDB.Dump(filePathDml, dmlArgs()); err != nil {
			return err
		}

		// PROCESS: 処理時間計測
		elapse := infra.ElapsedStr(now)

		// PROCESS: Log File出力
		msg := "# Data Transfer Result\n\n"
		msg += fmt.Sprintf("- **operation datetime**: %s\n", now.Format("2006/01/02 15:04:05"))
		msg += fmt.Sprintf("- **transfer tool version**: %s\n", config.Base.ToolVersion)
		msg += fmt.Sprintf("- **production schema version**: %s\n", config.Base.AppVersion)
		msg += fmt.Sprintf("- **load legacy DB key**: %s\n", config.Base.LegacyDataKey)
		msg += fmt.Sprintf("- **total elapsed time**: %s\n", elapse)
		msg += transferMsg

		logPath := path.Join(distDir, ".transfer-log.md")
		if err := infra.WriteLog(logPath, msg, &now); err != nil {
			return err
		}

		// PROCESS: cleansingLogのコピー
		infra.FileCopy(config.CleansingDir(), distDir, ".cleansing-log.md")

		// PROCESS: clean.sql(テーブル/シーケンス/ファンクション/EnumのDROP)のコピー
		infra.FileCopy("materials", distDir, "clean.sql")

		log.Printf("total elapsed time … %s\n", elapse)
		return nil
	},
}

// FUNCTION:
func init() {
}

// FUNCTION:
func dmlLocalArgs() []string {
	return []string{
		"--no-owner",
		"--no-privileges",
		"--no-security-labels",
		"--encoding=UTF-8",
		"--format=P",
		// INFO: DML情報のみ
		"--data-only",
		// INFO: 作成対象のテーブルを記載
		"--table=orders.operators",
		"--table=orders.products",
		// INFO: Load中のトリガー無効化
		"--disable-triggers",
	}
}

// FUNCTION:
func ddlArgs() []string {
	return []string{
		"--no-owner",
		"--no-privileges",
		"--no-security-labels",
		"--encoding=UTF-8",
		"--format=P",
		// INFO: DDL情報のみ
		"--schema-only",
		// INFO: 作成対象外のスキーマを記載
		"--exclude-schema=public",
	}
}

// FUNCTION:
func dmlArgs() []string {
	return []string{
		"--no-owner",
		"--no-privileges",
		"--no-security-labels",
		"--encoding=UTF-8",
		"--format=P",
		// INFO: DML情報のみ
		"--data-only",
		// INFO: 作成対象外のテーブルを記載
		"--exclude-table=public.*",
		// INFO: Load中のトリガー無効化
		"--disable-triggers",
	}
}
