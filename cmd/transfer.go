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
		dirPath := path.Join("dist", config.DirName())

		// PROCESS: データ移行実行
		transferMsg := service.Transfer(conns)

		// PROCESS: データダンプ
		container := infra.NewContainer("product-db", config.ProductDB)

		// PROCESS: ローカル用DML
		dumpfilePathLocal := path.Join(dirPath, DML_PROD_DB_LOCAL)
		if err := container.DumpDb(dumpfilePathLocal, localLoadCompose()); err != nil {
			return err
		}

		// PROCESS: AWS用DDL
		dumpfilePathDdlAws := path.Join(dirPath, DDL_PROD_DB_AWS)
		extDdlAwsArgs := []string{"--schema-only"}
		if err := container.DumpDb(dumpfilePathDdlAws, extDdlAwsArgs); err != nil {
			return err
		}

		// PROCESS: AWS用DML
		dumpfilePathDmlAws := path.Join(dirPath, DML_PROD_DB_AWS)
		extDmlAwsArgs := []string{"--data-only"}
		if err := container.DumpDb(dumpfilePathDmlAws, extDmlAwsArgs); err != nil {
			return err
		}

		// PROCESS: 処理時間計測
		elapse := infra.ElapsedStr(now)

		// PROCESS: Log File出力
		msg := "# Data Transfer Result\n\n"
		msg += fmt.Sprintf("- **operation datetime**: %s\n", now.Format("2006/01/02 15:04:05"))
		msg += fmt.Sprintf("- **transfer tool version**: %s\n", config.Base.ToolVersion)
		msg += fmt.Sprintf("- **load legacy DB key**: %s\n", config.Base.LegacyDataKey)
		msg += fmt.Sprintf("- **production schema version**: %s\n", config.Base.AppVersion)
		msg += fmt.Sprintf("- **total elapsed time**: %s\n", elapse)
		msg += transferMsg

		logPath := path.Join(dirPath, "transfer-log.md")
		if err := infra.WriteLog(logPath, msg, &now); err != nil {
			return err
		}

		// PROCESS: cleansingLogのコピー
		cleansingDir := path.Join("work", config.DirName())
		infra.FileCopy(path.Join(cleansingDir, "cleansing-log.md"), path.Join(dirPath, "cleansing-log.md"))

		log.Printf("total elapsed time … %s\n", elapse)
		return nil
	},
}

// FUNCTION:
func init() {
}

// FUNCTION:
func localLoadCompose() []string {
	return []string{
		"--data-only",
		"--schema=orders",
		// "-t orders.operators",
		"--exclude-table=orders.order_details",
	}
}
