package web

import (
	"net/http"
	"testing"
)

// ● 静态路由的Benchmark测试
func BenchmarkStaticRouter_middleware_test(b *testing.B) {
	var mdlBuilder = func(i byte) Middleware {
		return func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				ctx.RespData = append(ctx.RespData, i)
				next(ctx)
			}
		}
	}

	mdlRoutes := []struct {
		method string
		path   string
		mdls   []Middleware
	}{

		{
			method: http.MethodGet,
			path:   "/a/b",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b')},
		},
		{
			method: http.MethodDelete,
			path:   "/",
			mdls:   []Middleware{mdlBuilder('/')},
		},
	}

	r := newRouter()
	for _, mdlRoute := range mdlRoutes {
		r.addRoute(mdlRoute.method, mdlRoute.path, nil, mdlRoute.mdls...)
	}

	testCases := []struct {
		name   string
		method string
		path   string
		// 我们借助 ctx 里面的 RespData 字段来判断 middleware 有没有按照预期执行
		wantResp string
	}{
		{
			name:   "static, not match",
			method: http.MethodGet,
			path:   "/a",
		},
		{
			name:     "static, match",
			method:   http.MethodGet,
			path:     "/a/b",
			wantResp: "ab",
		},

		{
			name:     "root",
			method:   http.MethodDelete,
			path:     "/",
			wantResp: "/",
		},
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}

}

// ● 通配符匹配路由的Benchmark测试
func BenchmarkAnyRouter_middleware_test(b *testing.B) {
	var mdlBuilder = func(i byte) Middleware {
		return func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				ctx.RespData = append(ctx.RespData, i)
				next(ctx)
			}
		}
	}

	mdlRoutes := []struct {
		method string
		path   string
		mdls   []Middleware
	}{

		{
			method: http.MethodGet,
			path:   "/a/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*')},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/b/*",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('*')},
		},
		{
			method: http.MethodPost,
			path:   "/a/*/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('*'), mdlBuilder('c')},
		},
	}

	r := newRouter()
	for _, mdlRoute := range mdlRoutes {
		r.addRoute(mdlRoute.method, mdlRoute.path, nil, mdlRoute.mdls...)
	}

	testCases := []struct {
		name   string
		method string
		path   string
		// 我们借助 ctx 里面的 RespData 字段来判断 middleware 有没有按照预期执行
		wantResp string
	}{
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b",
			wantResp: "a*ab",
		},
		{
			name:     "static and star",
			method:   http.MethodGet,
			path:     "/a/b/c",
			wantResp: "a*abab*",
		},
		{
			name:     "abc",
			method:   http.MethodPost,
			path:     "/a/b/c",
			wantResp: "a*cab*abc",
		},
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}

}

