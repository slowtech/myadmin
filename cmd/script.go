package cmd

import (
	"github.com/spf13/cobra"
)


var (
	scriptCmd = &cobra.Command{
		Use:   "script",
		Short: "Some commonly used scripts,like mysql service script",
		Long:  `Some commonly used scripts`,
	}
)

func init() {
	rootCmd.AddCommand(scriptCmd)
}

