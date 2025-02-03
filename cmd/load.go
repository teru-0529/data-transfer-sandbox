/*
Copyright © 2025 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "clean database data load.",
	Long:  "clean database data load.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// PROCESS: 現在時刻(Elapse計測用)
		now := time.Now()

		// PROCESS: config, データベース(Sqlboiler)コネクションの取得
		config, conns, cleanUp := infra.LeadConfig(version)
		defer cleanUp()
		distDir := config.CleansingDir()

		// PROCESS: ファイルが存在しない場合エラー
		loadfilePath := path.Join(distDir, WORK_DML)
		if f, err := os.Stat(loadfilePath); os.IsNotExist(err) || f.IsDir() {
			return fmt.Errorf("not exist loadfile[%s]: %s", loadfilePath, err.Error())
		}

		// PROCESS: データロード先(workDB)トランケート
		service.TruncateCleanDbAll(conns)

		// PROCESS: データロード
		container := infra.NewContainer("work-db", config.WorkDB)
		if err := container.LoadDb(loadfilePath); err != nil {
			return err
		}

		// PROCESS: 処理時間計測
		log.Printf("total elapsed time … %s\n", infra.ElapsedStr(now))
		return nil
	},
}

// FUNCTION:
func init() {
}
