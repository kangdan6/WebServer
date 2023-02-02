package errhdl

import (
	"net/http"
	"testing"
	"web"
)

func TestMiddleware_Builder(t *testing.T) {
	builder := NewMiddlewareBuilder()
	builder = builder.AddCode(http.StatusNotFound, []byte(`
	<html>
		<body>
			<h1>哈哈哈, 走失了</h1>
		</body>
	</html>
	`)).AddCode(http.StatusBadRequest, []byte(`
	<html>
		<body>
			<h1>请求不对</h1>
		</body>
	</html>
	`))
	server := web.NewHTTPServer(web.ServerWithMiddleware(builder.Build()))
	server.Start(":8081")
}
