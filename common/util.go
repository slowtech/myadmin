package common

import (
	"os"
	"os/exec"
	"fmt"
	"github.com/mattn/go-shellwords"
	"net"
	"strconv"
	"strings"
	"math/rand"
	"time"
	"bytes"
	"syscall"
	"bufio"
	"io"
	log "github.com/sirupsen/logrus"
)

func FileExists(filename string, filetype string) bool {
	f, err := os.Stat(filename)
	if os.IsPermission(err) {
		log.Fatal(err)
	}
	if os.IsNotExist(err) {
		return false
	}
	if filetype == "file" {
		return ! f.IsDir()
	}
	if filetype == "dir" {
		return f.IsDir()
	}
	return true
}

//
//func FileNotExistsExit(filename string) {
//	finfo, err := os.Stat(filename)
//	if os.IsNotExist(err) || finfo.IsDir() {
//		//fmt.Printf("The file %s does not exist!\n", filename)
//		fmt.Printf("%s: no such file\n", filename)
//		os.Exit(1)
//	}
//}

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
	log.Infof("Command: %s", command)
	cmd := exec.Command("bash", "-c", command)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), err
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
	out, err := Run_cmd_direct(getMemoryCmd)
	if err != nil {
		fmt.Println(out)
	}
	totalMemInt, _ := strconv.Atoi(strings.TrimRight(out, "\n"))
	return totalMemInt / 1024
}

func GetCPUCore() (int) {
	getCPUCoreCmd := `grep "processor" /proc/cpuinfo | wc -l`
	out, err := Run_cmd_direct(getCPUCoreCmd)
	if err != nil {
		fmt.Println(out)
	}
	totalCPUcore, _ := strconv.Atoi(strings.TrimRight(out, "\n"))
	fmt.Println(totalCPUcore)
	return totalCPUcore
}

func GenerateRandomPassword(requirePasswordLen int) (string) {
	var chars = map[string]string{
		"upchars":  "QWERTYUIOPASDFGHJKLZXCVBNM",
		"lowchars": "qwertyuiopasdfghjklzxcvbnm",
		"numchars": "1234567890",
		"symchars": ",.-+*;:_!#%&/()=?><",
	}
	var charTypes = [4]string{"upchars", "lowchars", "numchars", "symchars"}
	rand.Seed(time.Now().UnixNano())
	var maxPasswordLength int
	if requirePasswordLen == 0 {
		maxPasswordLength = 8 + rand.Intn(4)
	} else {
		maxPasswordLength = requirePasswordLen
	}
	pickCharTypes := charTypes[:]
	for i := 0; i < maxPasswordLength-4; i++ {
		pickCharTypes = append(pickCharTypes, charTypes[rand.Intn(4)])
	}
	rand.Shuffle(len(pickCharTypes), func(i, j int) {
		pickCharTypes[i], pickCharTypes[j] = pickCharTypes[j], pickCharTypes[i]
	})
	var password = make([]byte, 0)
	for _, k := range pickCharTypes {
		charsValue := chars[k]
		password = append(password, charsValue[rand.Int()%len(charsValue)])
	}
	return string(password)
}

func UserAdd(user string) (string, error) {
	cmd := fmt.Sprintf("useradd %s", user)
	out, err := Run_cmd_direct(cmd)
	if err != nil {
		return out, err
	}
	randomPasswd := GenerateRandomPassword(8)
	cmd = fmt.Sprintf("echo '%s' | passwd --stdin %s", randomPasswd, user)
	out, err = Run_cmd_direct(cmd)
	if err != nil {
		return out, err
	}
	return randomPasswd, nil
}

func MkDir(dir string) {
	log.WithFields(log.Fields{
		"Dir": dir,
	}).Infof("Create Directory")
	if ! FileExists(dir, "dir") {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.WithFields(log.Fields{
			"Dir": dir,
		}).Warnf("Directory is already exist")
	}
}

func CheckProcessAlive(pid int) error {
	process, err := os.FindProcess(pid)
	//On Unix systems, FindProcess always succeeds and returns a Process for the given pid, regardless of whether the process exists.
	if err != nil {
		log.Warnf("Failed to find process: %s\n", err)
		return err
	}
	err = process.Signal(syscall.Signal(0))
	//fmt.Printf("process.Signal on pid %d returned: %v\n", pid, err)
	return err
}

func GrepLine(file string, pattern string) ([] string, error) {
	var matchLines = make([]string, 0)
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("error opening file ", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString(10)
		if err == io.EOF {

			if strings.Contains(line, pattern) {
				matchLines = append(matchLines, line)
			}

			break
		} else if err != nil {
			return []string{}, err
		}

		if strings.Contains(line, pattern) {
			matchLines = append(matchLines, line)
		}
	}
	return matchLines, nil
}
