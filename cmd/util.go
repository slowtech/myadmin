package cmd

import (
	"github.com/spf13/cobra"
)


var (
	utilCmd = &cobra.Command{
		Use:   "util",
		Short: "Some useful gadgets,like generate random password",
		Long:  `Some useful gadgets`,
	}
)

func init() {
	rootCmd.AddCommand(utilCmd)
}

