package config

import (
	"sync"

	c "github.com/tendermint/tendermint/rpc/client"
	tlog "github.com/tendermint/tmlibs/log"
)

var (
	once     sync.Once
	instance *appConfig
	log      tlog.Logger
)

type appConfig struct {
	PubKey []byte
	PriKey []byte
	client *c.HTTP

	Debug         bool   `mapstructure:"debug"`
	HomePath      string `mapstructure:"home"`
	Addr          string `mapstructure:"addr"`
	LogLevel      string `mapstructure:"log_level"`
	Abci          string `mapstructure:"abci"`
	PrivValidator string `mapstructure:"priv_validator"`
}
