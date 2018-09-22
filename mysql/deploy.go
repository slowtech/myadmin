package mysql

import (
	"fmt"
	"github.com/slowtech/myadmin/common"
)

func Initialize(configFile string, mysqld string) error{
	initializeCmd := fmt.Sprintf("%s --defaults-file=%s --initialize", mysqld, configFile)
	fmt.Println(initializeCmd)
	_, err := common.Run_cmd_direct(initializeCmd)
	return err
}

func StartMySQL(configFile string, mysqld string) {
	startCmd := fmt.Sprintf("%s --defaults-file=%s &", mysqld, configFile)
	fmt.Println(startCmd)
	common.Run_cmd_direct(startCmd)
}

func CheckInstanceAlive(pidfile string) {


}