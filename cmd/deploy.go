package cmd

import (
	"github.com/spf13/cobra"
	"github.com/slowtech/myadmin/common"
	"github.com/go-ini/ini"
	"fmt"
	"os"
	"os/user"
	"strings"
	"path/filepath"
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
	if ! common.FileExists(deploy_binary, "file") {
		fmt.Printf("%s: no such file\n", slowlog)
		os.Exit(1)
	}

	if ! common.FileExists(deploy_cnf, "file") {
		fmt.Printf("%s: no such file\n", pt)
		os.Exit(1)
	}

	variables := map[string]string{
		"user":      "",
		"basedir":   "",
		"datadir":   "",
		"log_bin":   "",
		"relay_log": "",
	}

	getConfigParameters(deploy_cnf, variables)

	//Create the user to run mysqld daemon

	runUser := variables["user"]
	if runUser == "" {
		fmt.Printf("Warning: Fail to find user in %s. Use mysql to run mysqld daemon\n", deploy_cnf)
		runUser = "mysql"
	}

	u, _ := user.Current()
	currentUser := u.Username

	if currentUser != "root" && runUser != currentUser {
		fmt.Printf("The User to run mysqld daemon is %s, But the current user is %s, I guess user %s doesn't have sufficient privileges.\n", runUser, currentUser, currentUser)
		os.Exit(1)
	}

	_, err := user.Lookup(runUser)
	if err != nil {
		var out string
		out, err = common.UserAdd(runUser)
		if err != nil {
			fmt.Println(out)
		}
		fmt.Printf("Successfully created user %s,Initial password: %s", out)
	} else {
		fmt.Printf("User %s already exist\n", runUser)
	}

	//Create the necessary directories

	basedir := variables["basedir"]
	fmt.Println(basedir)

	if basedir == "" {
		fmt.Printf("Warning: Fail to find basedir in %s. Use /user/local/mysql as the default basedir\n", deploy_cnf)
		variables["basedir"] = "/usr/local/mysql"
	}

	datadir := variables["datadir"]
	if datadir == "" {
		datadir := filepath.Join(variables["basedir"], "data")
		fmt.Printf("Warning: Fail to find datadir in %s. Use %s as the default datadir\n", deploy_cnf, datadir)
		variables["datadir"] = datadir
	}

	for k, v := range variables {
		if k == "user" || v==""{
			continue
		}
		if k != "datadir" {
			v = filepath.Dir(v)
		}
		common.MkDir(k,v)
		if k == "basedir" {
			continue
		}
		chownCmd := fmt.Sprintf("chown -R %s %s",variables["user"],v)
		common.Run_cmd_direct(chownCmd)
	}


	//untarCommand := fmt.Sprintf("tar -xvf %s -C %s", deploy_binary, filepath.Dir(variables["basedir"]))
	//fmt.Println(untarCommand)
	//var out string
	//out, err = common.Run_cmd_direct(untarCommand)
	//if err != nil {
	//	fmt.Println(out)
	//	os.Exit(1)
	//}
	untarDir := deploy_binary[0:strings.Index(deploy_binary, ".tar")]
	fmt.Println(untarDir)

	//err = os.Symlink(untarDir, basedir)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	mysqldPath := filepath.Join(basedir,"bin","mysqld")
	fmt.Println(mysqldPath)

	mysql.Initialize(deploy_cnf,mysqldPath)
	mysql.StartMySQL(deploy_cnf,mysqldPath)


	//fmt.Println(basedir)
}

func getConfigParameters(configFile string, variables map[string]string) {

	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys: true,
	}, configFile)

	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	for k, _ := range variables {
		value := cfg.Section("mysqld").Key(k).String()
		if value == "" {
			k_new := strings.Replace(k, "_", "-", -1)
			value = cfg.Section("mysqld").Key(k_new).String()
		}
		variables[k] = value
	}

}
