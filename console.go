package _1mylogger

import (
	"fmt"
	"time"
)

// 往终端中输出日志信息

type ConsoleLogger struct {
	level LogLevel
}

func NewConsoleLogger(level string) *ConsoleLogger {
	lv := parseStrToInt(level)
	if lv == UNKNOWN {
		return nil
	}
	return &ConsoleLogger{
		level: lv,
	}
}

func (c *ConsoleLogger) log(lv LogLevel, msg string, etc ...interface{}) {
	if c.enable(DEBUG) {
		msg = fmt.Sprintf(msg, etc...)
		now := time.Now().Format("2006-01-02 15:04:05")
		fileName, funcName, lineNo := getInfo(3)
		levelStr := unParseLogLevel(lv)
		fmt.Printf("[%s] [%s] [%s :%s :%d] %s\n",
			now,
			levelStr,
			fileName,
			funcName,
			lineNo,
			msg)
	}
}

func (c *ConsoleLogger) enable(lv LogLevel) bool {
	return c.level <= lv
}

func (c *ConsoleLogger) Debug(msg string, etc ...interface{}) {
	c.log(DEBUG, msg, etc...)
}

func (c *ConsoleLogger) Trace(msg string, etc ...interface{}) {
	c.log(TRACE, msg, etc...)
}

func (c *ConsoleLogger) Info(msg string, etc ...interface{}) {
	c.log(INFO, msg, etc...)
}

func (c *ConsoleLogger) Warning(msg string, etc ...interface{}) {
	c.log(WARNING, msg, etc...)
}

func (c *ConsoleLogger) Error(msg string, etc ...interface{}) {
	c.log(ERROR, msg, etc...)
}

func (c *ConsoleLogger) Fatal(msg string, etc ...interface{}) {
	c.log(FATAL, msg, etc...)
}
