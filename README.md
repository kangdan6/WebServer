# 实现一个简单的WebServer
通过封装net.http,实现自己的HttpServer， 实现了http.Handler接口，处理用户业务逻辑，分为route模块和server模块，router负责路由添加、查找的功能， server通过类型嵌入route, 负责封装用户请求信息，查找路由，并且执行命中的业务逻辑！


## 实现一个路由树
route中为每个method构建一颗树，支持静态路由匹配、通配符匹配、正则匹配和参数路由匹配！
#### 使用说明

1.  通配符匹配中不支持路由回溯，比如同时注册 /a/* 和 /a/b/c 查找/a/b/d时不会匹配/a/*
2.  不允许同时注册通配符路由和参数路由
3.  同时注册/user/*和/user/*/home时，以/home开头的路由，按最长的规则匹配/user/*/home，否则匹配到/user/*,这种情况/user/123/home/456匹配到/user/*

#### Benchmark测试
1. 输出cpu性能文件

go test -bench=. -benchmem -cpuprofile=cpu.out

2. 输出mem内存性能文件

go test -bench=. -benchmem -memprofile=mem.out

3. 生成的CPU、内存文件可以通过go tool pprof [file]进行查看， 例如：go tool pprof cpu.out，然后在pprof中通过list [Benchamark方法], 例如：list BenchmarkStaticRouter  查看CPU、内存的耗时情况


下面执行命令go test -bench=. -benchmem (使用-v可输出详细信息)，收到测试报告如下：
```
geektime-homework % go test -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: web
BenchmarkStaticRouter_test-8     4551340               256.4 ns/op           160 B/op          9 allocs/op
BenchmarkAnyRouter_test-8        4089038               284.9 ns/op           176 B/op          6 allocs/op
BenchmarkParamRouter_test-8      1391556               899.2 ns/op          1984 B/op         20 allocs/op
BenchmarkRegRouter_test-8        1634415               716.8 ns/op           897 B/op         15 allocs/op
PASS
ok      web     7.451s

```
通过分析可得出如下结论：

1. 参数路由匹配最耗时，静态路由匹配最快

2. 静态路由的qps > 通配符的qps > 正则匹配的qps > 参数路由的qps

3. 静态路由的内存使用 < 通配符的内存使用 < 正则匹配的内存使用 < 参数路由的内存使用

## 可路由的MiddleWare设计

#### 使用说明

##### 如何使用grafana

1) docker-compose.yaml中添加服务：
```
prometheus:
    container_name: prometheus
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

grafana:
    container_name: grafana
    image: grafana/grafana:latest
    restart: on-failure
    volumes:
      - /etc/localtime:/etc/localtime
      - ./data/grafana:/var/lib/grafana
    ports:
      - "3000:3000"
```

2) 网页访问 http://localhost:3000，输入默认用户名密码：admin/admin

3) 配置数据源Data sources，添加 Prometheus 数据源, 默认Name是Prometheus-1， Prometheus URl：http://prometheus:9090, 注意这里不能写localhost 或者 127.0.0.1， 在grafana里面用localhost肯定是访问不了Prometheus， 要使用服务名或者本机ip。
