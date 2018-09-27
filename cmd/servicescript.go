package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"github.com/fatih/color"
)

var (
	servicescriptCmd = &cobra.Command{
		Use:   "service",
		Short: "Generate mysql service script",
		Example: `
  $ myadmin util password
  $ myadmin util password -L 15`,
		Long: `Generate mysql service script`,
		Run:  getServiceScript,
	}
	servicescript_type    string
	servicescript_version string
	servicescript_conf    string
	servicescript_basedir string
	servicescript_pidfile string
)

func init() {
	scriptCmd.AddCommand(servicescriptCmd)
	servicescriptCmd.Flags().StringVarP(&servicescript_type, "type", "t", "init", `Specify the service script type: "init" or "systemd"`)
	servicescriptCmd.Flags().StringVarP(&servicescript_version, "mysql-version", "v", "5.7", "The MySQL Version")
	servicescriptCmd.Flags().StringVarP(&servicescript_conf, "defaults-file", "c", "", "The default config file")
	servicescriptCmd.Flags().StringVarP(&servicescript_basedir, "basedir", "", "/usr/local/mysql", "The path to the MySQL installation directory")
	servicescriptCmd.Flags().StringVarP(&servicescript_pidfile, "pid-file", "p", "", "Pid file used by systemed.")
	servicescriptCmd.MarkFlagRequired("default-file")
}

func checkServiceScriptArgs() {
	if servicescript_type == "systemd" {
		if servicescript_pidfile == "" {
			fmt.Println(`Error: flag(s) --pid-file must be set when the service script type is "systemd"`)
			os.Exit(1)
		}
		if servicescript_version == "5.6" {
			fmt.Println(`Error: flag(s) --mysql-version can not be set "5.6" when the service script type is "systemd", Cos it's not supported`)
			os.Exit(1)
		}
	}
}


func getServiceScript(cmd *cobra.Command, args []string) {
	checkServiceScriptArgs()
	//_ := mysql.GetServiceScript(servicescript_type, servicescript_version, servicescript_conf, servicescript_basedir, servicescript_pidfile)

	d := color.New(color.FgHiBlue, color.Bold)
	d.Printf("To Start MySQL Automatically on System Startup, Do as follows.\n")

	usage:=`# cp mysqld.service /usr/lib/systemd/system/
# chmod 644 /usr/lib/systemd/system/mysqld.service 
# systemctl enable mysqld

Start MySQL Service 
# systemctl start mysqld 

Stop MySQL Service
# systemctl stop mysqld

Check MySQL Service Status  
# systemctl status mysqld
`
	fmt.Print(usage)
}


