# 第 5 课：Context、超时与请求生命周期

## 学习目标

完成这一课后，你应该能够：

- 理解 `context.Context` 在后端请求生命周期中的作用。
- 使用 `context.WithTimeout` 给请求设置超时。
- 在业务函数中响应取消信号。
- 区分正常完成、超时和取消。
- 为超时行为编写 HTTP handler 测试。

## 前端架构师视角

`context.Context` 不是 React Context，也不是全局状态容器。它更像请求生命周期的控制句柄：请求什么时候开始、什么时候超时、客户端什么时候断开、跨层调用应该携带哪些请求级元信息。

如果你熟悉浏览器里的 `AbortController`，可以把 `context` 理解成后端调用链里的取消信号。区别是 Go 后端通常会把 `context` 作为第一个参数一路传下去，让数据库查询、HTTP 调用、后台任务都能响应同一个取消信号。

后端不能假设所有请求都会正常跑完。客户端可能断开，网关可能超时，服务自己也应该设置超时，避免资源被慢请求占住。

## 核心概念

### context.Context

`context.Context` 主要承载三类信息：

- 取消信号：`Done()` 返回一个 channel。
- 错误原因：`Err()` 返回 `context.Canceled` 或 `context.DeadlineExceeded`。
- 请求级值：`Value()` 可以携带少量跨层元信息。

业务代码最常用的是取消和超时。

### context.WithTimeout

`context.WithTimeout` 基于父 context 派生出一个带截止时间的子 context：

```go
ctx, cancel := context.WithTimeout(r.Context(), 150*time.Millisecond)
defer cancel()
```

`defer cancel()` 很重要。即使请求提前完成，也应该释放计时器资源。

### select 响应取消

一个会阻塞的业务函数应该同时等待业务结果和 context 取消：

```go
select {
case <-time.After(delay):
	return result, nil
case <-ctx.Done():
	return result, ctx.Err()
}
```

这让调用方可以主动停止无意义的工作。

## 代码结构

```text
examples/lesson05-context-timeout-and-lifecycle/
  go.mod
  README.md
  main.go
  server.go
  server_test.go
```

- `main.go`：启动带超时配置的 HTTP 服务。
- `server.go`：定义 report service、handler 和 JSON 响应。
- `server_test.go`：测试正常响应和超时响应。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson05-context-timeout-and-lifecycle
```

运行服务：

```bash
go run .
```

验证接口：

```bash
curl -i http://localhost:8080/reports/frontend-architecture
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

handler 为每个请求设置超时：

```go
ctx, cancel := context.WithTimeout(r.Context(), timeout)
defer cancel()
```

service 不直接 `time.Sleep` 到底，而是监听 `ctx.Done()`：

```go
select {
case <-time.After(s.Delay):
	return Report{ID: id, Status: "ready"}, nil
case <-ctx.Done():
	return Report{}, ctx.Err()
}
```

当 service 返回 `context.DeadlineExceeded` 时，handler 把它映射成 `504 Gateway Timeout`：

```go
if errors.Is(err, context.DeadlineExceeded) {
	writeJSON(w, http.StatusGatewayTimeout, errorResponse{Error: "report generation timed out"})
	return
}
```

测试里通过不同的 service delay 和 timeout 控制成功与失败：

```go
handler := NewHandler(ReportService{Delay: time.Millisecond}, 50*time.Millisecond)
```

这比在测试里等待很久更稳定。

## 练习

1. 新增一个 `GET /slow-reports/{id}` 路由，让它更容易触发超时。
2. 把超时时间改成从环境变量读取。
3. 在响应里返回 `request_id` 字段，为下一课中间件做准备。

## 常见坑

- 不要把 `context.Context` 存到 struct 里，通常应该作为函数第一个参数传入。
- 使用 `WithTimeout` 或 `WithCancel` 后要调用 `cancel()`。
- 不要用 context 传业务参数，普通参数更清晰。
- 阻塞函数如果不监听 `ctx.Done()`，调用方取消也不会真正停止工作。
- 超时不是业务失败，它是资源保护机制。

## 下一课

下一课会引入 `chi`，把路由、中间件、请求 ID、日志和统一错误响应组织起来。
