//go:build e2e
// +build e2e

package prometheus

import (
	"math/rand"
	"net/http"
	"testing"
	"time"
	"web"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{
		Namespace: "geektime_homework",
		Name:      "http_response",
		Subsystem: "web",
	}
	server := web.NewHTTPServer(web.ServerWithMiddleware(builder.Build()))

	//测试用例
	server.Get("/user", func(ctx *web.Context) {
		duration := rand.Intn(1000) + 1
		time.Sleep(time.Duration(duration) * time.Millisecond)
		ctx.RespJson(202, User{Name: "Tom"})
	})

	// 启动prometheus
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()

	server.Start(":8081")
}

type User struct {
	Name string
}
