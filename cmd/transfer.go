/*
Copyright © 2025 Teruaki Sato <andrea.pirlo.0529@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "data transfer service to product database.",
	Long:  "data transfer service to product database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// FIXME:
		fmt.Println("transfer called")
		return nil
	},
}

// FUNCTION:
func init() {
}
