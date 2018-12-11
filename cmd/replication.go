package cmd

import (
	"fmt"
	"github.com/slowtech/myadmin/common"
	"github.com/spf13/cobra"
	"time"
)

var (
	ReplicationCmd = &cobra.Command{
		Use:   "replication",
		Short: "Deploy replication",
		Long:  `Deploy replication`,
		Run:   DeployReplication,
	}
	deploy_repl_cnf    string
	deploy_repl_binary string
)

func init() {
	deployCmd.AddCommand(ReplicationCmd)
	ReplicationCmd.Flags().StringVarP(&deploy_repl_cnf, "defaults-file", "c", "", "The default config file")
	ReplicationCmd.Flags().StringVarP(&deploy_repl_binary, "binary", "b", "", "The MySQL Binary tarball. If not specified, Assume the binary files are already in basedir")
	ReplicationCmd.MarkFlagRequired("defaults-file")

}

func DeployReplication(cmd *cobra.Command, args []string) {
	var host common.Host
	host.Init("192.168.244.30","22","root","123456")
	host.Run("df -h")
	t1 := time.Now().Unix()
	host.Scp("/usr/local/goland-2018.3.1.tar.gz","/tmp/goland-2018.3.1.tar.gz")
	//mysql.DeployInstance(deploy_single_binary,deploy_single_cnf)
	t2 := time.Now().Unix()
	fmt.Println(t2-t1)
}

