package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"log"
	"net/url"
	"regexp"
	"sync"
	"time"
)

type Connetion struct {
	con *websocket.Conn
	mutex sync.Mutex
}
//定义命令行参数
var addr = flag.String("a", "ip:port", "http service address")
var clientUuid = flag.String("u", "", "uuid")
var c = flag.Int("c", 5, "number of connections")


func webSocketConn(wg sync.WaitGroup, msg []byte) {
	u := url.URL{Scheme: "ws", Host: *addr}
	var dialer *websocket.Dialer

	conn, _, err := dialer.Dial(u.String(), nil)

	if err != nil {
		fmt.Println(err)

		return
	}


	werr := conn.WriteMessage(websocket.TextMessage, msg)

	//fmt.Printf("发送信息：%s\n",string(msg))
	//2. 创建一个正则表达式对象
	regx, _:= regexp.Compile("\\w{8}(-\\w{4}){3}-\\w{12}")
	//3. 利用正则表达式对象, 匹配指定的字符串
	res := regx.FindString(string(msg))
	//fmt.Printf("匹配的clientId：%s\n",res)
	msg1 := make(map[string]interface{})
	msg2 := make(map[string]interface{})
	msg2["success"] = true

	msg1["clientId"] = res
	msg1["messageType"] = "ACK"
	msg1["messageId"] = "5e7d6e31e4b079c2b22876d8"
	msg1["data"] = msg2
	aMsg, _ := json.Marshal(msg1)

	if werr != nil {
		fmt.Println(werr)
	}
	//申明定时器10s，设置心跳时间为10s
	ticker := time.NewTicker(time.Second * 10)

	connect := &Connetion{
		con: conn,
	}
	//开启多线程
	go connect.timeWriter(ticker, conn)


	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		//互斥锁
		connect.mutex.Lock()
		werr2 := connect.con.WriteMessage(websocket.TextMessage, aMsg)
		connect.mutex.Unlock()

		if werr2 != nil {
			fmt.Println(werr2)
		}

		fmt.Printf("received: %s\n", message)
	}
	wg.Done()              // 每次把计数器-1

}

func (con *Connetion)timeWriter(ticker *time.Ticker, c *websocket.Conn) {


	for {
		<-ticker.C
		err := c.SetWriteDeadline(time.Now().Add(10 * time.Second))
		//fmt.Println(time.Now().Format(time.UnixDate))
		if err != nil {
			log.Printf("ping error: %s\n", err.Error())
		}

		con.mutex.Lock()
		if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("ping error: %s\n", err.Error())
		}
		con.mutex.Unlock()

	}
}


func NewConnMsg() []byte {

	msg := make(map[string]interface{})

	uuid1,_ := uuid.NewV4()
	//fmt.Printf("uuid值:%s\n",uuid1)
	id := uuid1.String()
	if *clientUuid == "" {
		msg["clientId"] = id
	} else {
		msg["clientId"] = *clientUuid
	}


	msg["messageId"] = "5e7d6e31e4b079c2b22876d8"
	msg["messageType"] = "LOGIN"
	msg["targetType"] = "PASSENGER"

	bMsg, _ := json.Marshal(msg)
	//log.Printf("%s\n", bMsg)

	return bMsg
}

func run() {

	flag.Parse()         //命令行参数
	var wg sync.WaitGroup         //申明计数器
	for i := 0; i < *c; i++ {
		wg.Add(1)         //设置计数器初始值
		go webSocketConn(wg, NewConnMsg())
		if (*c % 200) == 0 {
			time.Sleep(time.Millisecond * 50)
			//fmt.Println(time.Now().Format(time.UnixDate))
		}
	}
	log.Printf("creaate websocket connections: %v\n", *c)
	wg.Wait()            //阻塞代码的运行，直到计数器地值减为0
}


func main() {
	//NewConnMsg()
	run()
}