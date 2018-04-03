package tclient

import (
	"bufio"
	"bytes"
	"time"

	kcfg "p/config"

	klog "github.com/sirupsen/logrus"

	knet "k/utils/net"
	kts "p/types"
)

// Run app run
func Run() {

	log = kcfg.GetLogWithFields(klog.Fields{"module": "tclient"})
	log.Info("start")

	log.Info("connect tcp")

	if c, err := knet.ConnectTcpServer(cfg().TcpAddr); err != nil {
		panic(err.Error())
	} else {
		go handler(c)
	}
}

func handler(c knet.Conn) {
	c.SetReadDeadline(time.Now().Add(connReadTimeout))
	c.SetReadDeadline(time.Time{})
	read := bufio.NewReader(c)

	msg := &kts.KMsg{
		Account: "123456",
		Event:   "account",
		Token:   "123456",
	}

	// 注册客户端信息
	c.Write(msg.Dumps())

	for {

		message, err := read.ReadBytes('\n')
		if err != nil {
			log.Info("tcp error ", err.Error())
			break
		}
		message = bytes.Trim(message, "\n")
		log.Info(string(message))
	}
}
