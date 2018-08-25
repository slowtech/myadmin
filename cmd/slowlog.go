package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	//"reflect"
	"os"
	"github.com/slowtech/myadmin/common"
	"html/template"
	"github.com/slowtech/myadmin/mysql"
	"encoding/json"
	"time"
	"bufio"
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
        <th style="width:5%">Rank</th>        
        <th style="width:8%">Response time</th>
        <th style="width:7%">Response ratio</th>
        <th style="width:6%">Calls</th>        
        <th style="width:6%">R/Call</th>
        <th style="width:13%">QueryId</th>
        <th style="width:44%">Example</th>
	<th style="width:11%">Remark</th>
    </tr>
    </thead>
	{{range .slowlogs}}
    <tr>
        <td style="width:5%">{{ .Rank}}</td>        
        <td style="width:8%">{{ .Response_time}}</td>
        <td style="width:7%">{{ .Response_ratio}}</td>
	<td style="width:6%">{{ .Calls}}</td>        
        <td style="width:6%">{{ .R_Call}}</td>
        <td style="width:13%">{{ .QueryId}}</td>
        <td style="width:44%">{{ .Example}}</td>
	<td style="width:11%"> </td>   
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
		Long:  `Runs commands related to the slowlog.`,
		Run:   GetSlowLog,
	}
	pt        string
	since     string
	until     string
	slowlog   string
	all       bool
	yesterday bool
	output    string
)

func init() {
	rootCmd.AddCommand(slowlogCmd)
	slowlogCmd.Flags().StringVarP(&pt, "pt", "p", "/usr/local/bin/pt-query-digest", "The absolute path of pt-query-digest")
	slowlogCmd.Flags().StringVarP(&slowlog, "slowlog", "s", "", "The absolute path of slowlog")
	slowlogCmd.Flags().StringVarP(&since, "since", "", "", "Parse only queries newer than this value,YYYY-MM-DD [HH:MM:SS]")
	slowlogCmd.Flags().StringVarP(&until, "until", "", "", "Parse only queries older than this value,YYYY-MM-DD [HH:MM:SS]")
	slowlogCmd.Flags().BoolVarP(&all, "all", "a", false, "Parse the whole slowlog")
	slowlogCmd.Flags().BoolVarP(&yesterday, "yesterday", "y", true, "Parse yesterday's slowlog")
	slowlogCmd.Flags().StringVarP(&output, "output", "o", "", "Specify the file name to save the output")
	slowlogCmd.MarkFlagRequired("slowlog")
}

func GetSlowLog(cmd *cobra.Command, slowlog_args []string) {
	ptQueryDigestCmd := checkslowlog_args()
	slowlogResult := mysql.ParseSlowLog(ptQueryDigestCmd)
	type slowlog struct {
		Rank           string
		Response_time  string
		Response_ratio string
		Calls          string
		R_Call         string
		QueryId        string
		Example        string
	}

	//将二维数组转化为json
	slowlogs := []slowlog{}

	json.Unmarshal(slowlogResult, &slowlogs)

	now := time.Now().Format("2006-01-02 15:04:05")
	var report = template.Must(template.New("slowlog").Parse(temp))

	f, _ := os.Create(output)
	w := bufio.NewWriter(f)

	report.Execute(w, map[string]interface{}{"slowlogs": slowlogs, "now": now})

	w.Flush()
	f.Close()
	fmt.Printf("Success,Check \"%s\"!\n", output)


	//templates := template.Must(template.ParseFiles("cmd/slowlog.html"))
	//err = templates.ExecuteTemplate(os.Stdout, "slowlog.html", map[string]interface{}{"slowlogs":slowlogs,"now":now})
	//if err != nil {
	//	fmt.Println("Cannot Get View ", err)
	//}
}

func checkslowlog_args() string {
	if all && (len(since) != 0 || len(until) != 0) {
		fmt.Println("--all and --since(--until) are mutually exclusive")
		os.Exit(1)
	}

	common.FileNotExistsExit(slowlog)
	common.FileNotExistsExit(pt)

	slowlog_args := make(map[string]string)
	if all {
		slowlog_args["since"] = ""
		slowlog_args["until"] = ""
	} else if len(since) != 0 || len(until) != 0 {
		if len(since) != 0 {
			slowlog_args["since"] = "--since '" + since + "'"
		}
		if len(until) != 0 {
			slowlog_args["until"] = "--until '" + until + "'"
		}
	} else {
		today := time.Now().Format("2006-01-02")
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		slowlog_args["since"] = "--since " + yesterday
		slowlog_args["until"] = "--until " + today
	}

	if len(output) == 0 {
		output = fmt.Sprintf("%s_%s.html", "/tmp/slowlog", time.Now().Format("2006_01_02_15_04_05"))
	}

	if common.FileExists(output) {
		fmt.Printf("The file %s is already exists!\n", output)
		os.Exit(1)
	}

	//ptQueryDigestCmd :=  []string{pt,slowlog_args["since"],slowlog_args["until"],slowlog}
	ptQueryDigestCmd := fmt.Sprintf("%s %s %s %s", pt, slowlog_args["since"], slowlog_args["until"], slowlog)
	return ptQueryDigestCmd

}
