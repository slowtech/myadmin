package common

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

func GetClient(hostname string,port string,username string,password string)(*ssh.Client){
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	hostaddress := strings.Join([]string{hostname, port}, ":")
	Connection, err := ssh.Dial("tcp", hostaddress, config)
	if err != nil {
		panic(err.Error())
	}
	return Connection
}

type Host struct {
	Connection *ssh.Client
}

func (host *Host) Run(cmd string) {
	session, err := host.Connection.NewSession()
	if err != nil {
		panic(err.Error())
	}
	defer session.Close()
	var buff bytes.Buffer
	session.Stdout = &buff
	if err := session.Run(cmd); err != nil {
		panic(err.Error())
	}
	fmt.Println(buff.String())
}

func (host *Host) Scp(sourcePath string,destPath string)  {
	session, err := host.Connection.NewSession()
	if err != nil {
		panic(err.Error())
	}
	defer session.Close()

	destFile:= path.Base(destPath)
	destDir := path.Dir(destPath)

	go func() {
		Buf := make([]byte, 1024*1024*10)
		w, _ := session.StdinPipe()
		defer w.Close()
		f, _ := os.Open(sourcePath)
		fileInfo, _ := f.Stat()
		fmt.Fprintln(w, "C0644", fileInfo.Size(), destFile)
		for {
			n, err := f.Read(Buf)
			fmt.Fprint(w, string(Buf[:n]))
			time.Sleep(time.Second*1)
			if err != nil {
				if err == io.EOF {
					return
				} else {
					panic(err)
				}
			}
		}
	}()
	if err := session.Run("/usr/bin/scp -qrt "+ destDir); err != nil {
			fmt.Println(err)
		}

}
