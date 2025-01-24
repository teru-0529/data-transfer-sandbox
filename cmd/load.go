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

var loadkey string

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "clean database data load.",
	Long:  "clean database data load.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// STRUCT: 現在時刻(Elapse計測用)
		now := time.Now()

		// PROCESS: ファイルが存在しない場合エラー
		loadfile := fmt.Sprintf("%s.sql.gz", loadkey)
		loadfilePath := path.Join("dist", path.Join("cleansing", loadfile))
		if f, err := os.Stat(loadfilePath); os.IsNotExist(err) || f.IsDir() {
			return fmt.Errorf("not exist dumpfile[%s]: %s", loadfile, err.Error())
		}

		// PROCESS: config, データベース(Sqlboiler)コネクションの取得
		config, conns, cleanUp := infra.LeadConfig()
		defer cleanUp()

		// PROCESS: cleanDBトランケート
		service.TruncateCleanDbAll(conns)

		// PROCESS: データロード
		container := infra.NewContainer("work-db", config.WorkDB)
		if err := container.LoadDb(loadfilePath); err != nil {
			return err
		}
		service.RegisterDumpName(conns, loadfile)

		// PROCESS: 処理時間計測
		elapse := tZero.Add(time.Duration(time.Since(now))).Format("15:04:05.000")
		log.Printf("total elapsed time … %s\n", elapse)
		return nil
	},
}

// FUNCTION:
func init() {
	// PROCESS:フラグ値を変数にBind
	loadCmd.Flags().StringVarP(&loadkey, "loaddata", "L", "", "load data key.")
	loadCmd.MarkFlagRequired("loaddata")
}
