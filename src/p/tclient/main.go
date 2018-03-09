package tclient

import (
	"io/ioutil"
	"time"

	kcfg "p/config"

	klog "github.com/sirupsen/logrus"

	knet "k/utils/net"
)

// Run app run
func Run() {

	log = kcfg.GetLogWithFields(klog.Fields{"module": "app"})
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

	for {
		message, err := ioutil.ReadAll(c)
		if err != nil {
			break
		}
		log.Info(string(message))
	}
}
