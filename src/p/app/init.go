package app

import (
	"time"

	kcfg "p/config"

	klog "github.com/sirupsen/logrus"

	knet "k/utils/net"

	"github.com/gorilla/websocket"
)

var (
	clients   map[string]knet.Conn
	wsClients map[string]*websocket.Conn
	cfg       = kcfg.GetCfg()
	log       *klog.Entry
)

const (
	connReadTimeout time.Duration = 10 * time.Second
	msg_split                     = "[@@]"
)
