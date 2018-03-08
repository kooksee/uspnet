package config

import (
	"os"
	"path"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
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

func GetLogWithFields(keyvals log.Fields) *log.Entry {
	cfg := GetCfg()()

	log.Info()

	// set log output
	if cfg.Debug {
		log.SetOutput(os.Stdout)
		log.SetFormatter(&log.TextFormatter{})
	} else {

		if file, err := os.OpenFile(path.Join(cfg.HomePath, "app.log"), os.O_CREATE|os.O_WRONLY, 0666); err != nil {
			panic(err.Error())
		} else {
			log.SetFormatter(&log.JSONFormatter{})
			log.SetOutput(file)
		}
	}

	// set log level
	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		panic(err.Error())
	} else {
		log.SetLevel(l)
	}

	// set log fields
	return log.WithFields(keyvals)
}

func GetCfg() func() *appConfig {
	return func() *appConfig {
		once.Do(func() {
			instance = &appConfig{
				HomePath:      "./kdata",
				TcpAddr:       ":46380",
				UdpAddr:       ":46381",
				HttpAddr:      ":46382",
				WebSocketAddr: ":46383",
				Debug:         true,
				LogLevel:      "info",
			}
		})
		return instance
	}
}
