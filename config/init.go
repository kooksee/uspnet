package config

import (
	"sync"
)

var (
	once     sync.Once
	instance *Config
)

type Config struct {
	Version       uint32
	Time          int64
	Name          string `mapstructure:"name" yaml:"name"`
	Config        string `mapstructure:"config" yaml:"config"`
	Debug         bool   `mapstructure:"debug" yaml:"debug"`
	Seeds         string `mapstructure:"seeds" yaml:"seeds"`
	BindHost      string `mapstructure:"host" yaml:"host"`
	BindPort      int    `mapstructure:"port" yaml:"port"`
	AdvertiseHost string `mapstructure:"advertise_host" yaml:"advertise_host"`
	AdvertisePort int    `mapstructure:"advertise_port" yaml:"advertise_port"`
	LogLevel      string `mapstructure:"log_level" yaml:"log_level"`
}
