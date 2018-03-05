package main

import (
	"github.com/spf13/cobra"

	cmds "p/cmd/commands"
)

func main() {

	var RootCmd = &cobra.Command{
		Use:   "srelay",
		Short: "超级代理",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
	RootCmd.AddCommand(
		cmds.ServerCommand(),
		cmds.ClientCommand(),
	)
	if err := RootCmd.Execute(); err != nil {
		panic(err)
	}
}
