package log

import (
	"path"

	"github.com/xiangjiaoflb/jsonlog"
	"github.com/xiangjiaoflb/pathmanage"
)

var (
	runLogPath     = path.Join(pathmanage.GetLOGPATH("run"), "run.log")
	runLogConfPath = path.Join(pathmanage.GetCONFPATH("log"), "runlog")
	//RunLog 运行日志
	RunLog = jsonlog.NewJSONLog(runLogPath, runLogConfPath,
		map[string]interface{}{
			jsonlog.MaxAge:             "7d", //保留7天的日志
			jsonlog.LogLevel:           0,    //需要打印的最低的日志级别
			jsonlog.LoggingWriteOutput: true, //是否打印到标准输出
		})

	requestLogPath     = path.Join(pathmanage.GetLOGPATH("request"), "request.log")
	requestLogConfPath = path.Join(pathmanage.GetCONFPATH("log"), "requestlog")
	//RequestLog 请求日志
	RequestLog = jsonlog.NewJSONLog(requestLogPath, requestLogConfPath,
		map[string]interface{}{
			jsonlog.MaxAge:   "7d", //保留7天的日志
			jsonlog.LogLevel: 0,    //需要打印的最低的日志级别
		})
)
