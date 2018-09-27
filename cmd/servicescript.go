package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"github.com/fatih/color"
	"github.com/slowtech/myadmin/mysql"
	"path/filepath"
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
	servicescriptCmd.MarkFlagRequired("defaults-file")
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
	script := mysql.GetServiceScript(servicescript_type, servicescript_version, servicescript_conf, servicescript_basedir, servicescript_pidfile)

	systemdUsage := `# cp mysqld.service /usr/lib/systemd/system/
# chmod +x /usr/lib/systemd/system/mysqld.service 
# systemctl enable mysqld

Start MySQL Service 
# systemctl start mysqld 

Stop MySQL Service
# systemctl stop mysqld

Check MySQL Service Status  
# systemctl status mysqld
`
	initUsage := `# cp mysqld /etc/init.d
# chmod +x /etc/init.d/mysqld
# chkconfig mysqld on 

Start MySQL Service
# /etc/init.d/mysqld start

Stop MySQL Service
# /etc/init.d/mysqld stop

Check MySQL Service Status  
# /etc/init.d/mysqld status
`

	var filename string
	var usage string
	if servicescript_type == "init" {
		usage = initUsage
		filename = "mysqld"
	} else {
		usage = systemdUsage
		filename = "mysqld.service"
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	f.Write(script)

	currentDir, _ := os.Getwd()
	filenameAbs := filepath.Join(currentDir, filename)
	d := color.New(color.FgGreen, color.Bold)
	d.Printf(`The script file is already saved in the current directory: %s`, filenameAbs)
	d = color.New(color.FgHiBlue, color.Bold)
	d.Println()
	d.Println()
	d.Printf("To Start MySQL Automatically on System Startup, Do as follows.\n")
	fmt.Println(usage)

}
