package app

import (
	"fmt"
	"net/http"

	klog "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"

	kcfg "p/config"

	knet "k/utils/net"
)

// Run app run
func Run() {

	log = kcfg.GetLogWithFields(klog.Fields{"module": "app"})
	log.Info("start")

	log.Info("init clienst")
	tcpClients = make(map[string]knet.Conn)
	wsClients = make(map[string]*websocket.Conn)

	log.Info("init tcp")
	if listener, err := knet.ListenTcp(cfg().TcpAddr); err != nil {
		panic(fmt.Sprintf("Create server listener error, %v", err.Error()))
	} else {
		go TcpHandleListener(listener)
	}
	log.Info("tcp listen on", cfg().TcpAddr)

	log.Info("init udp")
	if l, err := knet.ListenUDP(cfg().UdpAddr); err != nil {
		panic(fmt.Sprintf("Create server listener error, %v", err.Error()))
	} else {
		go UdpHandleListener(l)
	}
	log.Info("udp listen on", cfg().UdpAddr)

	log.Info("init http")
	router := httprouter.New()
	router.POST("/", index)
	router.GET("/ping", pong)
	log.Info("http listen on", cfg().HttpAddr)
	go http.ListenAndServe(cfg().HttpAddr, router)

	log.Info("init websocket")
	http.HandleFunc("/ws", wsHandler)
	log.Info("websocket listen on", cfg().WebSocketAddr)
	go http.ListenAndServe(cfg().WebSocketAddr, nil)

	return
}
