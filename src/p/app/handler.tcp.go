package app

import (
	"io/ioutil"
	"time"

	knet "k/utils/net"

	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"

	kts "p/types"
)

func TcpHandleListener(l *knet.TcpListener) {

	var message []byte

	for {
		c, err := l.Accept()
		if err != nil {
			log.Error("Listener for incoming connections from client closed")
			return
		}
		log.Info("tcp client conneted ", c.RemoteAddr().String())

		// Start a new goroutine for dealing connections.
		go func(conn knet.Conn) {
			conn.SetReadDeadline(time.Now().Add(connReadTimeout))
			conn.SetReadDeadline(time.Time{})
			for {

				message, err = ioutil.ReadAll(conn)
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
				case "account":
					if msg.Token != cfg().Token {
						conn.Write(kts.ResultError("认证失败"))
					} else {
						tcpClients[msg.Account] = conn
						conn.Write(kts.ResultOk())
					}

				case "tcp":
					if c, ok := tcpClients[msg.Account]; ok {
						c.Write([]byte(msg.Msg))
						conn.Write(kts.ResultOk())
					} else {
						conn.Write(kts.ResultError("数据解析错误"))
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
