package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"

	kts "p/types"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		message []byte
		err     error
	)

	// 读取请求的内容
	message, err = ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, string(kts.ResultError(err.Error())))
		return
	}
	message = bytes.Trim(message, "\n")
	log.Info("http msg ", string(message))

	// 解析请求数据
	msg := &kts.KMsg{}
	if err := jsoniter.Unmarshal(message, msg); err != nil {
		fmt.Fprint(w, string(kts.ResultError(err.Error())))
		return
	}

	switch msg.Event {

	case "tcp":
		// 发送数据给tcp客户端

		if c, ok := tcpClients[msg.Account]; ok {
			c.Write([]byte(msg.Msg))
			fmt.Fprint(w, string(kts.ResultOk()))
		} else {
			fmt.Fprint(w, string(kts.ResultError("account不存在")))
		}

	case "ws":
		// 发送数据给ws客户端

		if c, ok := wsClients[msg.Account]; ok {
			c.WriteMessage(websocket.TextMessage, []byte(msg.Msg))
			fmt.Fprint(w, string(kts.ResultOk()))
		} else {
			fmt.Fprint(w, string(kts.ResultError("account不存在")))
		}
	}
	return
}

func pong(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, string(kts.ResultOk()))
}
