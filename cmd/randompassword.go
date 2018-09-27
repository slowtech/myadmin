package cmd

import (
	"github.com/spf13/cobra"
	"github.com/slowtech/myadmin/common"
	"fmt"
)


var (
	randomPasswordCmd = &cobra.Command{
		Use:   "password",
		Short: "Generate random password",
		Example: `
  $ myadmin util password
  $ myadmin util password -L 15`,
		Long:  `Generate random password`,
		Run:   GetRandomPassword,
	}
	randomPasswordLen int
)

func init() {
	utilCmd.AddCommand(randomPasswordCmd)
	randomPasswordCmd.Flags().IntVarP(&randomPasswordLen, "length", "L", 0, "The length of random password,if not specified,default 8~12")
}

func GetRandomPassword(cmd *cobra.Command, args []string) {
	randomPassword := common.GenerateRandomPassword(randomPasswordLen)
	fmt.Println(randomPassword)
}
