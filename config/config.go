package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
)

func (t *Config) GetBindAddr() string {
	return fmt.Sprintf("%s:%d", t.BindHost, t.BindPort)
}

func (t *Config) GetAdvertiseAddr() string {
	return fmt.Sprintf("%s:%d", t.AdvertiseHost, t.AdvertisePort)
}

func (t *Config) SetAdvertiseAddr(host string, port int) {
	t.AdvertiseHost = host
	t.AdvertisePort = port
}

func (t *Config) InitConfig() {

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	v.AddConfigPath(t.Config)

	if err := v.ReadInConfig(); err != nil {
		d, _ := json.Marshal(t)
		fmt.Println(string(d))
		panic(err.Error())
	}

	if err := v.Unmarshal(t); err != nil {
		panic(err.Error())
	}
}

func GetCfg() *Config {
	once.Do(func() {
		instance = &Config{
			Config:      "config",
			BindHost:      "0.0.0.0",
			BindPort:      8080,
			AdvertisePort: 8080,
			Debug:         true,
			LogLevel:      "info",
		}
	})
	return instance
}
