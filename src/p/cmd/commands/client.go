package commands

import "github.com/spf13/cobra"

func clientArgs(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&cfg().Addr, "addr", cfg().Addr, "web addr")
	cmd.Flags().StringVar(&cfg().Abci, "abci", cfg().Abci, "abci addr")
	return cmd
}

func ClientCommand() *cobra.Command {
	return clientArgs(&cobra.Command{
		Use:   "c",
		Short: "run universe node",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 初始化配置文件
			cfg().InitConfig()
			return nil
		},
	})
}
