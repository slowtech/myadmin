package mysql

import (
	"fmt"
	"github.com/slowtech/myadmin/common"
	"strings"
	"path/filepath"
	"os"
	"os/user"
	"github.com/go-ini/ini"
	"io/ioutil"
	"time"
	"strconv"
)

func DeployInstance(mysqldBinary string, configFile string) {

	if ! common.FileExists(mysqldBinary, "file") {
		fmt.Printf("%s: no such file\n", mysqldBinary)
		os.Exit(1)
	}

	if ! common.FileExists(configFile, "file") {
		fmt.Printf("%s: no such file\n", configFile)
		os.Exit(1)
	}

	variables := map[string]string{
		"user":      "",
		"basedir":   "",
		"datadir":   "",
		"log_bin":   "",
		"relay_log": "",
		"pid_file":  "",
	}

	getConfigParameters(configFile, variables)

	//Create the user to run mysqld daemon
	runUser := variables["user"]
	if runUser == "" {
		fmt.Printf("Warning: Fail to find user in %s. Use mysql to run mysqld daemon\n", configFile)
		runUser = "mysql"
	}
	createUser(runUser)

	//Create the necessary directories
	basedir := variables["basedir"]
	fmt.Println(basedir)

	if basedir == "" {
		fmt.Printf("Warning: Fail to find basedir in %s. Use /user/local/mysql as the default basedir\n", configFile)
		variables["basedir"] = "/usr/local/mysql"
	}

	datadir := variables["datadir"]
	if datadir == "" {
		datadir := filepath.Join(variables["basedir"], "data")
		fmt.Printf("Warning: Fail to find datadir in %s. Use %s as the default datadir\n", configFile, datadir)
		variables["datadir"] = datadir
	}

	for k, v := range variables {
		if k == "user" || v == "" {
			continue
		}
		if k != "datadir" {
			v = filepath.Dir(v)
		}
		common.MkDir(k, v)

		if k == "basedir" {
			continue
		}
		chownCmd := fmt.Sprintf("chown -R %s %s", variables["user"], v)

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
	untarDir := mysqldBinary[0:strings.Index(mysqldBinary, ".tar")]
	fmt.Println(untarDir)

	//err = os.Symlink(untarDir, basedir)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	mysqldPath := filepath.Join(basedir, "bin", "mysqld")
	fmt.Println(mysqldPath)

	//time.Sleep(1000 * time.Second)
	if ! initialize(configFile, mysqldPath) {
		fmt.Println("Fail to initialize mysqld, Check the error log in detail.")
		os.Exit(1)
	}
	startMySQL(configFile, mysqldPath)

	if ! checkInstanceAlive(variables["pid_file"], 30) {
		fmt.Println("Fail to start mysqld, Check the error log in detail.")
		os.Exit(1)
	}

	//fmt.Println(basedir)
}

func createUser(runUser string) {
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

func initialize(configFile string, mysqld string) bool {
	initializeCmd := fmt.Sprintf("%s --defaults-file=%s --initialize", mysqld, configFile)
	fmt.Println(initializeCmd)
	_, err := common.Run_cmd_direct(initializeCmd)
	if err != nil {
		return false
	}
	return true
}

func startMySQL(configFile string, mysqld string) {
	startCmd := fmt.Sprintf("%s --defaults-file=%s &", mysqld, configFile)
	fmt.Println(startCmd)
	common.Run_cmd_direct(startCmd)
}

func checkInstanceAlive(pidfile string, timeout int) bool {
	for i := 0; i < timeout; i++ {
		if common.FileExists(pidfile, "file") {
			content, err := ioutil.ReadFile(pidfile)
			if err != nil {
				fmt.Println(err)
			}
			pid, _ := strconv.Atoi(strings.TrimSuffix(string(content), "\n"))
			err = common.CheckProcessAlive(pid)
			if err == nil {
				fmt.Println("Success")
				return true
			}
		}
		time.Sleep(time.Second * 1)
	}
	return false
}
