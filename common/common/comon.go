package common

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

//获取绝对路径
func GetAbsPath(relativePath string) string {
	execPath, _ := os.Executable()
	path, _ := filepath.Split(execPath)
	if relativePath == "" {
		return ""
	}
	//兼容go run main.go模式，请在开发模式下使用，生产环境打包请注释掉
	//if gin.Mode() == gin.DebugMode {
		path, _ = os.Getwd()
	//}
	return filepath.Join(path, relativePath)
}

/*
* description: 获取当前执行程序绝对路径（不兼容于go run main.go运行模式）
 */
func GetCurrentAbsPath() string {
	execPath, _ := os.Executable()
	path, _ := filepath.Split(execPath)
	//兼容go run main.go模式，请在开发模式下使用，生产环境打包请注释掉
	//if gin.Mode() == gin.DebugMode {
	//	path, _ = os.Getwd()
	//*/
	return path
}

//获取当前时间yyyy-mm-dd hh:ii:ss
func GetTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

//判断文件是否存在
func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//判断文件是否存在,不存在则创建
func IsFileExistsAndCreate(path string) error {
	file, err := os.Open(path)
	defer func() { file.Close() }()
	if err != nil && os.IsNotExist(err) {
		file, err = os.Create(path)
	}
	return err
}

// 获取当前时间戳
func GetTimeUnix() int64 {
	return time.Now().Unix()
}

// MD5 方法
func MD5(str string) string {
	s := md5.New()
	s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

func StringIsMacAddr(macAddr string) bool {
	var trueMacAddr = `([A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2}`
	match, _ := regexp.MatchString(trueMacAddr, macAddr)
	return match
}

func StringIsIpAddr(macAddr string) bool {
	var trueIpAddr = `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])$`
	match, _ := regexp.MatchString(trueIpAddr, macAddr)
	return match
}

func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &m)
	fmt.Println(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}