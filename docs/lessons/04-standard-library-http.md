# 第 4 课：标准库 HTTP 服务

## 学习目标

完成这一课后，你应该能够：

- 使用 `net/http` 编写最小 HTTP API。
- 理解 handler、request、response 的职责。
- 使用 `encoding/json` 处理 JSON 请求和响应。
- 为 handler 编写 `httptest` 单元测试。
- 通过 `curl` 验证本地 HTTP 服务。

## 前端架构师视角

前端通常从“页面路由”理解路由：URL 对应页面、组件和状态。后端 HTTP handler 的关注点不同：它接收请求、解析输入、调用业务逻辑、返回状态码和响应体。

你可能习惯在前端 API client 里假设响应总是 JSON，但后端必须显式设置状态码、响应头和错误格式。不要把“能在浏览器里看到结果”当成后端测试。Go 标准库提供 `httptest`，可以在不启动真实端口的情况下测试 handler。

本课仍然不引入第三方路由。先掌握标准库的基本模型，后面引入 `chi` 时你会更清楚它解决了什么问题。

## 核心概念

### http.Handler

`http.Handler` 是标准库的核心接口：

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

只要一个类型实现了 `ServeHTTP`，它就可以处理 HTTP 请求。

### http.HandlerFunc

普通函数可以通过 `http.HandlerFunc` 适配成 handler：

```go
func healthHandler(w http.ResponseWriter, r *http.Request)
```

### ServeMux

`http.ServeMux` 是标准库自带的请求路由器。本课用 Go 1.22+ 支持的 method pattern：

```go
mux.HandleFunc("GET /healthz", healthHandler)
```

这样可以把 HTTP 方法和路径一起写进路由声明。

### JSON 编解码

请求体使用 `json.NewDecoder(r.Body).Decode(&req)` 解析。响应体使用 `json.NewEncoder(w).Encode(value)` 输出。

后端要显式处理解码失败，不能默认相信客户端传来的 JSON。

## 代码结构

```text
examples/lesson04-standard-library-http/
  go.mod
  README.md
  main.go
  server.go
  server_test.go
```

- `main.go`：启动 HTTP 服务。
- `server.go`：定义路由、handler 和 JSON 辅助函数。
- `server_test.go`：使用 `httptest` 测试 HTTP 行为。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson04-standard-library-http
```

运行服务：

```bash
go run .
```

打开另一个终端验证健康检查：

```bash
curl -i http://localhost:8080/healthz
```

验证 JSON API：

```bash
curl -i -X POST http://localhost:8080/greetings \
  -H 'Content-Type: application/json' \
  -d '{"name":"Frontend Architect"}'
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

路由集中在 `NewHandler`：

```go
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("POST /greetings", greetingsHandler)
	return mux
}
```

`main` 不直接写业务逻辑，只启动服务：

```go
func main() {
	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, NewHandler()))
}
```

JSON 响应通过辅助函数统一处理：

```go
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
```

这比在每个 handler 里重复设置 header 和 encoder 更容易维护。

测试不需要真实监听端口：

```go
req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
rec := httptest.NewRecorder()

NewHandler().ServeHTTP(rec, req)
```

`httptest.NewRecorder` 会记录状态码、响应头和响应体，适合验证 HTTP 行为。

## 练习

1. 新增 `GET /version`，返回当前 demo 名称和版本。
2. 给 `/greetings` 增加空名称校验，返回 `400 Bad Request`。
3. 为错误响应设计统一结构，例如 `{"error":"name is required"}`。

## 常见坑

- 不要忘记设置 `Content-Type: application/json`。
- JSON 解码失败时要返回明确的 `400`。
- handler 里不要直接 `panic`，HTTP 服务需要稳定地返回错误响应。
- 不要只用 `curl` 手动验证，handler 行为应该有自动化测试。
- `main` 应该保持薄，方便把 handler 拿出来测试。

## 下一课

下一课会讲 `context.Context`、请求超时和取消。你会看到一个请求从进入服务到被取消之间，后端如何控制资源生命周期。
