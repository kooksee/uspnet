package commands

import (
	"p/app"

	"github.com/spf13/cobra"
)

func serverArgs(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&cfg().TcpAddr, "tadr", cfg().TcpAddr, "tcp addr")
	cmd.Flags().StringVar(&cfg().UdpAddr, "uaddr", cfg().UdpAddr, "udp addr")
	cmd.Flags().StringVar(&cfg().HttpAddr, "haddr", cfg().HttpAddr, "http addr")
	cmd.Flags().StringVar(&cfg().WebSocketAddr, "waddr", cfg().WebSocketAddr, "websocket addr")
	return cmd
}

// ServerCommand
func ServerCommand() *cobra.Command {
	return serverArgs(&cobra.Command{
		Use:   "s",
		Short: "run srelay server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Run()
			
			select {}
			return nil
		},
	})
}
