# 第 6 课：路由、中间件与 API 组织

## 学习目标

完成这一课后，你应该能够：

- 使用 `chi` 组织 HTTP 路由。
- 理解中间件的执行顺序。
- 编写一个简单请求日志中间件。
- 为 API 设计版本前缀和资源路径。
- 统一 JSON 响应和错误响应。

## 前端架构师视角

前端里的拦截器、插件、路由守卫和后端中间件有相似之处：它们都在核心业务逻辑之前或之后插入横切逻辑，例如日志、鉴权、错误恢复、请求 ID、埋点等。

区别在于，后端中间件处在真实请求链路上。它影响的不只是 UI 状态，还影响状态码、响应头、日志、超时、恢复策略和生产排障能力。

本课第一次引入第三方库 `chi`。选择它是因为它贴近 `net/http`，不会强迫你进入一个庞大框架。handler 仍然是普通的 `func(w http.ResponseWriter, r *http.Request)`。

## 核心概念

### Router

`chi.NewRouter()` 返回一个路由器。你可以按 HTTP 方法和路径注册 handler：

```go
r.Get("/healthz", healthHandler)
r.Get("/api/v1/tasks/{taskID}", getTaskHandler)
```

路径参数通过 `chi.URLParam(r, "taskID")` 获取。

### Middleware

中间件本质上是函数：

```go
func(next http.Handler) http.Handler
```

它接收下一个 handler，返回一个新的 handler。多个中间件会形成调用链。

### API 版本和资源路径

本课把业务 API 放在 `/api/v1` 下：

```text
GET /api/v1/tasks/{taskID}
```

版本前缀不是所有项目都必须要有，但教学项目里显式写出来，可以帮助你建立“API 是长期契约”的意识。

### 统一错误响应

错误响应统一成 JSON：

```json
{"error":"task id is required"}
```

前端消费 API 时，稳定的错误结构比随意返回字符串更重要。

## 代码结构

```text
examples/lesson06-routing-middleware-and-api-shape/
  go.mod
  go.sum
  README.md
  main.go
  server.go
  server_test.go
```

- `main.go`：创建 logger 并启动服务。
- `server.go`：组织路由、中间件、handler 和响应辅助函数。
- `server_test.go`：测试健康检查、任务接口和 JSON 404。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson06-routing-middleware-and-api-shape
```

安装依赖：

```bash
go mod tidy
```

运行服务：

```bash
go run .
```

验证接口：

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/api/v1/tasks/task-1
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

路由和中间件集中在 `NewHandler`：

```go
r := chi.NewRouter()
r.Use(middleware.RequestID)
r.Use(middleware.Recoverer)
r.Use(logRequests(logger))
```

`middleware.RequestID` 给请求生成 ID，`middleware.Recoverer` 在 handler panic 时恢复请求，`logRequests` 是本课自己写的日志中间件。

中间件签名是：

```go
func logRequests(logger *log.Logger) func(http.Handler) http.Handler
```

内部通过包装 `http.ResponseWriter` 记录状态码：

```go
rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
next.ServeHTTP(rec, r)
```

业务路由按版本组织：

```go
r.Route("/api/v1", func(r chi.Router) {
	r.Get("/tasks/{taskID}", getTaskHandler)
})
```

handler 从路径中拿参数：

```go
taskID := chi.URLParam(r, "taskID")
```

## 练习

1. 新增 `GET /api/v1/tasks`，返回任务列表。
2. 在日志中输出请求耗时。
3. 新增一个中间件，把响应都加上 `X-App-Name: go-learning`。
4. 把 404 错误响应改成包含 `path` 字段。

## 常见坑

- 不要把所有路由都堆在 `main.go`，后续测试和维护会变难。
- 中间件顺序会影响行为，例如恢复中间件通常应该靠前注册。
- 错误响应结构要稳定，不要这个接口返回字符串，那个接口返回 JSON。
- 第三方路由库不是后端核心，核心仍然是 HTTP、handler、状态码和业务边界。
- 版本前缀一旦对外暴露，就要当作长期契约维护。

## 下一课

下一课会构建内存版 Task API，正式引入 handler、service、repository 的最小后端分层。
