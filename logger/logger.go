package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"
	"visualAlarmBroadcast/common/common"

	"github.com/gin-gonic/gin"
)

type Logger interface {
	Write(...interface{})
	Writeln(...interface{})
	Print(...interface{})
	GinLogger() gin.HandlerFunc
	GinRecoveryWriter() io.Writer
}

type logger struct {
	out           io.Writer
	logFilePrefix string
	logPath       string
	logExpire     string
	debug         bool
	pkgPrefix     string
	Logger
}

// new logger struct.
func New(LogFilePrefix, LogExpire string, Debug bool) Logger {
	l := &logger{
		logPath:       common.GetAbsPath("runtime/logs"),
		logFilePrefix: LogFilePrefix,
		logExpire:     LogExpire,
		debug:         Debug,
		pkgPrefix:     strings.Replace(nameOfFunction(nameOfFunction), string(os.PathSeparator)+"logger.nameOfFunction", "", -1) + string(os.PathSeparator),
	}

	l.checkLoggerWriter()
	l.startLogsCleaner()

	return l
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func (l *logger) Write(logs ...interface{}) {
	if l.out == nil {
		fmt.Println("no set logger out")
		l.out = os.Stdout
	}

	s := ""
	for _, v := range logs {
		s += " " + fmt.Sprint(v) + " "
	}

	_, f, line, _ := runtime.Caller(1)

	fArr := strings.Split(f, l.pkgPrefix)
	if len(fArr) == 2 {
		f = l.pkgPrefix + fArr[1]
	}

	l.checkLoggerWriter()

	_, _ = fmt.Fprintf(l.out, "[LOG] %s %s:%d \n%s\n",
		time.Now().Format("15:04:05.999999999"),
		f,
		line,
		s,
	)

	if l.debug == true {
		_, _ = fmt.Fprintf(io.MultiWriter(l.out, os.Stdout), "[LOG] %s %s:%d \n%s\n",
			time.Now().Format("15:04:05.999999999"),
			f,
			line,
			s,
		)
	}
}

func (l *logger) Writeln(logs ...interface{}) {
	if l.out == nil {
		fmt.Println("no set logger out")
		l.out = os.Stdout
	}

	s := ""
	for _, v := range logs {
		s += "    " + fmt.Sprint(v) + "\n"
	}

	_, f, line, _ := runtime.Caller(1)
	fArr := strings.Split(f, l.pkgPrefix)
	if len(fArr) == 2 {
		f = l.pkgPrefix + fArr[1]
	}

	l.checkLoggerWriter()

	_, _ = fmt.Fprintf(l.out, "[LOG] %s %s:%d \n%s",
		time.Now().Format("15:04:05.999999999"),
		f,
		line,
		s,
	)

	if l.debug == true {
		_, _ = fmt.Fprintf(os.Stdout, "[LOG] %s %s:%d \n%s",
			time.Now().Format("15:04:05.999999999"),
			f,
			line,
			s,
		)
	}
}

//自定义日志文件 按天记录
func (l *logger) checkLoggerWriter() {
	l.checkLoggerDir()

	name := l.logFilePrefix + time.Now().Format("2006-01-02") + ".log"
	file := filepath.Join(l.logPath, name)

	_, err := os.Stat(file)
	if err != nil {
		if f, ok := l.out.(*os.File); ok && l.out != nil {
			_ = f.Close()
		}

		var err error
		l.out, err = os.Create(file)
		if err != nil {
			fmt.Println("create", file, "err :", err.Error())
			l.out = os.Stdout
		}
		return
	}

	if l.out != nil {
		return
	}

	var err1 error
	l.out, err1 = os.OpenFile(file, syscall.O_APPEND|syscall.O_RDWR, 0666)
	if err1 != nil {
		fmt.Println("OpenFile", file, "err :", err1.Error())
		l.out = os.Stdout
	}

	return
}

//初始化生成日志文件目录
// 配置runtime目录 .env RUNTIME_PATH
// 配置logs目录名 .env LOG_PATH_NAME
func (l *logger) checkLoggerDir() {
	{
		_, err := os.Stat(l.logPath)
		if err != nil {
			err := os.MkdirAll(l.logPath, 0777)
			if err != nil {
				panic("mkdir " + l.logPath + " err:" + err.Error())
			}
		}
	}

	{
		_, err := os.Stat(l.logPath)
		if err != nil {
			err := os.Mkdir(l.logPath, 0777)
			if err != nil {
				panic("mkdir " + l.logPath + " err:" + err.Error())
			}
		}
	}
}

//定时清除日志
// 配置 .env LOG_EXPIRE 设置日志过期时间 以小时为单位
func (l *logger) startLogsCleaner() {
	day, _ := time.ParseDuration(l.logExpire + "h")
	t := time.NewTicker(day)

	go func() {
		for {
			select {
			case <-t.C:
				l.Write("start logs cleaner...")
				_, err := os.Stat(l.logPath)
				if err != nil {
					l.Write(err.Error())
					continue
				}

				files, rErr := ioutil.ReadDir(l.logPath)
				if rErr != nil {
					l.Write(rErr.Error())
					continue
				}

				for _, f := range files {
					if !f.IsDir() {
						err := os.Remove(filepath.Join(l.logPath, f.Name()))
						if err != nil {
							l.Write(err.Error())
							continue
						}
						l.Write("remove log file success:", filepath.Join(l.logPath, f.Name()))
					}
				}
			}
		}
	}()
}
