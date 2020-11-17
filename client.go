package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var conn *websocket.Conn
var err error

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	connServer()
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	done := make(chan struct{})

	//开启协程，等待接收回复的消息
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				time.Sleep(5 * time.Second)
				//connServer()
			}
			log.Printf("recv: %s", message)
		}
	}()

	sendErr := sendSubscribe(conn)
	if sendErr != nil {
		fmt.Println(sendErr)
		return
	}
	//定义一个定时器，定时执行；使用通道去发送消息到服务端
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println(t.String())
			health := make(map[string]interface{})
			health["type"] = "heartbeat"
			health["data"] = time.Now().UnixNano() / 1e6
			jsonStr, jsonStrErr := json.Marshal(health)
			if jsonStrErr != nil {
				log.Println("json write:", jsonStrErr)
				return
			}
			err := conn.WriteMessage(websocket.TextMessage, []byte(jsonStr))
			if err != nil {
				log.Println("write:", err)
				connServer()
			}
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}


}

func connServer()  {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo", RawQuery: "token=123"}
	log.Printf("connecting to %s", u.String())
	var err error
	conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("dial:", err)
	}
}

func sendSubscribe(c *websocket.Conn)(error) {
	health := make(map[string]interface{})
	health["type"] = "subscribe"
	health["data"] = "/1000"
	jsonStr, jsonStrErr := json.Marshal(health)
	if jsonStrErr != nil {
		log.Println("json write:", jsonStrErr)
		return jsonStrErr
	}
	err := c.WriteMessage(websocket.TextMessage, []byte(jsonStr))
	if err != nil {
		log.Println("write:", err)
		return err
	}
	return nil
}
