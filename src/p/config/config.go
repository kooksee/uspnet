package config

import (
	"os"

	"github.com/spf13/viper"

	c "github.com/tendermint/tendermint/rpc/client"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	tlog "github.com/tendermint/tmlibs/log"
)

func (t *appConfig) InitConfig() {

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	v.AddConfigPath(t.HomePath)

	if err := v.ReadInConfig(); err != nil {
		panic("error on parsing configuration file")
	}

	if err := v.Unmarshal(t); err != nil {
		panic(err.Error())
	}
}

func Abci() *c.HTTP {
	cfg := GetCfg()()
	return c.NewHTTP(cfg.Abci, "/ws")
}

func GetLogWithKeyVals(keyvals ...interface{}) tlog.Logger {
	cfg := GetCfg()()
	klog, _ := tmflags.ParseLogLevel(
		cfg.LogLevel,
		tlog.NewTMLogger(tlog.NewSyncWriter(os.Stderr)),
		"info",
	)
	return klog.With(keyvals...)
}

func GetCfg() func() *appConfig {
	return func() *appConfig {
		once.Do(func() {
			instance = &appConfig{
				HomePath:      "./kdata",
				Addr:          ":9000",
				Abci:          "tcp://0.0.0.0:46657",
				PrivValidator: "./kdata/priv_validator.json",
				Debug:         true,
				LogLevel:      "info",
			}
		})
		return instance
	}
}
