package app

import (
	"bufio"
	"strings"
	"time"

	knet "k/utils/net"

	"github.com/gorilla/websocket"
)

func TcpHandleListener(l *knet.TcpListener) {

	var message string

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
				message, err = reader.ReadString('\n');
				if err != nil {
					break
				}

				message = strings.Trim(message, "\n")

				if message == "" {
					continue
				}

				cData := strings.Split(message, msg_split)
				log.Info(cData)

				switch cData[0] {
				case "account":
					if len(cData) != 2 {
						conn.Write([]byte("数据解析错误"))
					} else {
						tcpClients[string(cData[1])] = conn
						conn.Write([]byte("ok"))
					}

				case "tcp":
					if len(cData) != 3 {
						conn.Write([]byte("数据解析错误"))
					} else {
						if c, ok := tcpClients[string(cData[1])]; ok {
							c.Write([]byte(cData[2]))
							conn.Write([]byte("ok"))
						} else {
							conn.Write([]byte("数据解析错误"))
						}
					}

				case "ws":
					if len(cData) != 3 {
						conn.Write([]byte("数据解析错误"))
					} else {
						if c, ok := wsClients[string(cData[1])]; ok {
							c.WriteMessage(websocket.TextMessage, []byte(cData[2]))
							conn.Write([]byte("ok"))
						} else {
							conn.Write([]byte("address不正确"))
						}
					}

				}
			}
		}(c)
	}
}
