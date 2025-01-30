/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// STRUCT: リリース情報
var (
	version     string
	releaseDate string
)

// STRUCT: ワークDBのダンプファイル名
const DML_WORK_DB = "dml-work.sql.gz"
const DML_PROD_DB_LOCAL = "dml-local.sql.gz"
const DDL_PROD_DB_AWS = "ddl-aws.sql.gz"
const DML_PROD_DB_AWS = "dml-aws.sql.gz"

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "data-transfer-sandbox",
	Short: "data transfer service from present system database to new system database.",
	Long:  "data transfer service from present system database to new system database.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// FUNCTION:
func Execute(ver string, date string) {
	version = ver
	releaseDate = date

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// FUNCTION:
func init() {
	// PROCESS:サブコマンドの追加
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cleansingCmd)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(transferCmd)
}
