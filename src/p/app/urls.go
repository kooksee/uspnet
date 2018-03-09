package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	if d, err := ioutil.ReadAll(r.Body); err != nil {
		fmt.Fprint(w, string(err.Error()))
		return
	} else {
		addrData := bytes.Split(d, []byte(msg_split))
		if len(addrData) != 2 {
			fmt.Fprint(w, "数据解析错误")
			return
		}

		if c, ok := clients[string(addrData[0])]; ok {
			c.Write(addrData[1])
			fmt.Fprint(w, "ok")
		} else {
			fmt.Fprint(w, "address不正确")
		}
		return
	}
}

func Pong(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "ok")
}
