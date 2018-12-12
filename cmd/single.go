package cmd

import (
	"github.com/spf13/cobra"
	"github.com/slowtech/myadmin/mysql"
)

var (
	singleCmd = &cobra.Command{
		Use:   "single",
		Short: "Deploy a single mysql instance",
		Long:  `Deploy a single mysql instance`,
		Run:   DeployInstance,
	}
	deploy_single_cnf    string
	deploy_single_binary string
	deploy_initial_pass string
)

func init() {
	deployCmd.AddCommand(singleCmd)
	singleCmd.Flags().StringVarP(&deploy_single_cnf, "defaults-file", "c", "", "The default config file")
	singleCmd.Flags().StringVarP(&deploy_single_binary, "binary", "b", "", "The MySQL Binary tarball. If not specified, Assume the binary files are already in basedir")
	singleCmd.Flags().StringVarP(&deploy_initial_pass, "rootpass", "p", "", "The root initial password. If not specified, A random password will be generated")
	singleCmd.MarkFlagRequired("defaults-file")

}

func DeployInstance(cmd *cobra.Command, args []string) {
	mysql.DeployInstance(deploy_single_binary,deploy_single_cnf,deploy_initial_pass)
}
