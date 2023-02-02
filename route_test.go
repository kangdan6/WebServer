package web

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// tdd 测试驱动开发
func TestHttpServer_addRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则路由
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}

	mockHandler := func(ctx *Context) {}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path:    "/",
				typ:     nodeTypeStatic,
				handler: mockHandler,
				children: map[string]*node{
					"user": {path: "user", typ: nodeTypeStatic, handler: mockHandler, children: map[string]*node{
						"home": {path: "home", typ: nodeTypeStatic, handler: mockHandler}}},
					"order": {path: "order", typ: nodeTypeStatic, children: map[string]*node{
						"detail": {path: "detail", typ: nodeTypeStatic, handler: mockHandler},
					}, starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler}},
					"param": {
						path: "param",
						typ:  nodeTypeStatic,
						paramChild: &node{
							path:      ":id",
							handler:   mockHandler,
							paramName: "id",
							typ:       nodeTypeParam,
							children: map[string]*node{
								"detail": &node{path: "detail", typ: nodeTypeStatic, handler: mockHandler}},
							starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler},
						}},
				},
				starChild: &node{
					path:      "*",
					typ:       nodeTypeAny,
					handler:   mockHandler,
					starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler},
					children: map[string]*node{
						"abc": &node{path: "abc", typ: nodeTypeStatic, handler: mockHandler,
							starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler}},
					},
				},
			},
			http.MethodPost: {path: "/", typ: nodeTypeStatic, children: map[string]*node{
				"order": {path: "order", typ: nodeTypeStatic, children: map[string]*node{
					"create": {path: "create", typ: nodeTypeStatic, handler: mockHandler},
				}},
				"login": {path: "login", handler: mockHandler, typ: nodeTypeStatic},
			}},
			http.MethodDelete: {
				path: "/",
				typ:  nodeTypeStatic,
				children: map[string]*node{"reg": &node{path: "reg", typ: nodeTypeStatic,
					regChild: &node{path: ":id(.*)", typ: nodeTypeReg, handler: mockHandler, paramName: "id"}}},
				regChild: &node{path: ":name(^.+$)", typ: nodeTypeReg, paramName: "name",
					children: map[string]*node{"abc": &node{path: "abc", typ: nodeTypeStatic, handler: mockHandler}}},
			},
		},
	}

	//判断wantRouter和r是否相等
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	// 非法用例
	r = newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	// 同时注册通配符路由和参数路由
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由、正则路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由，已有路径参数路由。不允许同时注册通配符路由、正则路由和参数路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})

}

func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 里面没有方法 %s 的路由树", k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if y.path != n.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	// 新增类型
	if y.typ != n.typ {
		return fmt.Sprintf("%s 节点 typ 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	if y.paramName != n.paramName {
		return fmt.Sprintf("%s 节点参数名字不相等 x %s, y %s", n.path, n.paramName, y.paramName), false
	}

	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}

	if len(n.children) == 0 {
		return "", true
	}

	//判断参数路径节点是否相等
	if n.paramChild != nil {
		str, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}

	//判断正则匹配节点是否相等
	if n.regChild != nil {
		str, ok := n.regChild.equal(y.regChild)
		if !ok {
			return fmt.Sprintf("%s 路径参数节点不匹配 %s", n.path, str), false
		}
	}
	//判断通配符节点是否相等
	if n.starChild != nil {
		str, ok := n.starChild.equal(y.starChild)
		if !ok {
			return fmt.Sprintf("%s 通配符节点不匹配 %s", n.path, str), false
		}
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},

		//必做题
		//通配符匹配：现在我们的通配符匹配只能匹配一段，现在要你修改为，如果通配符出现在路由的末尾，
		//例如 /a/b/*, 那么它能够匹配到后面多段路由，例如 /a/b/c/d/e/f，而目前我们只支持 /a/b/c
		{
			method: http.MethodDelete,
			path:   "/user/*",
		},
		{
			method: http.MethodDelete,
			path:   "/user/*/home",
		},
		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}

	r := newRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
	}{
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "/",
					typ:     nodeTypeStatic,
					handler: mockHandler,
					children: map[string]*node{
						"user": &node{
							path:      "user",
							typ:       nodeTypeStatic,
							handler:   mockHandler,
							starChild: &node{path: "*", children: map[string]*node{"home": &node{path: "home", handler: mockHandler}}},
						},
						"param": &node{
							path: "param",
							typ:  nodeTypeStatic,
							paramChild: &node{path: ":id", typ: nodeTypeStatic, handler: mockHandler,
								children: map[string]*node{
									"detail": &node{path: "detail", typ: nodeTypeStatic, handler: mockHandler},
								},
								starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler},
							},
						},
					},
				},
			},
		},
		{
			name:   "method not found",
			method: http.MethodConnect,
			path:   "/",
			found:  false,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/aaa",
			found:  false,
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			mi: &matchInfo{
				n: &node{
					typ:     nodeTypeStatic,
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path: "order",
					typ:  nodeTypeStatic,
					children: map[string]*node{
						"create": &node{
							path:    "create",
							typ:     nodeTypeStatic,
							handler: mockHandler,
						},
					},
					starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler},
				},
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "create",
					typ:     nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},

		// 通配符匹配
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					typ:     nodeTypeAny,
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "匹配/order/*",
			method: http.MethodPost,
			path:   "/order/delete/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					typ:     nodeTypeAny,
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:   "star in middle",
			method: http.MethodGet,
			path:   "/user/Tom/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					typ:     nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},
		//参数路由
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
			found:  true,
			mi: &matchInfo{
				n: &node{path: ":id", paramName: "id", typ: nodeTypeParam, handler: mockHandler, children: map[string]*node{
					"detail": &node{path: "detail", typ: nodeTypeStatic, handler: mockHandler},
				}, starChild: &node{path: "*", typ: nodeTypeAny, handler: mockHandler}},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					typ:     nodeTypeAny,
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},

		{
			// 命中 /param/:id/detail
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/detail",
			found:  true,
			mi: &matchInfo{
				n: &node{
					typ:     nodeTypeStatic,
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},

		//必做题
		//通配符匹配：现在我们的通配符匹配只能匹配一段，现在要你修改为，如果通配符出现在路由的末尾，
		//例如 /a/b/*, 那么它能够匹配到后面多段路由，例如 /a/b/c/d/e/f，而目前我们只支持 /a/b/c
		{
			// 命中 /user/*
			name:   "* 可以匹配多段路由",
			method: http.MethodDelete,
			path:   "/user/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					typ:     nodeTypeAny,
					handler: mockHandler,
					children: map[string]*node{
						"home": &node{path: "home", typ: nodeTypeStatic, handler: mockHandler},
					},
				},
			},
		},
		{
			// 命中 /user/*
			name:   "/user/456/home/abc 命中 /user/*",
			method: http.MethodDelete,
			path:   "/user/456/home/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					typ:     nodeTypeAny,
					handler: mockHandler,
					children: map[string]*node{
						"home": &node{path: "home", typ: nodeTypeStatic, handler: mockHandler},
					},
				},
			},
		},

		{
			// 命中 /user/*/home
			name:   "/user/123/home 命中 /user/*/home",
			method: http.MethodDelete,
			path:   "/user/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					typ:     nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},

		{
			// /user/123/345/home 未命中 /user/*/home 命中 /user/*
			name:   "/user/123/345/home 未命中 /user/*/home 命中 /user/*",
			method: http.MethodDelete,
			path:   "/user/123/345/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					typ:     nodeTypeAny,
					handler: mockHandler,
					children: map[string]*node{
						"home": &node{path: "home", typ: nodeTypeStatic, handler: mockHandler},
					},
				},
			},
		},

		// 正则
		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				pathParams: map[string]string{"id": "123"},
				n: &node{
					path:      ":id(.*)",
					paramName: "id",
					typ:       nodeTypeReg,
					handler:   mockHandler,
				},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:   "命中 /:id([0-9]+)/home",
			method: http.MethodDelete,
			path:   "/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:      "home",
					typ:       nodeTypeStatic,
					handler:   mockHandler,
					paramName: "",
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 未命中 /:id([0-9]+)/home
			name:   "not :id([0-9]+)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(mi.n.handler)
			assert.Equal(t, wantVal, nVal)
			msg, b := mi.n.equal(tc.mi.n)
			assert.True(t, b, msg)
		})
	}
}

// Test_router_findRoute_middleware 测试middleware
func Test_router_findRoute_middleware(t *testing.T) {
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
		{
			method: http.MethodPost,
			path:   "/a/b/c",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('b'), mdlBuilder('c')},
		},
		{
			method: http.MethodDelete,
			path:   "/*",
			mdls:   []Middleware{mdlBuilder('*')},
		},
		{
			method: http.MethodDelete,
			path:   "/",
			mdls:   []Middleware{mdlBuilder('/')},
		},
		//参数路由
		{
			method: http.MethodOptions,
			path:   "/a/:id",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder(':')},
		},
		//正则路由
		{
			method: http.MethodOptions,
			path:   "/a/:id(.*)",
			mdls:   []Middleware{mdlBuilder('a'), mdlBuilder('r')},
		},
	}

	r := newRouter()
	mockHandler := func(ctx *Context) {}
	for _, mdlRoute := range mdlRoutes {
		r.addRoute(mdlRoute.method, mdlRoute.path, mockHandler, mdlRoute.mdls...)
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
			path:     "/a/c",
			wantResp: "a*",
		},
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
		{
			name:     "root",
			method:   http.MethodDelete,
			path:     "/",
			wantResp: "/",
		},
		{
			name:     "root star",
			method:   http.MethodDelete,
			path:     "/a",
			wantResp: "/*",
		},
		//参数路由
		{
			name:   "param, match",
			method: http.MethodOptions,
			path:   "/a/123/c",
			wantResp: "ara:",
		},


		//正则路由   先匹配正则路由 /a/:id(.*) 再匹配参数符路由/a/:id
		{
			name:   "reg, match",
			method: http.MethodOptions,
			path:   "/a/123/c",
			wantResp: "ara:",
		},

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, _ := r.findRoute(tc.method, tc.path)
			mdls := mi.mdls
			var root HandleFunc = func(ctx *Context) {
				//fmt.Println(tc.wantResp, string(ctx.RespData))
				assert.Equal(t, tc.wantResp, string(ctx.RespData))
			}
			for i := len(mdls) - 1; i >= 0; i-- {
				root = mdls[i](root)
			}

			//开始调度
			root(&Context{
				RespData: make([]byte, 0, len(tc.wantResp)),
			})

		})
	}
}
