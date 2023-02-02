package web

import (
	"log"
	"net"
	"net/http"
)

// 确保HttpServer实现Server接口
var _ Server = &HttpServer{}

type HandleFunc func(ctx *Context)

// 只包含核心接口
type Server interface {
	http.Handler

	Start(addr string) error

	//添加路由
	addRoute(method string, path string, handlerFunc HandleFunc, mdls ...Middleware)
	// 我们并不采取这种设计方案
	// addRoute(method string, path string, handlers... HandleFunc)

}

type HttpServer struct {
	// addr string 创建的时候传递，而不是 Start 接收。这个都是可以的
	router

	mdls []Middleware

	log func(msg string, args ...any)
}

//这种方法也可以，但是缺少拓展性
// func NewHTTPServerV1(mdls []Middleware) *HttpServer{
// 	return &HttpServer{
// 		router: newRouter(),
// 		mdls: mdls,
// 	}
// }

type HTTPServerOption func(server *HttpServer)

func NewHTTPServer(opts ...HTTPServerOption) *HttpServer {
	res := &HttpServer{
		router: newRouter(),
		log: func(msg string, args ...any) {
			log.Printf(msg, args...)
		},
	}

	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HttpServer) {
		server.mdls = mdls //直接覆盖
	}
}

// http.Handler接口中的方法  所有请求都经过这里
func (h *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 你的框架代码就在这里
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}

	//h.server(ctx)

	// 最后一个是这个
	root := h.server
	//从后往前  设置调用逻辑，把后一个的返回值参数，作为前一个next 组装链条
	for i := len(h.mdls) - 1; i >= 0; i-- {
		root = h.mdls[i](root)
	}

	// 第一个应该是回写响应的
	// 因为它在调用next之后才回写响应，
	// 所以实际上 flashResp 是最后一个步骤
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}
	root = m(root)
	//这里执行的时候，就是从前往后了
	root(ctx)
}

func (h *HttpServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	datalen, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || datalen != len(ctx.RespData) {
		h.log("回写相应失败: %v", err)
	}
}

func (h *HttpServer) server(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑
	//before route
	mi, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	//after route
	if !ok || mi.n.handler == nil {
		// ctx.Resp.WriteHeader(404)
		// ctx.Resp.Write([]byte("Not Found"))
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("Not Found")
		return
	}
	ctx.PathParams = mi.pathParams
	ctx.MatchedRoute = mi.n.route
	//before exec
	mi.n.handler(ctx)
	//after exec

}

// Start 启动服务器时，用户传入指定端口
// 这种就是编程接口
func (h *HttpServer) Start(addr string) error {
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 在这里，可以让用户注册所谓的 after start 回调
	// 比如说往你的 admin 注册一下自己这个实例
	// 在这里执行一些你业务所需的前置条件

	return http.Serve(ln, h)
}

func (h *HttpServer) Get(path string, handler HandleFunc) {
	h.addRoute(http.MethodGet, path, handler)
}

func (h *HttpServer) POST(path string, handler HandleFunc) {
	h.addRoute(http.MethodPost, path, handler)
}


func (h *HttpServer) Use(path string, handler HandleFunc, mdls ...Middleware) {
	h.addRoute(http.MethodGet, path, handler, mdls...)
}
