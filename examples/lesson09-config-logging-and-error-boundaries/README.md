# 第 9 课 Demo：配置、日志与错误边界

这个 demo 展示环境变量配置、`log/slog` 结构化日志、请求日志中间件和统一错误响应。

## 安装依赖

```bash
go mod tidy
```

## 运行

使用默认配置：

```bash
go run .
```

使用环境变量：

```bash
PORT=9090 REQUEST_TIMEOUT=1500ms LOG_LEVEL=debug go run .
```

验证错误响应：

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":" "}'
```

## 测试

```bash
go test ./...
```
