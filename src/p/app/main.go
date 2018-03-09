package app

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	klog "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"

	kcfg "p/config"

	knet "k/utils/net"
)

func TcpHandleListener(l knet.Listener) {
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
				if message, err := reader.ReadString('\n'); err == nil {
					log.Info(message)
					message = strings.Trim(message, "\n")
					cData := strings.Split(message, msg_split)
					if len(cData) != 2 {
						conn.Write([]byte("数据解析错误"))
						continue
					}

					if cData[0] == "account" {
						clients[cData[1]] = conn
						conn.Write([]byte("ok"))
					}
				} else {
					break
				}
			}
		}(c)
	}
}

// Run app run
func Run() {

	log = kcfg.GetLogWithFields(klog.Fields{"module": "app"})
	log.Info("start")

	// init clienst
	clients = make(map[string]knet.Conn)

	// init tcp
	if listener, err := knet.ListenTcp(cfg().TcpAddr); err != nil {
		panic(fmt.Sprintf("Create server listener error, %v", err))
	} else {
		go TcpHandleListener(listener)
	}
	log.Info("tcp listen on", cfg().TcpAddr)

	// init http
	router := httprouter.New()
	//router.GET("/", knet.HttprouterBasicAuth(Index, "", ""))
	router.POST("/", Index)
	router.GET("/ping", Pong)
	if err := http.ListenAndServe(":9000", router); err != nil {
		panic(err.Error())
	}
	log.Info("http listen on", "9000")

	// init websocket
	http.HandleFunc("/ws", handler)
	if err := http.ListenAndServe(":9001", nil); err != nil {
		panic(err.Error())
	}
	log.Info("websocket listen on", "9001")
	return
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return false
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go func(conn *websocket.Conn) {
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			if messageType != websocket.TextMessage {
				conn.WriteMessage(websocket.TextMessage, []byte("数据类型错误"))
				continue
			}

			p = bytes.Trim(p, "\n")
			cData := bytes.Split(p, []byte(msg_split))
			if len(cData) != 2 {
				conn.WriteMessage(websocket.TextMessage, []byte("数据解析错误"))
				continue
			}

			if string(cData[0]) == "account" {
				wsClients[string(cData[1])] = conn
				conn.WriteMessage(websocket.TextMessage, []byte("ok"))
			}
		}
	}(conn)
}
