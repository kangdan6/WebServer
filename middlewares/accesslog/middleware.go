package accesslog

import (
	"encoding/json"
	"web"
)

type MiddlewareBuilder struct {
	logFunc func(log string)
}

// build 模式
func (m *MiddlewareBuilder) LogFunc(fn func(log string)) *MiddlewareBuilder  {
	m.logFunc = fn
	return m
}

func (m MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			//要记录请求
			defer func() {
				access := accessLog{
					Host: ctx.Req.Host,
					Route: ctx.MatchedRoute,
					Path: ctx.Req.URL.Path,
					HttpMethod: ctx.Req.Method,
				}
				data, _ := json.Marshal(access)
				m.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

type accessLog struct {
	Host       string `json:"host"`
	Route      string `json:"route"`
	Path       string `json:"path"`
	HttpMethod string `json:"http_method"`
}
