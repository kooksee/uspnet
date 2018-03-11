package app

import (
	"bufio"
	"bytes"
	"time"

	knet "k/utils/net"

	kts "p/types"

	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
)

func TcpHandleListener(l *knet.TcpListener) {

	var message []byte

	for {
		c, err := l.Accept()
		if err != nil {
			log.Error("Listener for incoming connections from client closed")
			return
		}

		log.Info("tcp client ", c.RemoteAddr().String())

		// Start a new goroutine for dealing connections.
		go func(conn knet.Conn) {
			conn.SetReadDeadline(time.Now().Add(connReadTimeout))
			conn.SetReadDeadline(time.Time{})
			read := bufio.NewReader(conn)
			for {

				message, err = read.ReadBytes('\n')
				if err != nil {
					log.Info("tcp error ", err.Error())
					break
				}
				message = bytes.Trim(message, "\n")

				log.Info("tcp msg ", string(message))

				// 解析请求数据
				msg := &kts.KMsg{}
				if err := jsoniter.Unmarshal(message, msg); err != nil {
					conn.Write(kts.ResultError(err.Error()))
					continue
				}

				switch msg.Event {
				case "account":
					if msg.Token != cfg().Token {
						log.Error("认证失败")
						conn.Write(kts.ResultError("认证失败"))
					} else {
						log.Info("tcp client conneted ", c.RemoteAddr().String(), " ok")
						tcpClients[msg.Account] = conn
						conn.Write(kts.ResultOk())
					}

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

// {"event":"account","account":"123456","token":"123456"}
// {"event":"ws","account":"123456","msg":"hello"}
// {"event":"tcp","account":"123456","msg":"hello"}
