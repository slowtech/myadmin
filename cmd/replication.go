package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/slowtech/myadmin/common"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
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
	deploy_repl_master string
	deploy_repl_slave string
)

func init() {
	deployCmd.AddCommand(ReplicationCmd)
	ReplicationCmd.Flags().StringVarP(&deploy_repl_cnf, "defaults-file", "c", "", "The default config file")
	ReplicationCmd.Flags().StringVarP(&deploy_repl_binary, "binary", "b", "", "The MySQL Binary tarball. If not specified, Assume the binary files are already in basedir")
	ReplicationCmd.Flags().StringVarP(&deploy_repl_master, "master", "m", "", "The Master Address ")
	ReplicationCmd.Flags().StringVarP(&deploy_repl_slave, "slave", "s", "", "The Slave Address")
	ReplicationCmd.MarkFlagRequired("defaults-file")

}

func parseHostInfo(hostinfo string)([]map[string]string) {
	hostList := make([]map[string]string,0)
	hosts := strings.Split(hostinfo,";")
	for _,each_host := range hosts {
		var host = make(map[string]string)
		var ip string
		port := "22"
		ip_port := strings.Split(each_host,":")
		if len(ip_port) == 1 {
			ip = ip_port[0]
		} else {
			ip,port = ip_port[0],ip_port[1]
		}

		if ip != "localhost" {
			if ! common.CheckIP(ip) {
				log.Fatalf("%s is not an effecitve IP",ip)
			}
		}
		host["ip"]=ip
		host["port"]=port
		hostList=append(hostList, host)
	}
	return hostList
}

func DeployReplication(cmd *cobra.Command, args []string) {
	masterInfo := parseHostInfo(deploy_repl_master)
	if len(masterInfo) > 1 {
		log.Fatalf("Only one master is allowed, %s",deploy_repl_master)
	}
	slaveInfo := parseHostInfo(deploy_repl_slave)
	fmt.Println(slaveInfo)
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir)

	for _,each_host := range(append(masterInfo,slaveInfo...)) {
		fmt.Println(each_host)
		host,port:= each_host["ip"],each_host["port"]
		deploySingleInstance(host,port,"root","123456",filepath.Join(dir,"myadmin"),deploy_repl_binary,deploy_repl_cnf,"/tmp")
	}
	//var host common.Host
	//host.Init("192.168.244.30","22","root","123456")
	//host.Run("df -h")
	//t1 := time.Now().Unix()
	//host.Scp("/usr/local/goland-2018.3.1.tar.gz","/tmp/goland-2018.3.1.tar.gz")
	////mysql.DeployInstance(deploy_single_binary,deploy_single_cnf)
	//t2 := time.Now().Unix()
	//fmt.Println(t2-t1)
}

func deploySingleInstance(hostname string,port string,username string,password string,myadmin string,binary string,cnf string,dest_dir string) {
	var host common.Host
	host.Init(hostname,port,username,password)
	fmt.Println(myadmin)
	fmt.Println(dest_dir)
	host.Scp(myadmin,filepath.Join(dest_dir,"123.txt"))
}

//func main() {
//	fmt.Println(parseHostInfo("192.168.244.10;192.168.244.20:3307"))
//}
//
