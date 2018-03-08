package commands

import "github.com/spf13/cobra"

func clientArgs(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&cfg().KcpAddr, "kaddr", cfg().KcpAddr, "kcp addr")
	return cmd
}

func ClientCommand() *cobra.Command {
	return clientArgs(&cobra.Command{
		Use:   "c",
		Short: "run srelay client",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 初始化配置文件
			cfg().InitConfig()
			return nil
		},
	})
}
