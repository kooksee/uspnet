package app

import (
	"bufio"
	"bytes"
	"time"

	knet "k/utils/net"

	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"

	kts "p/types"
)

func UdpHandleListener(l *knet.UdpListener) {

	var message []byte
	for {
		c, err := l.Accept()
		if err != nil {
			log.Warn("Listener for incoming connections from client closed")
			return
		}

		log.Info("udp client conneted", c.RemoteAddr().String())

		// Start a new goroutine for dealing connections.
		go func(conn knet.Conn) {
			conn.SetReadDeadline(time.Now().Add(connReadTimeout))
			conn.SetReadDeadline(time.Time{})
			read := bufio.NewReader(conn)
			for {

				message, err = read.ReadBytes('\n')
				if err != nil {
					log.Info("udp error ", err.Error())
					break
				}
				message = bytes.Trim(message, "\n")

				log.Info("udp msg ", string(message))

				// 解析请求数据
				msg := &kts.KMsg{}
				if err := jsoniter.Unmarshal(message, msg); err != nil {
					log.Error(err.Error())
					conn.Write(kts.ResultError(err.Error()))
					return
				}

				switch msg.Event {

				case "tcp":
					if c, ok := tcpClients[msg.Account]; ok {
						c.Write([]byte(msg.Msg+"\n"))
						conn.Write(kts.ResultOk())
					} else {
						conn.Write(kts.ResultError("account不存在"))
					}

				case "ws":
					if c, ok := wsClients[msg.Account]; ok {
						c.WriteMessage(websocket.TextMessage, []byte(msg.Msg+"\n"))
						conn.Write(kts.ResultOk())
					} else {
						conn.Write(kts.ResultError("account不存在"))
					}
				}
			}
		}(c)
	}
}
