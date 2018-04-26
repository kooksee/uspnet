package main

import (
	"runtime"
	"flag"
	"github.com/kooksee/uspnet/config"
	"github.com/kooksee/uspnet/node"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GC()

	cfg := config.GetCfg()
	cfg.InitConfig()

	flag.BoolVar(&cfg.Debug, "debug", cfg.Debug, "debug mode")
	flag.StringVar(&cfg.Name, "name", cfg.Name, "app name")
	flag.IntVar(&cfg.BindPort, "port", cfg.BindPort, "app port")
	flag.StringVar(&cfg.LogLevel, "level", cfg.LogLevel, "log level")
	flag.Parse()

	n := node.NewNoder(cfg)
	if err := n.Start(); err != nil {
		panic(err.Error())
	}

	// 等待退出信号
	waitToExit()
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			fmt.Println("Ontology received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
