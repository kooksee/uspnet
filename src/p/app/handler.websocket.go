package app

import (
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"

	kts "p/types"
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
					conn.WriteMessage(websocket.TextMessage, kts.ResultError(err.Error()))
				}
				return
			}

			if messageType != websocket.TextMessage {
				conn.WriteMessage(websocket.TextMessage, kts.ResultError("数据类型错误"))
				continue
			}

			// 解析请求数据
			msg := &kts.KMsg{}
			if err := jsoniter.Unmarshal(p, msg); err != nil {
				conn.WriteMessage(websocket.TextMessage, kts.ResultError(err.Error()))
				return
			}

			switch msg.Event {
			case "account":
				if msg.Token != cfg().Token {
					conn.WriteMessage(websocket.TextMessage, kts.ResultError("认证失败"))
				} else {
					wsClients[msg.Account] = conn
					conn.WriteMessage(websocket.TextMessage, kts.ResultOk())
				}

			case "tcp":
				if c, ok := tcpClients[msg.Account]; ok {
					c.Write([]byte(msg.Msg))
					conn.WriteMessage(websocket.TextMessage, kts.ResultOk())
				} else {
					conn.WriteMessage(websocket.TextMessage, kts.ResultError("address不正确"))
				}

			case "ws":
				if c, ok := wsClients[msg.Account]; ok {
					c.WriteMessage(websocket.TextMessage, []byte(msg.Msg))
					conn.WriteMessage(websocket.TextMessage, kts.ResultOk())
				} else {
					conn.WriteMessage(websocket.TextMessage, kts.ResultError("address不正确"))
				}
			}
		}
	}(conn)
}
