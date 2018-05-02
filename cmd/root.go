package cmd

import (
	"os"
        "github.com/spf13/cobra"
	"fmt"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "myadmin",
	Short: "Installs multiple MySQL servers on the same host",
	Long: `myadmin provides a comprehensive way to manage MySQL`,
	Version: "0.1",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

