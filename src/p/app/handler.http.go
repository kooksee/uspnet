package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	if d, err := ioutil.ReadAll(r.Body); err != nil {
		fmt.Fprint(w, string(err.Error()))
	} else {
		cData := bytes.Split(d, []byte(msg_split))

		if len(cData) != 3 {
			fmt.Fprint(w, "数据解析错误")
			return
		}

		switch string(cData[0]) {
		case "tcp":
			if c, ok := tcpClients[string(cData[1])]; ok {
				c.Write(cData[2])
				fmt.Fprint(w, "ok")
			} else {
				fmt.Fprint(w, "address不正确")
			}

		case "ws":
			if c, ok := wsClients[string(cData[1])]; ok {
				c.WriteMessage(websocket.TextMessage, cData[2])
				fmt.Fprint(w, "ok")
			} else {
				fmt.Fprint(w, "address不正确")
			}
		}
	}
	return
}

func pong(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "ok")
}
