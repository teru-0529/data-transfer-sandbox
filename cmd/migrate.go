/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teru-0529/data-transfer-sandbox/infra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "data transfer service from present system database to new system database.",
	Long:  "data transfer service from present system database to new system database.",
	Run: func(cmd *cobra.Command, args []string) {

		// PROCESS: データベース(Sqlboiler)の設定
		_, cleanUp := infra.InitDB()
		defer cleanUp()

		// config := infra.LeadEnv()
		// fmt.Println(config.SourceDB.Db)
		// fmt.Println(config.DistDB.Db)

		fmt.Println("migrate called")
	},
}

// FUNCTION:
func init() {
}
