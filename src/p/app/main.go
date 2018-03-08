package app

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"

	klog "github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"

	kcfg "p/config"

	knet "k/utils/net"
)

func TcpHandleListener(l knet.Listener) {
	log.Info("Listen for incoming connections from client")
	for {
		c, err := l.Accept()
		if err != nil {
			log.Warn("Listener for incoming connections from client closed")
			return
		}

		// Start a new goroutine for dealing connections.
		go func(conn knet.Conn) {
			conn.SetReadDeadline(time.Now().Add(connReadTimeout))
			conn.SetReadDeadline(time.Time{})
			reader := bufio.NewReader(conn)
			for {
				if message, err := reader.ReadString('\n'); err == nil {
					log.Info(message)
					message = strings.Trim(message, "\n")
					cData := strings.Split(message, msg_split)
					if len(cData) != 2 {
						conn.Write([]byte("数据解析错误"))
						continue
					}

					if cData[0] == "account" {
						clients[cData[1]] = conn
						conn.Write([]byte("ok"))
					}
				} else {
					break
				}
			}
		}(c)
	}
}

// Run app run
func Run() error {

	log = kcfg.GetLogWithFields(klog.Fields{
		"module": "app",
	})

	log.Info("start")

	clients = make(map[string]knet.Conn)

	if listener, err := knet.ListenTcp(cfg().TcpAddr); err != nil {
		panic(fmt.Sprintf("Create server listener error, %v", err))
	} else {
		go TcpHandleListener(listener)
	}
	log.Info("tcp listen on", cfg().TcpAddr)

	router := httprouter.New()
	//router.GET("/", knet.HttprouterBasicAuth(Index, "", ""))
	router.POST("/", Index)
	router.GET("/ping", Pong)

	return http.ListenAndServe(":9000", router)
}
