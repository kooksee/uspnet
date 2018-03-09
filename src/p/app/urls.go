package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	if d, err := ioutil.ReadAll(r.Body); err != nil {
		fmt.Fprint(w, string(err.Error()))
	} else {
		cData := bytes.Split(d, []byte(msg_split))
		switch string(cData[0]) {
		case "tcp":
			if len(cData) != 3 {
				fmt.Fprint(w, "数据解析错误")
			} else {
				if c, ok := tcpClients[string(cData[1])]; ok {
					c.Write([]byte(cData[2]))
					fmt.Fprint(w, "ok")
				} else {
					fmt.Fprint(w, "address不正确")
				}
			}

		case "ws":
			if len(cData) != 3 {
				fmt.Fprint(w, "数据解析错误")
			} else {
				if c, ok := wsClients[string(cData[1])]; ok {
					c.WriteMessage(websocket.TextMessage, []byte(cData[2]))
					fmt.Fprint(w, "ok")
				} else {
					fmt.Fprint(w, "address不正确")
				}
			}
		}
	}
	return
}

func Pong(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "ok")
}
