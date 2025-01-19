/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/teru-0529/data-transfer-sandbox/infra"
	"github.com/teru-0529/data-transfer-sandbox/service"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "data transfer service from present system database to new system database.",
	Long:  "data transfer service from present system database to new system database.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// PROCESS: Fileの取得
		now := time.Now()
		path := path.Join("dist", fmt.Sprintf("migration-log-%s.md", now.Format("060102-150405")))
		file, fCleanup, err := infra.NewFile(path)
		if err != nil {
			return err
		}
		defer fCleanup()
		baseInfo(file, now)

		// PROCESS: データベース(Sqlboiler)コネクションの取得
		conns, dCleanUp := infra.InitDB()
		defer dCleanUp()

		// service.LegacyInfo(file, conns)
		service.Cleansing(file, conns)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(num)
		// config := infra.LeadEnv()
		// fmt.Println(config.SourceDB.Db)
		// fmt.Println(config.DistDB.Db)

		fmt.Println("migrate called")
		return nil
	},
}

// FUNCTION:
func init() {
}

// FUNCTION: 基本情報の書き込み
func baseInfo(file *os.File, now time.Time) {
	file.WriteString("# Data Transfer Result\n\n")
	file.WriteString(fmt.Sprintf("- **operation datetime**: %s\n", now.Format("2006/01/02 15:04:05")))
	file.WriteString(fmt.Sprintf("- **product version**: %s\n", version))
	file.WriteString("\n-----\n")
}
