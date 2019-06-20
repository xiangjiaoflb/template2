package jsonlog

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

//回复客户端的json
type httpReturn struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

//sendJSON 发送json数据给前端
//flog 可以为nil ，用来记录日志的
//data 为数据
//status 默认由400 和 200 两个
func sendJSON(flog *zerolog.Event, w http.ResponseWriter, err error, data interface{}, status int) error {
	var hr httpReturn
	if status == 0 {
		if err != nil {
			hr.Status = 400
			hr.Msg = err.Error()
		} else {
			hr.Status = 200
			hr.Msg = "success"
		}
	} else {
		hr.Status = status
		if err != nil {
			hr.Msg = err.Error()
		}
	}
	hr.Data = data
	buf, err := json.Marshal(hr)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	if len(buf) != 0 && flog != nil {
		if json.Valid(buf) {
			flog.RawJSON("Response", buf)
		}
	}
	return nil
}

//requestLog http请求的日志
func requestLog(r *http.Request, flog *zerolog.Event) []byte {
	//打印请求信息
	flog.Str("IP", getIP(getRealRemoteAddr(r))).
		Str("Method", r.Method).
		Str("URL", r.URL.String()).
		Str("Referer", r.Referer()).
		Str("UserAgent", r.UserAgent())

	//打印 Header
	headerbuf, err := json.Marshal(r.Header)
	if err != nil {
		flog.Str("headererror", err.Error())
	}
	if len(headerbuf) != 0 {
		if json.Valid(headerbuf) {
			flog.RawJSON("Header", headerbuf)
		}
	}

	//打印 body 内容
	bodybuf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		flog.Str("bodyerror", err.Error())
	}
	if len(bodybuf) != 0 {
		if json.Valid(bodybuf) {
			var jsonmap interface{}
			err = json.Unmarshal(bodybuf, &jsonmap)
			if err != nil {
				flog.Str("Unmarshalerror", err.Error())
			}
			bodybuf, err = json.Marshal(jsonmap)
			if err != nil {
				flog.Str("Marshalerror", err.Error())
			}
			flog.RawJSON("Request", bodybuf)
		} else {
			flog.Str("Request", strings.Replace(string(bodybuf), "\n", "换行符", -1))
		}
	}
	return bodybuf
}
