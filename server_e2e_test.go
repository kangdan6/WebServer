// +build e2e

// go tag 是为了防止后面跑单元测试跑不起来
//集成测试在这里跑，因为要起端口
package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T){
	h := NewHTTPServer()

	h.addRoute(http.MethodGet, "/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello user!"))
	})

	//获取参数路由的 参数
	h.addRoute(http.MethodGet, "/user/:id", func(ctx *Context) {
		id, ok := ctx.PathParams["id"]
		if !ok{
			ctx.Resp.Write([]byte(fmt.Sprintf("id is empty")))
		}
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %s", id )))

	})


	//获取参数路由的 参数
	h.addRoute(http.MethodGet, "/a/b/*", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("命中路径/a/b/*" )))

	})

	h.Get("/form", func(ctx *Context) {
		ctx.Resp.Write([]byte("aaa"))
	})


	h.Get("/values/:id", func(ctx *Context) {
		id, err := ctx.PathValue1("id").AsInt64()
		if err != nil{
			ctx.Resp.WriteHeader(400)
			ctx.Resp.Write([]byte("id 输入不对"))
			return
		}
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %d", id)))
	})

	type User struct{
		Name string `json:"name"`
	}
	h.Get("/user/123", func(ctx *Context) {
		ctx.RespJson(http.StatusOK,User{
			Name: "Tom",
		})
	})


	//使用安全的context
	//h.Get("/user/456", func(ctx *Context) {
	//	safeCtx := SafeContext{
	//		Context: *ctx,
	//	}
	//	safeCtx.Context.RespJsonOK(User{
	//		Name: "Tom",
	//	})
	//})

	// 用法一 完全委托给 http 包
	//http.ListenAndServe("8081", h)


	h.Start(":8081")
}


//Context是线程不安全的，可以自己加装饰器实现一个安全的Context
//type SafeContext struct {
//	Context
//	mutex sync.RWMutex
//}
//
//func (c *SafeContext) RespJsonOK( val any)error {
//	c.mutex.Lock()
//	defer c.mutex.Unlock()
//	return c.Context.RespJsonOK(val)
//}
