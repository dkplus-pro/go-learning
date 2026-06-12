# 第 5 课 Demo：Context、超时与请求生命周期

这个 demo 展示如何用 `context.Context` 控制请求级超时。

## 运行

```bash
go run .
```

验证接口：

```bash
curl -i http://localhost:8080/reports/frontend-architecture
```

## 测试

```bash
go test ./...
```
