# 第 11 课：并发、后台任务与优雅关闭

## 学习目标

完成这一课后，你应该能够：

- 使用 goroutine 启动后台 worker。
- 使用 channel 投递后台任务。
- 使用 mutex 保护共享内存状态。
- 使用 context 控制 worker 生命周期。
- 使用 `http.Server.Shutdown` 实现优雅关闭。

## 前端架构师视角

如果你熟悉 Promise、任务队列、Web Worker 或 Node.js 事件循环，很容易把 goroutine 理解成“轻量 Promise”。这个类比不准确。goroutine 是由 Go runtime 调度的并发执行单元，它没有自带返回值、错误通道或生命周期管理。

channel 也不是事件总线万能替代品。它更适合表达明确的并发协作关系：谁发送、谁接收、什么时候关闭、满了怎么办。

后端服务还要考虑退出过程。进程收到 SIGTERM 时，不能粗暴退出：正在处理的请求要给一点时间完成，后台任务要响应取消，资源要释放。

## 核心概念

### goroutine

使用 `go` 关键字启动 goroutine：

```go
go q.worker(ctx)
```

goroutine 很便宜，但不是免费。启动后要知道它什么时候退出。

### channel

本课用 channel 投递任务 ID：

```go
pending chan string
```

HTTP handler 创建任务后，把任务 ID 送进 channel。worker 从 channel 取出 ID 并处理。

### context 取消

worker 同时监听任务和取消信号：

```go
select {
case id := <-q.pending:
	q.process(ctx, id)
case <-ctx.Done():
	return
}
```

当服务关闭时，context 被取消，worker 退出。

### 优雅关闭

`http.Server.Shutdown` 会停止接收新请求，并等待已有请求结束：

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(shutdownCtx)
```

## 代码结构

```text
examples/lesson11-concurrency-background-jobs-and-shutdown/
  go.mod
  README.md
  main.go
  server.go
  server_test.go
  internal/
    jobs/
      queue.go
      queue_test.go
```

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson11-concurrency-background-jobs-and-shutdown
```

运行服务：

```bash
go run .
```

创建后台任务：

```bash
curl -i -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"payload":"send weekly digest"}'
```

查询任务：

```bash
curl -i http://localhost:8080/jobs/job-1
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

队列创建任务后通过 channel 投递：

```go
select {
case q.pending <- job.ID:
	return job, nil
case <-ctx.Done():
	return Job{}, ctx.Err()
}
```

worker 循环监听任务和取消：

```go
for {
	select {
	case id := <-q.pending:
		q.process(ctx, id)
	case <-ctx.Done():
		return
	}
}
```

共享 map 用 mutex 保护：

```go
q.mu.Lock()
defer q.mu.Unlock()
```

`main.go` 通过信号创建根 context：

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

服务退出时先取消 worker，再给 HTTP server 一个关闭窗口。

## 练习

1. 支持多个 worker 并发处理任务。
2. 给任务增加失败状态，并让 payload 为空时失败。
3. 给 `/jobs` 增加列表接口。
4. 把队列满时的错误映射成 `429 Too Many Requests`。

## 常见坑

- goroutine 启动后要有退出路径。
- channel 满了时发送会阻塞，生产服务要考虑缓冲和背压。
- map 并发读写必须加锁。
- 不要在 signal handler 里做复杂逻辑，用 context 传递关闭信号更清晰。
- `Shutdown` 不是 `Close`，它会尽量等待正在处理的请求完成。

## 下一课

下一课会把前面课程整合成一个本地可交付的 Task API demo，补齐 Docker Compose、Makefile、集成测试和 smoke test。
