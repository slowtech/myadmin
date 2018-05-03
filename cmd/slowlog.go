package cmd

import (
	"github.com/spf13/cobra"
        "fmt"
        //"reflect"
        //"os"
        "strings"
        "time"
        "myadmin/common"
)


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
   checkArgs()
}


func checkArgs(){
      if all && (len(since) !=0 || len(until) !=0)  {
         fmt.Println("--all and --since(--until) are mutually exclusive")
         return
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
        
        ptQueryDigestCmd :=  strings.Join([]string{common.Which("perl"),pt,parameters["since"],parameters["until"],slowlog}," ")
        fmt.Println(ptQueryDigestCmd)
  
}

