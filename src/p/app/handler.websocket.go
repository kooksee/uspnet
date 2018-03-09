package app

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Info("tcp client conneted", conn.RemoteAddr().String())

	go func(conn *websocket.Conn) {
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				if err != io.EOF {
					log.Error(err.Error())
				}
				return
			}

			if messageType != websocket.TextMessage {
				conn.WriteMessage(websocket.TextMessage, []byte("数据类型错误"))
				continue
			}

			p = bytes.Trim(p, "\n")
			if len(p) == 0 {
				continue
			}
			log.Info(string(p))
			cData := bytes.Split(p, []byte(msg_split))

			if len(cData) != 3 {
				conn.WriteMessage(websocket.TextMessage, []byte("数据解析错误"))
				continue
			}

			switch string(cData[0]) {
			case "account":
				if string(cData[2]) != cfg().Token {
					log.Error(cfg().Token)
					log.Error(string(cData[2]))
					conn.WriteMessage(websocket.TextMessage, []byte("认证失败"))
				} else {
					wsClients[string(cData[1])] = conn
					conn.WriteMessage(websocket.TextMessage, []byte("ok"))
				}

			case "tcp":
				if c, ok := tcpClients[string(cData[1])]; ok {
					c.Write(cData[2])
					conn.WriteMessage(websocket.TextMessage, []byte("ok"))
				} else {
					conn.WriteMessage(websocket.TextMessage, []byte("address不正确"))
				}

			case "ws":
				if c, ok := wsClients[string(cData[1])]; ok {
					c.WriteMessage(websocket.TextMessage, cData[2])
					conn.WriteMessage(websocket.TextMessage, []byte("ok"))
				} else {
					conn.WriteMessage(websocket.TextMessage, []byte("address不正确"))
				}
			}
		}
	}(conn)
}
