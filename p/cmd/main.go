package main

import (
	"fmt"
	"os"

	"github.com/json-iterator/go"
	"github.com/spf13/cobra"

	cmds "p/cmd/commands"

	kcfg "p/config"
)

func main() {
	cfg := kcfg.GetCfg()()

	rootCmd := &cobra.Command{
		Use:   "srelay",
		Short: "超级代理",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {

			if _, err = os.Stat(cfg.HomePath); err == nil {
				// 初始化配置文件
				cfg.InitConfig()
			}

			if cfg.Debug {
				d, _ := jsoniter.MarshalToString(cfg)
				fmt.Println("config: ", d)
			}

			return nil
		},
	}
	rootCmd.PersistentFlags().StringVar(&cfg.HomePath, "home", "./kdata", "config home path")

	rootCmd.AddCommand(
		cmds.ServerCommand(),
		cmds.ClientCommand(),
		cmds.InitFileCommand(),
	)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
