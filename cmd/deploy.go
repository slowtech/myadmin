package cmd

import (
	"github.com/spf13/cobra"
	"github.com/slowtech/myadmin/mysql"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy MySQL Instance",
		Long:  `Deploy a single mysql instance`,
		Run:   Deploy,
	}
	deploy_cnf    string
	deploy_binary string
)

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&deploy_cnf, "defaults-file", "c", "", "The default config file")
	deployCmd.Flags().StringVarP(&deploy_binary, "binary", "b", "", "The MySQL Binary tarball. If not specified, Assume the binary files are already in basedir")
	deployCmd.MarkFlagRequired("defaults-file")

}

func Deploy(cmd *cobra.Command, args []string) {
	mysql.DeployInstance(deploy_binary,deploy_cnf)
}

