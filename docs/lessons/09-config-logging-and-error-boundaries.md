# 第 9 课：配置、日志与错误边界

## 学习目标

完成这一课后，你应该能够：

- 从环境变量读取服务配置。
- 使用 `log/slog` 输出结构化日志。
- 编写请求日志中间件。
- 设计稳定的 API 错误响应结构。
- 区分公开错误信息和内部排障信息。

## 前端架构师视角

前端应用也有配置，例如 API base URL、构建环境、功能开关。后端配置更直接影响服务运行：监听端口、数据库连接、请求超时、日志级别、第三方凭证等都应该由运行环境注入。

日志也不是 `console.log` 的简单替代。生产后端需要结构化日志，方便按字段检索请求路径、状态码、耗时、错误类型和请求 ID。

错误边界同样重要。前端需要稳定、可解析的错误结构；后端排障需要完整错误上下文。不要把内部错误原样暴露给客户端，也不要把对前端有用的错误全部压成 `internal error`。

## 核心概念

### 环境变量配置

本课支持这些配置：

```text
PORT=8080
REQUEST_TIMEOUT=2s
LOG_LEVEL=info
```

配置加载代码会提供默认值，也会验证非法值。

### slog

`log/slog` 是 Go 标准库提供的结构化日志包。JSON handler 可以输出适合日志系统采集的结构：

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
```

### 请求日志中间件

中间件在请求完成后记录：

- HTTP 方法
- URL 路径
- 状态码
- 请求耗时
- request id

这些字段能帮助你在生产环境定位慢请求和失败请求。

### 错误边界

service 返回稳定的领域错误。handler 把它映射成 HTTP 状态码和统一 JSON：

```json
{
  "error": {
    "code": "validation_error",
    "message": "title is required"
  }
}
```

客户端依赖 `code` 做分支判断，比解析错误字符串更稳。

## 代码结构

```text
examples/lesson09-config-logging-and-error-boundaries/
  go.mod
  go.sum
  README.md
  main.go
  internal/
    config/
      config.go
      config_test.go
    task/
      errors.go
      handler.go
      handler_test.go
      service.go
```

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson09-config-logging-and-error-boundaries
```

安装依赖：

```bash
go mod tidy
```

使用默认配置运行：

```bash
go run .
```

使用环境变量运行：

```bash
PORT=9090 REQUEST_TIMEOUT=1500ms LOG_LEVEL=debug go run .
```

验证错误响应：

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":" "}'
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

配置加载函数允许注入 `getenv`，方便测试：

```go
func LoadFromEnv(getenv func(string) string) (Config, error)
```

生产代码传 `os.Getenv`，测试传一个 map-backed 函数。

日志级别用 `slog.Level` 表达：

```go
level := new(slog.LevelVar)
level.Set(cfg.LogLevel)
```

handler 里不直接返回 `err.Error()` 给客户端，而是通过 `writeError` 做边界映射：

```go
case errors.Is(err, ErrValidation):
	writeAPIError(w, http.StatusBadRequest, "validation_error", publicMessage(err))
```

未知错误会记录到日志，但客户端只收到稳定的内部错误响应：

```go
logger.Error("unhandled application error", "error", err)
```

## 练习

1. 增加 `APP_ENV` 配置，支持 `development` 和 `production`。
2. 让请求日志在 5xx 时使用 `logger.Error`，其他状态使用 `logger.Info`。
3. 给错误响应增加 `request_id` 字段。
4. 给配置加载增加 `DATABASE_URL` 必填校验。

## 常见坑

- 不要把运行配置写死在代码里。
- 日志字段要稳定，方便检索和聚合。
- 不要把数据库错误、栈信息、密钥等内部细节返回给客户端。
- 前端需要稳定错误码，不应该依赖自然语言错误字符串。
- 配置解析失败应该在服务启动时尽早失败。

## 下一课

下一课会加入用户、密码哈希和 JWT 认证，把 Task API 放到认证边界之后。
