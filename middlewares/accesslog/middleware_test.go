package accesslog

import (
	"fmt"
	"net/http"
	"testing"
	"web"
)

func TestMiddlewareBuilderBuild(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl := builder.LogFunc(func(log string) {
		fmt.Println(log)
	}).Build()//返回一个Middleware
	server := web.NewHTTPServer(web.ServerWithMiddleware(mdl))
	server.POST("/a/b/*", func(ctx *web.Context) {
		fmt.Println("hello, it's me")
	})
	req, err := http.NewRequest(http.MethodPost, "/a/b/c", nil)
	if err != nil{
		t.Fatal(err)
	}
	server.ServeHTTP(nil, req)
}
