# 第 11 课 Demo：并发、后台任务与优雅关闭

这个 demo 实现一个内存后台任务队列，展示 goroutine、channel、context 取消和 HTTP 服务优雅关闭。

## 运行

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

## 测试

```bash
go test ./...
```
