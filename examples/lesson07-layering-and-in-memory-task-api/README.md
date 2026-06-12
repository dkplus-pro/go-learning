# 第 7 课 Demo：后端分层与内存版 Task API

这个 demo 实现一个内存版 Task API，展示 handler、service、repository 的最小分层。

## 安装依赖

```bash
go mod tidy
```

## 运行

```bash
go run .
```

创建任务：

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Read Go documentation"}'
```

查询任务：

```bash
curl -i http://localhost:8080/api/v1/tasks/task-1
```

更新状态：

```bash
curl -i -X PATCH http://localhost:8080/api/v1/tasks/task-1/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"doing"}'
```

## 测试

```bash
go test ./...
```
