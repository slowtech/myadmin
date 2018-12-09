package cmd

import (
	"github.com/spf13/cobra"
	"strconv"
	"github.com/slowtech/myadmin/mysql"
	"fmt"
)

var (
	mycnfCmd = &cobra.Command{
		Use:   "mycnf",
		Short: "Generate my.cnf according to hardware configuration",
		Example: `
  $ myadmin mycnf --basedir /usr/local/mysql --datadir /data --memory 10G --port 3306 --ver 8.0 --ssd
  $ myadmin mycnf --datadir /data`,
		Long:  `Runs commands related to the my.cnf.`,
		Run:   GetMyCnf,
	}
	mycnf_port    int
	mycnf_basedir string
	mycnf_datadir string
	mycnf_ssd     bool
	mycnf_mgr     bool
	mycnf_memory  string
	mycnf_mysqld_version string
)

func init() {
	rootCmd.AddCommand(mycnfCmd)
	mycnfCmd.Flags().IntVarP(&mycnf_port, "port", "P", 3306, "Port number")
	mycnfCmd.Flags().StringVarP(&mycnf_basedir, "basedir", "", "/usr/local/mysql", "The path to the MySQL installation directory")
	mycnfCmd.Flags().StringVarP(&mycnf_datadir, "datadir", "", "", "The path to the MySQL server data directory")
	mycnfCmd.Flags().StringVarP(&mycnf_memory, "memory", "", "", `Server Memory,valid units are "M","G"`)
	mycnfCmd.Flags().BoolVarP(&mycnf_ssd, "ssd", "",false, "Is it SSD?")
	mycnfCmd.Flags().StringVarP(&mycnf_mysqld_version, "ver", "","5.7", "The mysqld version,valid input 5.6,5.7,8.0")
	mycnfCmd.MarkFlagRequired("datadir")
}

func GetMyCnf(cmd *cobra.Command, args []string) {
	mycnf_args := make(map[string]string)
	mycnf_args["basedir"] = mycnf_basedir
	mycnf_args["datadir"] = mycnf_datadir
	mycnf_args["port"] = strconv.Itoa(mycnf_port)
	mycnf_args["memory"] = mycnf_memory
	mycnf_args["mysqld_version"] = mycnf_mysqld_version
	if mycnf_ssd == false {
		mycnf_args["ssd"] = "0"
	}

	mycnf := mysql.GenerateMyCnf(mycnf_args)
	fmt.Println(mycnf)
}
