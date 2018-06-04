package mysql

import (
        "fmt"
        //"reflect"
        "os"
        "strings"
        "github.com/slowtech/myadmin/common"
        "regexp"
        "encoding/json"
)


func ParseSlowLog(ptQueryDigestCmd string) []byte {
	//slowLog,err := common.Run_cmd(common.Which("perl"), ptQueryDigestCmd)
	slowLog,err := common.Run_cmd(ptQueryDigestCmd)
    	if err != nil {
    		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", ptQueryDigestCmd)
		os.Exit(1)
    	}
    	lines := strings.Split(string(slowLog), "\n")
    	linesNums := len(lines)
    	profileFlag := false
    	exampleFlag := false
    	exampleSQL := []string{}
	slowLogProfile := [][]string{}
	exampleSQLs := make(map[string]string)
	var queryID string
	for k,line := range lines {
                //判断Profile部分是否开始，如果开始，则将profileFlag设置为true
		if strings.Contains(line,"# Profile"){
			profileFlag = true
                        continue
		} else if profileFlag && (len(line) == 0 || strings.HasPrefix(line,"# MISC 0xMISC")) {
                        //判断Profile是否结束
			profileFlag = false
                        continue
		}
                //如果Profile开始，则首先剔除掉Profile部分的前两行 
		if profileFlag {
			if strings.HasPrefix(line, "# Rank") || strings.HasPrefix(line, "# ====") {
				continue
			}
			re, _ := regexp.Compile(" +")
                        //将行以空格分割
			rowToArray := re.Split(line, 9)
			slowLogProfile = append(slowLogProfile, rowToArray)
		} else if strings.Contains(line,"concurrency, ID 0x"){
                        //如果某行有"0x concurrency, ID 0xF9A57DD5A41825CA"，则代表是某个query的开始，这个时候，需要获取这个查询的ID
			re := regexp.MustCompile(`(?U)ID (0x.*) `)
			queryID = re.FindStringSubmatch(line)[1]
			exampleFlag = true
                        //保存这个example SQL用了字符数组，考虑到这个SQL可能是update等，其同时会带上等价的select语句，在MySQL 5.5中，并不支持查看dml操作的执行计划。
			exampleSQL = []string{}
		}else if exampleFlag && (! strings.HasPrefix(line,"#")) && len(line) !=0 {
			exampleSQL=append(exampleSQL,line)
		}else if exampleFlag && (len(line) == 0 || k == (linesNums-1)){
			exampleFlag = false
			exampleSQLs[queryID] = strings.Join(exampleSQL,"\n")
		}
	}
   
        //基于Query ID,将Item那一列替换为具体的example SQL
        for _,v := range slowLogProfile {
            v[8] = exampleSQLs[v[2]]
           }

        type slowlog struct {
                Rank string
                Response_time string
                Response_ratio string
                Calls string
                R_Call string
                QueryId string
                Example string
        }
        
        //将二维数组转化为json
        slowlogs := []slowlog{}
        for _,value := range slowLogProfile {
            slowlogrecord := slowlog{value[1],value[3],value[4],value[5],value[6],value[2],value[8]}
            slowlogs = append(slowlogs,slowlogrecord)
        }
        b, _ := json.MarshalIndent(slowlogs,"","     ")
        return b

}
