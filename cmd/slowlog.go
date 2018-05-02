package cmd

import (
	"github.com/spf13/cobra"
)


var (
	slowlogCmd = &cobra.Command{
		Use:   "slowlog",
		Short: "sandbox management tasks",
		Aliases: []string{"manage"},
		Long: `Runs commands related to the slowlog.`,
	}
)
func init() {
	rootCmd.AddCommand(slowlogCmd)
}
