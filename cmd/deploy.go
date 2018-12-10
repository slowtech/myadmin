package cmd

import (
	"github.com/spf13/cobra"
)


var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy MySQL Instance",
		Long:  `Deploy single or replicated environment`,
	}
)

func init() {
	rootCmd.AddCommand(deployCmd)
}



