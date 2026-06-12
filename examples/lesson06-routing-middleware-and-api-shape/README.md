# 第 6 课 Demo：路由、中间件与 API 组织

这个 demo 使用 `chi` 组织路由，并实现请求日志、请求 ID、恢复和统一 JSON 错误响应。

## 安装依赖

```bash
go mod tidy
```

## 运行

```bash
go run .
```

验证接口：

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/api/v1/tasks/task-1
```

## 测试

```bash
go test ./...
```
