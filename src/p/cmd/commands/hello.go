package commands

import (
	"github.com/spf13/cobra"
	"fmt"
)

var Hello = &cobra.Command{
	Use:   "hello",
	Short: "签名命令",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("ok")
		return nil
	},
}
