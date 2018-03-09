package app

import (
	"net/http"
	"time"

	kcfg "p/config"

	klog "github.com/sirupsen/logrus"

	knet "k/utils/net"

	"github.com/gorilla/websocket"
)

var (
	tcpClients map[string]knet.Conn
	wsClients  map[string]*websocket.Conn
	cfg        = kcfg.GetCfg()
	log        *klog.Entry
	upgrader   = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return false
		},
	}
)

const (
	connReadTimeout time.Duration = 10 * time.Second
	msg_split                     = "[@@]"
)
