package app

import (
	"bufio"
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

		log.Info("tcp client conneted", c.RemoteAddr().String())

		// Start a new goroutine for dealing connections.
		go func(conn knet.Conn) {
			conn.SetReadDeadline(time.Now().Add(connReadTimeout))
			conn.SetReadDeadline(time.Time{})
			reader := bufio.NewReader(conn)
			for {

				message, err = reader.ReadBytes('\n')
				if err != nil {
					break
				}

				// 解析请求数据
				msg := &kts.KMsg{}
				if err := jsoniter.Unmarshal(message, msg); err != nil {
					conn.Write(kts.ResultError(err.Error()))
					return
				}

				switch msg.Event {
				case "tcp":
					if c, ok := tcpClients[msg.Account]; ok {
						c.Write([]byte(msg.Msg))
						conn.Write(kts.ResultOk())
					} else {
						conn.Write(kts.ResultError("address不正确"))
					}

				case "ws":
					if c, ok := wsClients[msg.Account]; ok {
						c.WriteMessage(websocket.TextMessage, []byte(msg.Msg))
						conn.Write(kts.ResultOk())
					} else {
						conn.Write(kts.ResultError("address不正确"))
					}
				}
			}
		}(c)
	}
}
