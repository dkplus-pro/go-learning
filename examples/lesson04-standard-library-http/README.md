# 第 4 课 Demo：标准库 HTTP 服务

这个 demo 使用 Go 标准库实现一个最小 JSON HTTP API。

## 运行

```bash
go run .
```

健康检查：

```bash
curl -i http://localhost:8080/healthz
```

创建问候语：

```bash
curl -i -X POST http://localhost:8080/greetings \
  -H 'Content-Type: application/json' \
  -d '{"name":"Frontend Architect"}'
```

## 测试

```bash
go test ./...
```
