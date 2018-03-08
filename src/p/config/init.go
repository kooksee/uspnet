package config

import (
	"sync"
)

var (
	once     sync.Once
	instance *appConfig
)

type appConfig struct {
	Debug         bool   `mapstructure:"debug"`
	HomePath      string `mapstructure:"home"`
	TcpAddr       string `mapstructure:"tcp_addr"`
	UdpAddr       string `mapstructure:"udp_addr"`
	HttpAddr      string `mapstructure:"http_addr"`
	WebSocketAddr string `mapstructure:"ws_addr"`
	LogLevel      string `mapstructure:"log_level"`
}
