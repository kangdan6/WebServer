// +build e2e

package accesslog

import (
	"fmt"
	"testing"
	"web"
)

func TestMiddlewareBuilderBuilderE2E(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl := builder.LogFunc(func(log string) {
		fmt.Println(log)
	}).Build()
	server := web.NewHTTPServer(web.ServerWithMiddleware(mdl))
	server.Get("/a/b/*", func(ctx *web.Context) {
		ctx.Resp.Write([]byte("hello, it's me"))
	})

	server.Start(":8081")
}
