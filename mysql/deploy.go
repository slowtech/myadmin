package mysql

import (
	"fmt"
	"github.com/go-ini/ini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/slowtech/myadmin/common"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func DeployInstance(mysqldBinary string, configFile string) {

	if ! common.FileExists(mysqldBinary, "file") {
		log.Fatalf("%s: no such file\n", mysqldBinary)
	}

	if ! common.FileExists(configFile, "file") {
		log.Fatalf("%s: no such file\n", configFile)
	}

	variables := map[string]string{
		"user":      "",
		"basedir":   "",
		"datadir":   "",
		"log_bin":   "",
		"relay_log": "",
		"pid_file":  "",
		"log_error": "",
		"socket":    "",
	}

	getConfigParameters(configFile, variables)


	// Check whether the pid related mysqld process is exists
	if checkInstanceAlive(variables["pid_file"], 1) {
		log.Fatalf("A mysqld process already exists")
	}

	// Check whether the data directory is empty
	datadir := variables["datadir"]
    if common.FileExists(datadir,"dir") && ! common.IsEmpty(datadir) {
		log.Fatalf("Data directory %s is not empty!",datadir)
	}


	//Create the user to run mysqld daemon
	runUser := variables["user"]
	if runUser == "" {
		log.Warnf("Fail to find user in %s. Use mysql to run mysqld daemon\n", configFile)
		runUser = "mysql"
	}
	log.Infof("---- Step 1, Create user %s ----", runUser)
	createUser(runUser)

	//Create the necessary directories
	log.Infof("---- Step 2, Create the necessary directories && Chown ----")

	basedir := variables["basedir"]

	if basedir == "" {
		log.Warnf("Fail to find basedir in %s. Use /user/local/mysql as the default basedir", configFile)
		variables["basedir"] = "/usr/local/mysql"
	}

	if datadir == "" {
		datadir := filepath.Join(variables["basedir"], "data")
		log.Warnf("Fail to find datadir in %s. Use %s as the default datadir", configFile, datadir)
		variables["datadir"] = datadir
	}

	var variablesNew = make(map[string]string, 0)

	for k, v := range variables {
		if k == "user" || v == "" {
			continue
		}

		if k != "datadir" {
			v = filepath.Dir(v)
		}

		variablesNew[v] = k
	}
	var out string
	var err error

	for k, v := range variablesNew {
		common.MkDir(k)
		if v == "basedir" {
			continue
		}
		chownCmd := fmt.Sprintf("chown -R %s %s", variables["user"], k)
		_, err = common.Run_cmd_direct(chownCmd)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("---- Step 3, Untar the MySQL binary tarball && Create a soft link ----")

	untarDir := strings.Split(mysqldBinary, ".tar")[0]
	if ! common.FileExists(untarDir, "dir") {
		untarCommand := fmt.Sprintf("tar -xvf %s -C %s", mysqldBinary, filepath.Dir(variables["basedir"]))
		out, err = common.Run_cmd_direct(untarCommand)
		if err != nil {
			log.Fatalf(out)
		}
	} else {
		log.Warnf("No need to untar the MySQL binary tarball, which is already exists")
	}

	needCreateSymlink := true
	fileInfo, err := os.Lstat(basedir)
	if err == nil {
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			source_file, _ := os.Readlink(basedir)
			if source_file == filepath.Base(untarDir) {
				needCreateSymlink = false
				log.Warnf("No need to create a soft link, which is already exists")
			}
		}
	}

	if needCreateSymlink {
		err = os.Symlink(untarDir, basedir)
		if err != nil {
			log.WithFields(log.Fields{
				"LINK_NAME": basedir,
				"TARGET":    untarDir,
			}).Fatal("Fail to create a soft link")
		}
	}
	log.Infof("---- Step 4, Initialize MySQL Instance ----")

	mysqldPath := filepath.Join(basedir, "bin", "mysqld")

	//time.Sleep(1000 * time.Second)
	var initializeCmd string
	out, err, initializeCmd = initialize(configFile, mysqldPath)
	if err != nil {
		log.WithFields(log.Fields{
			"Command": initializeCmd,
		}).Fatalf("Fail to initialize mysqld\n%s",out)
		log.Fatalf("Fail to initialize mysqld\n%s", out)
	}

	log.Infof("---- Step 5, Start MySQL ----")

	mysqldSafePath := filepath.Join(basedir, "bin", "mysqld_safe")
	go startMySQL(configFile, mysqldSafePath)

	log_error := variables["log_error"]

	if ! checkInstanceAlive(variables["pid_file"], 30) {
		log.Fatalf("Fail to start mysqld, Check the error log %s in detail.",log_error)
	}

	log.Infof("---- Step 6, Reset root password ----")
	matchLines, gerr := common.GrepLine(log_error, "temporary password")
	if gerr != nil {
		fmt.Println(gerr)
	}
	temporaryPasswordLine := strings.TrimSuffix(matchLines[len(matchLines)-1], "\n")
	temporaryPasswordLineSplit := strings.Split(temporaryPasswordLine, " ")

	temporaryPassword := temporaryPasswordLineSplit[len(temporaryPasswordLineSplit)-1]
	out, err = resetPassword(temporaryPassword, variables["socket"])
	if err != nil {
		log.Warnf("Fail to reset root password: %s, Do it manually", out)
	}
	log.Infof("New Password: %s", out)
	log.Infof("Success!")
}

func createUser(runUser string) {
	u, _ := user.Current()
	currentUser := u.Username

	if currentUser != "root" && runUser != currentUser {
		log.Fatalf("The User to run mysqld daemon is %s, But the current user is %s, I guess user %s doesn't have sufficient privileges.\n", runUser, currentUser, currentUser)
	}

	_, err := user.Lookup(runUser)
	if err != nil {
		var out string
		out, err = common.UserAdd(runUser)
		if err != nil {
			fmt.Println(out)
		}
		log.Infof("Successfully created user %s,Initial password: %s", runUser, out)
	} else {
		log.Infof("No need to create, which is already exist")
	}
}

func getConfigParameters(configFile string, variables map[string]string) {

	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys: true,
	}, configFile)

	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
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

func initialize(configFile string, mysqld string) (string, error,string) {
	// Anyway,you can use --initialize-insecure,but it conflicts with plugin_load = "validate_password.so"
	initializeCmd := fmt.Sprintf("%s --defaults-file=%s --initialize", mysqld, configFile)
	out, err := common.Run_cmd_direct(initializeCmd)
	return out, err,initializeCmd
}

func startMySQL(configFile string, mysqld_safe string) {
	startCmd := fmt.Sprintf("%s --defaults-file=%s --disconnect-on-expired-password=0 &", mysqld_safe, configFile)
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
				return true
			}
		}
		time.Sleep(time.Second * 1)
	}
	return false
}

func resetPassword(initialPassword string, socket string) (string, error) {
	connectUrl := fmt.Sprintf("%s:%s@unix(%s)/mysql?charset=utf8", "root", initialPassword, socket)
	db, err := sqlx.Open(`mysql`, connectUrl)
	if err != nil {
		return "", err
	}
	randomPassword := common.GenerateRandomPassword(8)
	alterPassSQL := fmt.Sprintf("alter user root@localhost identified by '%s'", randomPassword)
	_, err = db.Query(alterPassSQL)
	if err != nil {
		return "", err
	}
	return randomPassword, nil
}
