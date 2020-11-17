package main

import (
	"fmt"
	"runtime"
	"time"
	"visualAlarmBroadcast/common/ws"
	"visualAlarmBroadcast/config"
	myLogger "visualAlarmBroadcast/logger"

	"github.com/kardianos/service"
)

type program struct{}


//load environmental variable.

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config.LoadENV()
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	//定义一个定时器，定时执行；使用通道去发送消息到服务端
	ticker := time.NewTicker(10 * time.Second)
	done := make(chan struct{})

	defer func() {
		defer close(done)
		ticker.Stop()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	var logger = myLogger.New("gin-", config.Cfg.LogExpire, config.Cfg.Debug)
	//logger.Write("err:")


	//发起websock连接并订阅报警消息
	wsUrl := fmt.Sprintf("ws://%s:%s%s",
		config.Cfg.WsHost,
		config.Cfg.WsPort,
		config.Cfg.WsApi,
	)
	ws.Start(wsUrl, logger)

	time.Sleep(1 * time.Second)

	//定时发起心跳检测
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if ok := ws.Status(); !ok {
				logger.Write("error：heart beat status error")
				continue
			}
			timestamp := time.Now().UnixNano() / 1e6
			err := ws.Send("heartbeat", timestamp)
			if err != nil {
				logger.Write("error：heart beat response error")
				continue
			}
		}
	}
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	svcConfig := &service.Config{
		Name:        "visualAlarmBroadcast",
		DisplayName: "visualAlarmBroadcast",
		Description: "ip广播中间服务",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Println(err)
	}
	_, err = s.Logger(nil)
	if err != nil {
		fmt.Println(err)
	}
	err = s.Run()
	if err != nil {
		fmt.Println(err)
	}
}


