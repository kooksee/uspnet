package app

import (
	"time"

	kcfg "p/config"

	klog "github.com/sirupsen/logrus"

	knet "k/utils/net"
)

var (
	clients map[string]knet.Conn
	cfg     = kcfg.GetCfg()
	log     *klog.Entry
)

const (
	connReadTimeout time.Duration = 10 * time.Second
	msg_split                     = "[@@]"
)
