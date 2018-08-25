package common

import (
	"os"
	"os/exec"
	"fmt"
	"github.com/mattn/go-shellwords"
	"net"
	"strconv"
	"strings"
)

func FileExists(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func FileNotExistsExit(filename string) {
	finfo, err := os.Stat(filename)
	if os.IsNotExist(err) || finfo.IsDir() {
		fmt.Printf("The file %s is not exists!\n", filename)
		os.Exit(1)
	}
}

func Which(command string) string {
	path, err := exec.LookPath(command)
	if err == nil {
		return path
	}
	return ""
}

/*
func Run_cmd(c string, args []string) (string,error){
	cmd := exec.Command(c, args...)
	var out []byte
	var err error
	out, err = cmd.Output()
	return string(out),err
}
*/

func Run_cmd(command string) (string, error) {
	cmdArray, _ := shellwords.Parse(command)
	c, args := cmdArray[0], cmdArray[1:]
	fmt.Println(c, args)
	cmd := exec.Command(c, args...)
	out, err := cmd.Output()
	return string(out), err
}

func Run_cmd_direct(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.Output()
	return string(out), err
}

func GetIP() (IpAddr string) {
	addrSlice, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				IpAddr = ipnet.IP.String()
				break
			}
		}
	}
	return
}

func GetTotalMem() (int) {
	getMemoryCmd := `grep "MemTotal" /proc/meminfo | awk '{print $2}'`
	totalMem, err := Run_cmd_direct(getMemoryCmd)
	if err != nil {
		fmt.Println(err)
	}
	totalMemInt, _ := strconv.Atoi(strings.TrimRight(totalMem, "\n"))
	return totalMemInt / 1024
}

func GetCPUCore() (int) {
	getCPUCoreCmd := `grep "processor" /proc/cpuinfo | wc -l`
	cpuCore, err := Run_cmd_direct(getCPUCoreCmd)
	fmt.Println(cpuCore)
	if err != nil {
		fmt.Println(err)
	}
	totalCPUcore, _ := strconv.Atoi(strings.TrimRight(cpuCore, "\n"))
	fmt.Println(totalCPUcore)
	return totalCPUcore
}
