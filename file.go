package _1mylogger

import (
	"fmt"
	"os"
	"path"
	"time"
)

// 往文件中输出日志信息

type FileLogger struct {
	maxFileSize int64
	filePath    string
	fileName    string
	level       LogLevel

	logFile *os.File
	errFile *os.File

	logChan chan *LogMsg // 用来实现异步写日志的功能
}

// LogMsg 针对异步写日志功能增加的结构体
type LogMsg struct {
	Level     LogLevel
	msg       string
	funcName  string
	fileName  string
	timeStamp string
	line      int
}

var (
	// MaxSize 异步写日志chan的大小
	MaxSize = 50000
)

func NewFileLogger(lv string, filePath string, fileName string, maxSize int64) *FileLogger {
	level0 := parseStrToInt(lv)
	fileLogObj := &FileLogger{
		maxFileSize: maxSize,
		filePath:    filePath,
		fileName:    fileName,
		level:       level0,
		logChan:     make(chan *LogMsg, MaxSize),
	}
	err := fileLogObj.initFile()
	if err != nil {
		panic(err)
	}
	return fileLogObj
}

func (f *FileLogger) initFile() error {
	logPath := path.Join(f.filePath, f.fileName+".log")
	logFile, err1 := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err1 != nil {
		fmt.Printf("open log file failed, err: %v", err1)
		return err1
	}

	errPath := path.Join(f.filePath, f.fileName+"err.log")
	errFile, err2 := os.OpenFile(errPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err2 != nil {
		fmt.Printf("open err log file failed, err: %v", err2)
		return err2
	}

	f.logFile = logFile
	f.errFile = errFile
	// 开启一个后台的goroutine去往文件里写日志
	go f.writeLogBackground() // 开启一个goroutine
	// 开启五个goroutine【不可用】。但是需要注意如果开设多个goroutine，有可能多个goroutine同时操作一个文件句柄，是不被允许的操作
	//for i := 0; i < 5; i++ {
	//	go f.writeLogBackground()
	//}
	return nil
}

func (f *FileLogger) Close() {
	f.logFile.Close()
	f.errFile.Close()
}

func (f *FileLogger) checkFileSize(fileObj *os.File) bool {
	info, err1 := fileObj.Stat()
	if err1 != nil {
		fmt.Printf("file error, err1")
		return false
	}
	return info.Size() >= f.maxFileSize
}

func (f *FileLogger) splitFile(fileObj *os.File) (*os.File, error) {
	// 1 拿到文件信息
	info, err1 := fileObj.Stat()
	if err1 != nil {
		fmt.Printf("file error, err1")
		return nil, err1
	}
	now := time.Now().Format("2006010215040506")

	logFullPath := path.Join(f.filePath, info.Name())
	newLogFullPath := fmt.Sprintf("%s.bak%s", logFullPath, now)
	// 2 把老文件关闭
	fileObj.Close()
	// 4 完成重命名替换文件
	os.Rename(logFullPath, newLogFullPath)
	// 3 通过文件信息打开新文件
	newLog, err2 := os.OpenFile(logFullPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err2 != nil {
		panic(err2)
	}

	return newLog, nil
}

// writeLogBackground 真正异步写日志的方式。应该在打印文件的时候在后台开启一个goroutine来异步写日志，在initFile中
func (f *FileLogger) writeLogBackground() {
	for {
		if f.checkFileSize(f.logFile) {
			newFile, err1 := f.splitFile(f.logFile)
			if err1 != nil {
				return
			}
			f.logFile = newFile
		}
		select {
		case logTmp := <-f.logChan: // 如果取不到日志，会陷入阻塞状态。加一个select，如果去不到日志，走default分支
			logInfo := fmt.Sprintf("[%s] [%s] [%s :%s :%d] %s\n",
				logTmp.timeStamp,
				unParseLogLevel(logTmp.Level),
				logTmp.fileName,
				logTmp.funcName,
				logTmp.line,
				logTmp.msg)
			fmt.Fprintf(f.logFile, logInfo)
			if logTmp.Level >= ERROR {
				if f.checkFileSize(f.errFile) {
					newFile, err1 := f.splitFile(f.errFile)
					if err1 != nil {
						return
					}
					f.errFile = newFile
				}
				// 如果要记录的日志级别大于等于ERROR级别，我还要在err日志文件中再记录一遍
				fmt.Fprintf(f.errFile, logInfo)
			}
		default: // 取不到日志，先休息一秒
			time.Sleep(time.Second)
		}

	}
}

func (f *FileLogger) log(lv LogLevel, format string, etc ...interface{}) {
	if f.enable(lv) {
		msg := fmt.Sprintf(format, etc...)
		now := time.Now()
		funcName, fileName, lineNo := getInfo(3)
		//levelStr := unParseLogLevel(lv)
		// 先把日志发送到通道中，后台干活的goroutine从通道取出再写入到文件里面
		logTmp := &LogMsg{
			Level:     lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			timeStamp: now.Format("2006-01-02 15:04:05"),
			line:      lineNo,
		}
		//f.logChan <- logTmp
		// 正常期望来说，log函数的实现应该马上执行。但可能会阻塞，如果chan写满了，再写入数据仍然会阻塞，可以使用select来进行区分
		select {
		case f.logChan <- logTmp:
		default:
			// 把日志丢掉了保证不出现阻塞。这是极端情况，正常情况下一直有goroutine从logChan中读取日志写日志
		}
	}
}

func (f *FileLogger) enable(lv LogLevel) bool {
	return f.level <= lv
}

func (f *FileLogger) Debug(msg string, etc ...interface{}) {
	f.log(DEBUG, msg, etc...)
}

func (f *FileLogger) Trace(msg string, etc ...interface{}) {
	f.log(TRACE, msg, etc...)
}

func (f *FileLogger) Info(msg string, etc ...interface{}) {
	f.log(INFO, msg, etc...)
}

func (f *FileLogger) Warning(msg string, etc ...interface{}) {
	f.log(WARNING, msg, etc...)
}

func (f *FileLogger) Error(msg string, etc ...interface{}) {
	f.log(ERROR, msg, etc...)
}

func (f *FileLogger) Fatal(msg string, etc ...interface{}) {
	f.log(FATAL, msg, etc...)
}
