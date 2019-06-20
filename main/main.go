package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"template/log"
	"template/route"
	"template/src/database"

	"github.com/xiangjiaoflb/httpframe"
	"github.com/xiangjiaoflb/jsonlog"
)

//版本号等
var (
	// VERSION 版本
	VERSION = "1.0.1"

	// BUILDTIME 编译时间
	BUILDTIME = ""

	// GOVERSION go 版本
	GOVERSION = ""

	// GITHASH 代码hash值
	GITHASH = ""
)

func main() {
	port := 0
	flag.IntVar(&port, "port", 9000, "server port")
	flag.Parse()

	//程序启动打印日志
	jsonlog.Info(log.RunLog).
		Str("VERSION", VERSION).
		Str("BUILDTIME", BUILDTIME).
		Str("GOVERSION", GOVERSION).
		Str("GITHASH", GITHASH).Msg("begin run!")

	servermux := http.NewServeMux()

	//注册路由 不走中间件
	httpframe.RegisterHandle(servermux, nil, &route.Api{})

	//其他路由走中间件
	servermux.HandleFunc("/", httpframe.NewMiddleware(append([]httpframe.HandlerFunc{},
		func(ctx *httpframe.Context) { http.FileServer(http.Dir(".")).ServeHTTP(ctx.W, ctx.R) })).HandleFunc)

	//查看版本号
	servermux.HandleFunc("/version", httpframe.NewMiddleware(append([]httpframe.HandlerFunc{},
		func(ctx *httpframe.Context) {
			jsonlog.SendJSON(nil, ctx.W, nil, map[string]string{
				"VERSION":   VERSION,
				"BUILDTIME": BUILDTIME,
				"GOVERSION": GOVERSION,
				"GITHASH":   GITHASH,
			}, 200)
		})).HandleFunc)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), servermux)
	if err != nil {
		jsonlog.Error(log.RunLog).Err(err).Msg("")
	}
}

//创建数据库连接和表
func init() {
	db, err := database.Open("root:root@tcp(192.168.216.129:3306)/mydata?charset=utf8&parseTime=True")
	if err != nil {
		jsonlog.Error(log.RunLog).Err(err).Msg("")
		os.Exit(-1)
	}
	if DEBUG {
		db.LogMode(true)
	}
}

var (
	//DEBUG 调试开关
	DEBUG = false
)
