package common

import (
        "os"
        "os/exec"
        "fmt"
)

func FileExists(filename string) bool {
	_,err := os.Stat(filename)
	if os.IsNotExist(err) {
    	    return false
	}
	return true
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
