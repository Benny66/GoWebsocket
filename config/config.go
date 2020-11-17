package config

import (
	"visualAlarmBroadcast/common/common"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type config struct {
	HttpHost  string
	HttpPort  string
	WsHost    string
	WsPort    string
	WsApi     string
	LogExpire string
	Debug     bool
}

var Cfg = config{
	HttpHost:  "127.0.0.1",
	HttpPort:  "8080",
	WsHost:    "127.0.0.1",
	WsPort:    "8080",
	WsApi:     "/echo",
	LogExpire: "72",  //小时
	Debug:     false, //小时
}

func LoadENV() {
	if !common.IsFileExists(common.GetAbsPath(".env")) {
		return
	}
	err := godotenv.Load(common.GetAbsPath(".env"))
	if err != nil {
		fmt.Println(err)
	}
	if v := os.Getenv("HttpHost"); v != "" {
		Cfg.HttpHost = v
	}
	if v := os.Getenv("HttpPort"); v != "" {
		Cfg.HttpPort = v
	}
	if v := os.Getenv("WsHost"); v != "" {
		Cfg.WsHost = v
	}
	if v := os.Getenv("WsPort"); v != "" {
		Cfg.WsPort = v
	}
	if v := os.Getenv("WsApi"); v != "" {
		Cfg.WsApi = v
	}
	if v := os.Getenv("LogExpire"); v != "" {
		Cfg.LogExpire = v
	}
	if v := os.Getenv("Debug"); v != "" {
		if v == "true" {
			Cfg.Debug = true
		}
	}
}
