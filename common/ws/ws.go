package ws

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
	"visualAlarmBroadcast/common/common"
	"visualAlarmBroadcast/common/http"
	"visualAlarmBroadcast/config"
	L "visualAlarmBroadcast/logger"

	"github.com/gorilla/websocket"
)

//全局数据
var (
	logger L.Logger
	addr   string //websocket地址

	client = conn{mtx: new(sync.Mutex), c: nil} // websocket client

	TokenStr    string
	TokenExpire int64
)

/**
	接受websocket响应请求数据
 */
func ReceiveMsg(res []byte) {
	var msg baseMsg
	err := json.Unmarshal(res, &msg)
	if err != nil {
		logger.Write("error: get msg data error")
		return
	}
	switch msg.Type {
		case "heartbeat":
			logger.Write("send heart status success")
		case "monitorAlarm":
			//添加业务代码，处理报警消息
			logger.Write("添加业务代码，处理报警消息")
	default:
		logger.Write("error: get msg type undefined")
	}
	return
}

/**
	连接结构体
 */
type conn struct {
	c   *websocket.Conn
	mtx *sync.Mutex
}

/**
	响应数据格式
 */
type baseMsg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

/**
	查看websocket连接状态
 */
func Status() bool {
	if client.c != nil {
		return true
	}
	return false
}

/**
	启动协程，发起连接并订阅
 */
func Start(Addr string, l L.Logger) {
	logger = l
	addr = Addr

	go connServer()
}

/**
	发送websocket请求
 */
func Send(Type string, Data interface{}) (error) {
	msg := baseMsg{
		Type: Type,
		Data: Data,
	}

	msgByte, _ := json.Marshal(msg)
	sErr := client.c.WriteMessage(websocket.BinaryMessage, msgByte)
	if sErr != nil {
		logger.Write("error: send msg failure", sErr.Error(), "data:", msg)
		return errors.New("send msg error")
	}
	return nil
}

/**
	建立连接，并订阅报警消息
 */
func connServer() {
	var err error
	defer func() {
		//连接断开，重新发起连接。3s一次
		client.mtx.Lock()
		if client.c != nil {
			_ = client.c.Close()
		}
		client.c = nil
		client.mtx.Unlock()
		logger.Write("error:", err.Error())
		time.Sleep(3 * time.Second)
		logger.Write("retry conn ", addr)
		connServer()
	}()
	//http 获取token
	//_, tokenErr := getToken()
	//if tokenErr != nil {
	//	fmt.Println(tokenErr)
	//	return
	//}
	TokenStr = "benny12123123123"
	client.c, _, err = websocket.DefaultDialer.Dial(addr + "?token=" + TokenStr, nil)

	if err != nil {
		logger.Write("error: get token ", err.Error())
		return
	}
	err2 := Send("subscribe", "/1000")
	if err2 != nil {
		logger.Write("error: subscribe 1000 ")
		return
	}
	logger.Write("info：subscribe ws success ", addr)

	for {
		var res []byte
		_, res, err = client.c.ReadMessage()
		if err != nil {
			logger.Write("error: read message ", err.Error())
			return
		}
		//处理返回数据
		ReceiveMsg(res)
	}
}

func getToken() (token string, err error) {
	currentTimestamp := common.GetTimeUnix()
	if currentTimestamp - TokenExpire < 1790 {
		logger.Write("token already exists")
		return TokenStr, nil
	}
	data := make(map[string]interface{})
	data["username"] = "admin"
	data["password"] = "123456"
	h := http.NewHttpSend("http://" + config.Cfg.HttpHost + ":" + config.Cfg.HttpPort + "/apis/login")
	h.SetSendType("JSON")
	h.SetBody(data)
	result, err2 := h.Post()
	if err2 != nil {
		logger.Write("request token error:", err2.Error())
		return "", err2
	}
	response, _ := common.JsonToMap(string(result))
	if response["token"] == "" {
		logger.Write("token not exists")
		return "", errors.New("token not exists")
	}
	TokenStr = response["token"].(string)
	TokenExpire = common.GetTimeUnix()
	return response["token"].(string), nil
}