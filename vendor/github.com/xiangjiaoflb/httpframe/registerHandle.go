package httpframe

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

//Context http 上下文
type Context struct {
	W    http.ResponseWriter
	R    *http.Request
	next bool

	Data interface{} //上下文的信息传递
}

// Next 执行下一个handle
// 执行Next后上一个handle应该立即返回
func (c *Context) Next() {
	c.next = true
}

//HandlerFunc ...
type HandlerFunc func(*Context)

// RegisterHandle 注册自动路由
func RegisterHandle(servermux *http.ServeMux, middleware []HandlerFunc, classes ...interface{}) {
	fixName := func(name string) string {
		r := []rune(name)
		a := map[rune]rune{'A': 'a', 'B': 'b', 'C': 'c', 'D': 'd', 'E': 'e', 'F': 'f', 'G': 'g', 'H': 'h', 'I': 'i', 'J': 'j', 'K': 'k', 'L': 'l', 'M': 'm', 'N': 'n', 'O': 'o', 'P': 'p', 'Q': 'q', 'R': 'r', 'S': 's', 'T': 't', 'U': 'u', 'V': 'v', 'W': 'w', 'X': 'x', 'Y': 'y', 'Z': 'z'}
		b := map[string]string{"A": "_a", "B": "_b", "C": "_c", "D": "_d", "E": "_e", "F": "_f", "G": "_g", "H": "_h", "I": "_i", "J": "_j", "K": "_k", "L": "_l", "M": "_m", "N": "_n", "O": "_o", "P": "_p", "Q": "_q", "R": "_r", "S": "_s", "T": "_t", "U": "_u", "V": "_v", "W": "_w", "X": "_x", "Y": "_y", "Z": "_z"}

		// 首字母小写
		if v, ok := a[r[0]]; ok {
			r[0] = v
		}

		// 除首字母外，其它大写字母替换成下划线加小写
		s := string(r)
		for k, v := range b {
			s = strings.Replace(s, k, v, -1)
		}
		return s
	}

	for _, c := range classes {
		name := reflect.TypeOf(c).Elem().Name()
		if strings.HasPrefix(name, "Controller") {
			name = name[len("Controller"):]
		}
		name = "/" + fixName(name)
		for i := 0; i < reflect.TypeOf(c).NumMethod(); i++ {
			method := "/" + fixName(reflect.TypeOf(c).Method(i).Name)
			path := name + method

			//注册路由
			servermux.HandleFunc(path, NewMiddleware(append(middleware, func(v reflect.Value) HandlerFunc {
				return func(c *Context) { v.Call([]reflect.Value{reflect.ValueOf(c)}) }
			}(reflect.ValueOf(c).Method(i)))).HandleFunc)

			//打印监听的路由
			fmt.Printf("监听路由:%s\n", path)

		}
	}
}

//Mymiddleware 中间件支持
type Mymiddleware struct {
	handlearr []HandlerFunc
}

//NewMiddleware 创建支持中间件的实例
func NewMiddleware(handlearr []HandlerFunc) *Mymiddleware {
	return &Mymiddleware{handlearr: handlearr}
}

//HandleFunc ...
func (m *Mymiddleware) HandleFunc(w http.ResponseWriter, r *http.Request) {
	ct := Context{
		W:    w,
		R:    r,
		next: false, //默认不运行下一个handle
	}
	for _, v := range m.handlearr {
		v(&ct)
		if ct.next {
			continue
		}
		break
	}
}
