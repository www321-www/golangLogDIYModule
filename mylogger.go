package _1mylogger

import (
	"runtime"
	"strings"
)

/* 需求分析
	1.支持往不同的地方输出日志
	2.日志分级别
		debug
		trace
		info
		warning
		error
		fatal
	3.日志要支持开关控制，比如说开发的时候什么级别都能输出，但是上线之后只有info级别往下的才能输出
	4.完整的日志记录要包含有时间、行号、文件名、日志级别、日志信息
	5.日志文件要切割
		5.1 按文件大小切割。
			每次记录日志之前判断一下当前写的这个文件的大小
		5.2 按日期切割。
			在日志结构体中设置一个字段记录上一次切割的小时数
			在写日志之前检查一下当前时间的小时数和之前保存的是否一致，不一致就要切割

升级版：【针对写入文件的功能】
    原代码版本是串行的写日志内容(字符串)，如果出现要写入日志比较多的情况，可能会导致程序运行速度比较慢。
	现要求将同步的日志功能变成异步写日志功能

实现：
1 构造一个chan，通道里面最好不要放字符串，因为字符串太大了(字符串指针同理)。可以存放结构体的指针，需要设计一个LogMsg的结构体
2 结构体需要含有想要往文件中写的日志内容，方便从chan中取出时拼接字符串内容
3 chan的初始化，还可以从外部设置开设的通道的大小
*/

type LogLevel uint16

type Logger interface {
	Debug(format string, a ...interface{})
	Trace(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warning(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(format string, a ...interface{})
}

func BuildLogger(t string, level string) Logger {
	var logObj Logger
	switch t {
	case "file":
		logObj = NewFileLogger(level, ".", "log", 1*1024*1024)
	case "console":
		logObj = NewConsoleLogger(level)
	default:
		panic("wrong input type")
	}
	return logObj
}

const (
	UNKNOWN LogLevel = iota
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

func parseStrToInt(lv string) LogLevel {
	lv = strings.ToLower(lv)
	switch lv {
	case "debug":
		return DEBUG
	case "trace":
		return TRACE
	case "info":
		return INFO
	case "warning":
		return WARNING
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return UNKNOWN
	}
}

func unParseLogLevel(lv LogLevel) string {
	switch lv {
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func getInfo(skip int) (fileName, funcName string, lineNo int) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		return
	}
	idx := strings.LastIndex(file, "/")
	fileName = file[idx+1:]
	info := runtime.FuncForPC(pc)
	funcName = strings.Split(info.Name(), ".")[1]
	return fileName, funcName, lineNo
}
