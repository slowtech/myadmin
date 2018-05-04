package cmd

import (
	"github.com/spf13/cobra"
        "fmt"
        //"reflect"
        "os"
        "strings"
        "time"
        "myadmin/common"
        "regexp"
        "html/template"
)

const temp = `
<!DOCTYPE html>
<html>
<head>
    <title>Slow Log</title>
<style>
body {
     font-family: Arial,sans-serif;
     background: #e8eaee;
     font-size: 14px;
}
.d1 {
    background: #fff;
    padding: 10px 30px 40px;
    width: 83.33333333%;
}
.d2 {
    align-items: center;
    justify-content: center;
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    margin: 0;
    padding: 0;
}
table {
    *border-collapse: collapse; /* IE7 and lower */
    border-spacing: 0;
    width: 100%;    
}
.bordered {
    border: solid #ccc 1px;
    -moz-border-radius: 6px;
    -webkit-border-radius: 6px;
    border-radius: 6px;
    -webkit-box-shadow: 0 1px 1px #ccc; 
    -moz-box-shadow: 0 1px 1px #ccc; 
    box-shadow: 0 1px 1px #ccc;         
    table-layout: fixed; 
    width:100%;  
}
.bordered tr:hover {
    background: #fbf8e9;
    -o-transition: all 0.1s ease-in-out;
    -webkit-transition: all 0.1s ease-in-out;
    -moz-transition: all 0.1s ease-in-out;
    -ms-transition: all 0.1s ease-in-out;
    transition: all 0.1s ease-in-out;     
}    
    
.bordered td, .bordered th {
    border-left: 1px solid #ccc;
    border-top: 1px solid #ccc;
    padding: 10px;
    text-align: left;    
    word-wrap:break-word;   
}
.bordered th {
    background-color: #dce9f9;
    background-image: -webkit-gradient(linear, left top, left bottom, from(#ebf3fc), to(#dce9f9));
    background-image: -webkit-linear-gradient(top, #ebf3fc, #dce9f9);
    background-image:    -moz-linear-gradient(top, #ebf3fc, #dce9f9);
    background-image:     -ms-linear-gradient(top, #ebf3fc, #dce9f9);
    background-image:      -o-linear-gradient(top, #ebf3fc, #dce9f9);
    background-image:         linear-gradient(top, #ebf3fc, #dce9f9);
    -webkit-box-shadow: 0 1px 0 rgba(255,255,255,.8) inset; 
    -moz-box-shadow:0 1px 0 rgba(255,255,255,.8) inset;  
    box-shadow: 0 1px 0 rgba(255,255,255,.8) inset;        
    border-top: none;
    text-shadow: 0 1px 0 rgba(255,255,255,.5); 
}
.bordered td:first-child, .bordered th:first-child {
    border-left: none;
}
.bordered th:first-child {
    -moz-border-radius: 6px 0 0 0;
    -webkit-border-radius: 6px 0 0 0;
    border-radius: 6px 0 0 0;
}
.bordered th:last-child {
    -moz-border-radius: 0 6px 0 0;
    -webkit-border-radius: 0 6px 0 0;
    border-radius: 0 6px 0 0;
}
.bordered th:only-child{
    -moz-border-radius: 6px 6px 0 0;
    -webkit-border-radius: 6px 6px 0 0;
    border-radius: 6px 6px 0 0;
}
.bordered tr:last-child td:first-child {
    -moz-border-radius: 0 0 0 6px;
    -webkit-border-radius: 0 0 0 6px;
    border-radius: 0 0 0 6px;
}
.bordered tr:last-child td:last-child {
    -moz-border-radius: 0 0 6px 0;
    -webkit-border-radius: 0 0 6px 0;
    border-radius: 0 0 6px 0;
}
 
</style>
</head>
<body>
<div class="d2">
<div class="d1">
<h2 style="text-align: center;margin-bottom:0px">Slow Log</h2>
<span style="font-weight: bold;float:right;font-size:12px;margin-bottom:15px">生成时间：{{.now}}</span> 
<table class="bordered">
    <thead>
    <tr>
        <th style="width:4%">Rank</th>        
        <th style="width:7%">Response time</th>
        <th style="width:6%">Response ratio</th>
        <th style="width:5%">Calls</th>        
        <th style="width:6%">R/Call</th>
        <th style="width:15%">QueryId</th>
        <th style="width:44%">Example</th>
	<th style="width:13%">Remark</th>
    </tr>
    </thead>
	{{range .slowlogs}}
    <tr>
        <td style="width:4%">{{ .Rank}}</td>        
        <td style="width:7%">{{ .Response_time}}</td>
        <td style="width:6%">{{ .Response_ratio}}</td>
	<td style="width:5%">{{ .Calls}}</td>        
        <td style="width:6%">{{ .R_Call}}</td>
        <td style="width:15%">{{ .QueryId}}</td>
        <td style="width:44%">{{ .Example}}</td>
	<td style="width:13%"> </td>   
    </tr>  
    {{end}}	
</table>
</div>
</div>
</body>
</html>
`

var (
	slowlogCmd = &cobra.Command{
		Use:   "slowlog",
		Short: "Summarize the slow log",
		Long: `Runs commands related to the slowlog.`,
                Run: GetSlowLog,
      	}
      	pt string
      	since string
      	until string
      	slowlog string
      	all bool
      	yesterday bool
)

func init() {
	rootCmd.AddCommand(slowlogCmd)
        slowlogCmd.Flags().StringVarP(&pt, "pt", "p", "/usr/local/bin/pt-query-digest", "The absolute path of pt-query-digest")
        slowlogCmd.Flags().StringVarP(&slowlog, "slowlog", "s", "", "The absolute path of slowlog")
        slowlogCmd.Flags().StringVarP(&since, "since", "", "", "Parse only queries newer than this value,YYYY-MM-DD [HH:MM:SS]")
        slowlogCmd.Flags().StringVarP(&until, "until", "", "", "Parse only queries older than this value,YYYY-MM-DD [HH:MM:SS]")
        slowlogCmd.Flags().BoolVarP(&all, "all", "a", false, "Parse the whole slowlog")
        slowlogCmd.Flags().BoolVarP(&yesterday, "yesterday", "y",true, "Parse yesterday's slowlog")
        slowlogCmd.MarkFlagRequired("slowlog")
}

func GetSlowLog(cmd *cobra.Command,args []string) {
   	ptQueryDigestCmd := checkArgs()
   	fmt.Println(ptQueryDigestCmd)
   	parseSlowLog(ptQueryDigestCmd)
}


func checkArgs() []string {
      	if all && (len(since) !=0 || len(until) !=0)  {
        	fmt.Println("--all and --since(--until) are mutually exclusive")
        	os.Exit(1)
      	}
         
      	common.FileNotExistsExit(slowlog)
      	common.FileNotExistsExit(pt)

      	parameters := make(map[string]string)
      	if all {
            parameters["since"]=""
            parameters["until"]=""
        } else if len(since) !=0 || len(until) !=0 { 
            if len(since) !=0 {
               parameters["since"]="--since "+since
            }
            if len(until) !=0 {
               parameters["until"]="--until "+until
            }
        } else {
            today := time.Now().Format("2006-01-02")
            yesterday := time.Now().AddDate(0,0,-1).Format("2006-01-02")
            parameters["since"]="--since "+yesterday
            parameters["until"]="--until "+today
        }
        
        ptQueryDigestCmd :=  []string{pt,parameters["since"],parameters["until"],slowlog}
        return ptQueryDigestCmd
  
}

func parseSlowLog(ptQueryDigestCmd []string){
	slowLog,err := common.Run_cmd(common.Which("perl"), ptQueryDigestCmd)
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
	now := time.Now().Format("2006-01-02 15:04:05")
        slowlogs := []slowlog{}
        for _,value := range slowLogProfile {
            slowlogrecord := slowlog{value[1],value[3],value[4],value[5],value[6],value[2],value[8]}
            slowlogs = append(slowlogs,slowlogrecord)
        }
        fmt.Println(slowlogs,now)
        var report = template.Must(template.New("slowlog").Parse(temp))
        report.Execute(os.Stdout,map[string]interface{}{"slowlogs":slowlogs,"now":now})
        /*
        templates := template.Must(template.ParseFiles("cmd/slowlog.html"))
        err = templates.ExecuteTemplate(os.Stdout, "slowlog.html", map[string]interface{}{"slowlogs":slowlogs,"now":now})
   	if err != nil {
        	fmt.Println("Cannot Get View ", err)
    	}
        */
}


