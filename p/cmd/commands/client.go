package commands

import (
	"p/tclient"

	"github.com/spf13/cobra"
)

func clientArgs(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&cfg().TcpAddr, "kaddr", cfg().TcpAddr, "kcp addr")
	return cmd
}

func ClientCommand() *cobra.Command {
	return clientArgs(&cobra.Command{
		Use:   "c",
		Short: "run srelay client",
		RunE: func(cmd *cobra.Command, args []string) error {
			tclient.Run()
			select {}
			return nil
		},
	})
}
