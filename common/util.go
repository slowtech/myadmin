package common

import (
        "os"
        "os/exec"
        "fmt"
        "github.com/mattn/go-shellwords"
	"net"
)

func FileExists(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}


func FileNotExistsExit(filename string) {
        finfo,err := os.Stat(filename)
        if os.IsNotExist(err) || finfo.IsDir() {
            fmt.Printf("The file %s is not exists!\n",filename)
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

func Run_cmd(command string) (string,error) {
        cmdArray,_:= shellwords.Parse(command)
        c,args := cmdArray[0],cmdArray[1:]
        cmd := exec.Command(c, args...)
        out, err := cmd.Output()
        return string(out),err
}

func GetIP()(IpAddr string){
        addrSlice, err := net.InterfaceAddrs()
        if err != nil {
                fmt.Println(err)
		return ""
        }
        for _, addr := range addrSlice {
        if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() !=nil {
                IpAddr = ipnet.IP.String()
                break
                }
        }
        }
        return 
}
