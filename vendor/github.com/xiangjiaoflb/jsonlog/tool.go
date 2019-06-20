package jsonlog

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
)

// fileAndLine 打印文件和对应的行数
func fileAndLine(calldepth int, e *zerolog.Event) *zerolog.Event {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	// short := file
	// for i := len(file) - 1; i > 0; i-- {
	// 	if file[i] == '/' {
	// 		short = file[i+1:]
	// 		break
	// 	}
	// }
	// file = short

	return e.Str("file", file).Int("line", line)
}

func getIP(ipAndPort string) string {
	return strings.Split(ipAndPort, ":")[0]
}

func getRealRemoteAddr(r *http.Request) string {
	if r.Header.Get("X-FORWARDED-FOR") != "" {
		xf := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ",")
		return strings.TrimSpace(xf[len(xf)-1]) + ":2345"
	}

	return r.RemoteAddr
}
