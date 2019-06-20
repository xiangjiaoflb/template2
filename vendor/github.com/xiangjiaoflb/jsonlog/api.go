package jsonlog

import (
	"context"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

//NewJSONLog 创建日志库
func NewJSONLog(filepath, confpath string, arg ...interface{}) *zerolog.Logger {
	return newJSONLog(filepath, confpath, arg...)
}

//LogClose 关闭日志库
func LogClose(sg os.Signal) {
	logClose(sg)
}

// Debug ..
func Debug(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Debug())
}

// Info ..
func Info(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Info())
}

// Warn ..
func Warn(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Warn())
}

// Error ..
func Error(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Error())
}

// Fatal ..
func Fatal(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Fatal())
}

// Panic ..
func Panic(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Panic())
}

// Log ..
func Log(Logger *zerolog.Logger) *zerolog.Event {
	return fileAndLine(2, Logger.Log())
}

// Ctx ..
func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

//RequestLog http请求的日志 返回值, body里的数据
func RequestLog(r *http.Request, flog *zerolog.Event) (bodybuf []byte) {
	return requestLog(r, flog)
}

//SendJSON 发送json数据给前端
//flog 可以为nil,用来记录日志的
//data 为数据
//status 默认由400 和 200 两个
func SendJSON(flog *zerolog.Event, w http.ResponseWriter, err error, data interface{}, status int) error {
	return sendJSON(flog, w, err, data, status)
}
