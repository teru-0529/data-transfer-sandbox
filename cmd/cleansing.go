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

// cleansingCmd represents the cleansing command
var cleansingCmd = &cobra.Command{
	Use:   "cleansing",
	Short: "data check and clensing service from legacy database.",
	Long:  "data check and clensing service from legacy database.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// PROCESS: 現在時刻(Elapse計測用)
		now := time.Now()

		// PROCESS: config, データベース(Sqlboiler)コネクションの取得
		config, conns, cleanUp := infra.LeadConfig(version)
		defer cleanUp()
		distDir := config.CleansingDir()

		// PROCESS: クレンジング実行
		clensingMsg := service.Cleansing(conns)

		// PROCESS: データダンプ
		container := infra.NewContainer("work-db", config.WorkDB)
		filePath := path.Join(distDir, WORK_DML)
		if err := container.DumpDb(filePath, dmlWorkArgs()); err != nil {
			return err
		}

		// PROCESS: 処理時間計測
		elapse := infra.ElapsedStr(now)

		// PROCESS: Log File出力
		msg := "# Data Cleansing Result\n\n"
		msg += fmt.Sprintf("- **operation datetime**: %s\n", now.Format("2006/01/02 15:04:05"))
		msg += fmt.Sprintf("- **transfer tool version**: %s\n", config.Base.ToolVersion)
		msg += fmt.Sprintf("- **load legacy DB key**: %s\n", config.Base.LegacyDataKey)
		msg += fmt.Sprintf("- **total elapsed time**: %s\n", elapse)
		msg += clensingMsg

		logPath := path.Join(distDir, ".cleansing-log.md")
		if err := infra.WriteLog(logPath, msg, &now); err != nil {
			return err
		}

		log.Printf("total elapsed time … %s\n", elapse)
		return nil
	},
}

// FUNCTION:
func init() {
}

// FUNCTION:
func dmlWorkArgs() []string {
	return []string{
		"--no-owner",
		"--no-privileges",
		"--no-security-labels",
		"--encoding=UTF-8",
		"--format=P",
		"--data-only",
		"--schema=clean",
	}
}
