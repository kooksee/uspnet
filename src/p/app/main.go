package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	klog "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"

	kcfg "p/config"

	knet "k/utils/net"
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

				log.Info(message)
				cData := strings.Split(strings.Trim(message, "\n"), msg_split)

				switch cData[0] {
				case "account":
					if len(cData) != 2 {
						conn.Write([]byte("数据解析错误"))
						continue
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
						} else {
							c.Write([]byte("address不正确"))
						}
					}

				case "ws":
					if len(cData) != 3 {
						conn.Write([]byte("数据解析错误"))
					} else {
						if c, ok := wsClients[string(cData[1])]; ok {
							c.WriteMessage(websocket.TextMessage, []byte(cData[2]))
						} else {
							c.WriteMessage(websocket.TextMessage, []byte("address不正确"))
						}
					}

				}
			}
		}(c)
	}
}

func UdpHandleListener(l *knet.UdpListener) {
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
					cData := strings.Split(strings.Trim(message, "\n"), msg_split)
					if len(cData) != 2 {
						conn.Write([]byte("数据解析错误"))
						continue
					}

					if cData[0] == "account" {
						tcpClients[cData[1]] = conn
						conn.Write([]byte("ok"))
					}

				} else {
					break
				}
			}
		}(c)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

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

			cData := bytes.Split(bytes.Trim(p, "\n"), []byte(msg_split))

			switch string(cData[0]) {
			case "account":
				if len(cData) != 2 {
					conn.WriteMessage(websocket.TextMessage, []byte("数据解析错误"))
					continue
				} else {
					wsClients[string(cData[1])] = conn
					conn.WriteMessage(websocket.TextMessage, []byte("ok"))
				}

			case "tcp":
				if len(cData) != 3 {
					conn.WriteMessage(websocket.TextMessage, []byte("数据解析错误"))
				} else {
					if c, ok := tcpClients[string(cData[1])]; ok {
						c.Write(cData[2])
					} else {
						c.Write([]byte("address不正确"))
					}
				}

			case "ws":
				if len(cData) != 3 {
					conn.WriteMessage(websocket.TextMessage, []byte("数据解析错误"))
					continue
				} else {
					if c, ok := wsClients[string(cData[1])]; ok {
						c.WriteMessage(websocket.TextMessage, cData[2])
					} else {
						c.WriteMessage(websocket.TextMessage, []byte("address不正确"))
					}
				}
			}
		}
	}(conn)
}

// Run app run
func Run() {

	log = kcfg.GetLogWithFields(klog.Fields{"module": "app"})
	log.Info("start")

	// init clienst
	tcpClients = make(map[string]knet.Conn)
	wsClients = make(map[string]*websocket.Conn)

	// init tcp
	if listener, err := knet.ListenTcp(cfg().TcpAddr); err != nil {
		panic(fmt.Sprintf("Create server listener error, %v", err.Error()))
	} else {
		go TcpHandleListener(listener)
	}
	log.Info("tcp listen on", cfg().TcpAddr)

	if l, err := knet.ListenUDP(cfg().UdpAddr); err != nil {
		panic(fmt.Sprintf("Create server listener error, %v", err.Error()))
	} else {
		go UdpHandleListener(l)
	}
	log.Info("udp listen on", cfg().UdpAddr)

	// init http
	router := httprouter.New()
	//router.GET("/", knet.HttprouterBasicAuth(Index, "", ""))
	router.POST("/", Index)
	router.GET("/ping", Pong)
	if err := http.ListenAndServe(cfg().HttpAddr, router); err != nil {
		panic(err.Error())
	}
	log.Info("http listen on", cfg().HttpAddr)

	// init websocket
	http.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe(cfg().WebSocketAddr, nil); err != nil {
		panic(err.Error())
	}
	log.Info("websocket listen on", cfg().WebSocketAddr)
	return
}
