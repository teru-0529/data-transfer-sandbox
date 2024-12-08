/*
Copyright © 2024 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var full bool

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show semantic version.",
	Long:  "Show semantic version.",
	Run: func(cmd *cobra.Command, args []string) {
		if full {
			fmt.Printf("version: %s (releasedAt: %s)", version, releaseDate)
		} else {
			fmt.Print(version)
		}
	},
}

// FUNCTION:
func init() {
	// INFO:フラグ値を変数にBind
	versionCmd.Flags().BoolVarP(&full, "full", "F", false, "show with release-date by option flag.")
}
