package errhdl

import "web"

type MiddlewareBuilder struct {
	//这种设计只能返回固定的值，不能动态渲染页面  key是状态码
	resp map[int][]byte
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		resp: make(map[int][]byte),
	}
}

func (m *MiddlewareBuilder) AddCode(status int, data []byte) *MiddlewareBuilder {
	m.resp[status] = data
	return m
}

func (m MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			next(ctx)
			if resp, ok := m.resp[ctx.RespStatusCode]; ok {
				//篡改返回结果
				ctx.RespData = resp
			}
		}
	}
}
