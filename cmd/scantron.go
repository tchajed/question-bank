/*
Copyright © 2026 Tej Chajed <tchajed@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var scantronCmd = &cobra.Command{
	Use:   "scantron",
	Short: "Process scantron results",
	Long:  `Commands for processing scantron CSV results from the testing center.`,
}

func init() {
	rootCmd.AddCommand(scantronCmd)
}
