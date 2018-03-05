package commands

import "github.com/spf13/cobra"

func serverArgs(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&cfg().Addr, "addr", cfg().Addr, "web addr")
	cmd.Flags().StringVar(&cfg().Abci, "abci", cfg().Abci, "abci addr")
	return cmd
}

// ServerCommand
func ServerCommand() *cobra.Command {
	return serverArgs(&cobra.Command{
		Use:   "s",
		Short: "run universe node",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 初始化配置文件
			cfg().InitConfig()
			return nil
		},
	})
}
