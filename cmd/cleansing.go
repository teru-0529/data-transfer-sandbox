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

		// STRUCT: 現在時刻(Elapse計測用)
		now := time.Now()

		// PROCESS: config, データベース(Sqlboiler)コネクションの取得
		config, conns, cleanUp := infra.LeadConfig()
		defer cleanUp()

		// PROCESS: クレンジング実行
		clensingMsg := service.Cleansing(conns)

		// PROCESS: cleanダンプ
		dumpPath := path.Join("dist/cleansing", fmt.Sprintf("%s-clean-dump-%s.sql", config.Base.LegacyFile, now.Format("060102-150405")))
		extArgs := []string{"--data-only", "--schema=clean"}
		infra.WriteDump(dumpPath, "work-db", config.WorkDB, extArgs)
		// TODO: cleanダンプファイル更新

		// PROCESS: 処理時間計測
		elapse := tZero.Add(time.Duration(time.Since(now))).Format("15:04:05.000")

		// PROCESS: Log File出力
		msg := "# Data Cleansing Result\n\n"
		msg += fmt.Sprintf("- **operation datetime**: %s\n", now.Format("2006/01/02 15:04:05"))
		msg += fmt.Sprintf("- **transfer tool version**: %s\n", version)
		msg += fmt.Sprintf("- **load legacy file**: %s\n", fmt.Sprintf("%s.sql", config.Base.LegacyFile))
		msg += fmt.Sprintf("- **total elapsed time**: %s\n", elapse)
		msg += clensingMsg

		logPath := path.Join("log/cleansing", fmt.Sprintf("cleansing-log-%s.md", now.Format("060102-150405")))
		if err := infra.WriteText(logPath, msg); err != nil {
			return err
		}

		log.Printf("total elapsed time … %s\n", elapse)
		return nil
	},
}

func init() {
}
