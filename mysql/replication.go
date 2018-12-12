package main

import "fmt"


type master struct {
	repl_user     string
	repl_password string
	slaves         [] string
}

func (p *master) createReplUser() {
	for _,each_slave := range p.slaves {
		createUserSQL := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", p.repl_user, each_slave, p.repl_user)
		fmt.Println(createUserSQL)
		grantSQL := fmt.Sprintf("GRANT REPLICATION SLAVE ON *.* TO '%s'@'%s'", p.repl_user, each_slave)
		fmt.Println(grantSQL)
	}
}

func CreateReplUser(repl_user string,repl_password string,slaves []string) {
	var newmaster = & master {
		repl_user,repl_password,slaves,
	}
	newmaster.createReplUser()
}

type slave struct {
	master_host     string
	master_port     int
	repl_user       string
	repl_password   string
	master_log_file string
	master_log_pos  int64
	gtid            bool
	gtid_purged     string
}

func (p *slave) setGtidPurged() {
	fmt.Println("RESET MASTER")
	sql := fmt.Sprintf("SET GLOBAL GTID_PURGED='%s'", p.gtid_purged)
	fmt.Println(sql)
}

func (p *slave) changeMasterTo() {
	var changeMasterToSQL string
	if p.gtid {
		if len(p.gtid_purged) != 0 {
			p.setGtidPurged()
		}
		changeMasterToSQL = fmt.Sprintf("CHANGE MASTER TO MASTER_HOST='%s',MASTER_PORT=%d MASTER_USER='%s',MASTER_PASSWORD='%s',MASTER_AUTO_POSITION=1",p.master_host,p.master_port,p.repl_user,p.repl_password)
	} else {
		changeMasterToSQL = fmt.Sprintf("CHANGE MASTER TO MASTER_HOST='%s',MASTER_PORT=%d MASTER_USER='%s',MASTER_PASSWORD='%s',MASTER_LOG_FILE='%s',MASTER_LOG_POS=%d",p.master_host,p.master_port,p.repl_user,p.repl_password,p.master_log_file,p.master_log_pos)
	}
	fmt.Println(changeMasterToSQL)
}

func (p *slave)startSlave() {
	sql := "START SLAVE"
	fmt.Println(sql)
}

func (p *slave)checkSlaveStatus() {
	sql := "SHOW SLAVE STATUS"
	fmt.Println(sql)
}


func SetupReplication(master_host string,master_port int,repl_user string,repl_password string,master_log_file string,master_log_pos  int64,gtid bool,gtid_purged string) {
	var newslave = & slave {
		master_host,master_port,repl_user,repl_password,master_log_file,master_log_pos,gtid,gtid_purged,
	}
	newslave.changeMasterTo()
	newslave.startSlave()
	newslave.checkSlaveStatus()
}

func main() {
	CreateReplUser("repl","repl123",[]string{"192.168.244.10","192.168.244.20"})
	SetupReplication("192.168.244.10",3306,"repl","repl123","mysql-bin.00001",4,false,"")
	SetupReplication("192.168.244.10",3306,"repl","repl123","",0,true,"")
	SetupReplication("192.168.244.10",3306,"repl","repl123","",0,true,"23423sdsdfdsf")
}