# 第 7 课：后端分层与内存版 Task API

## 学习目标

完成这一课后，你应该能够：

- 建立 handler、service、repository 的最小分层。
- 实现内存版 Task CRUD 的核心链路。
- 把业务规则放在 service 层，而不是散落在 handler。
- 用接口隔离 service 和具体存储实现。
- 分别测试 service 和 HTTP handler。

## 前端架构师视角

前端项目里你可能会拆 `components`、`hooks`、`stores`、`api`。后端也需要分层，但分层目标不同。后端分层首先是为了隔离协议、业务规则和数据访问：

- handler：处理 HTTP 协议细节，解析请求，返回响应。
- service：处理业务规则，例如任务状态能不能流转。
- repository：处理数据读写边界。

不要把后端分层理解成“照搬前端目录结构”。一个 handler 不应该直接操作 map 或数据库；一个 repository 也不应该决定任务能不能从 `done` 回到 `todo`。业务规则应该集中在 service。

## 核心概念

### Handler

handler 是 HTTP 边界。它关心：

- URL 和 HTTP 方法。
- JSON 请求体。
- 状态码和响应格式。
- 把错误映射成 HTTP 响应。

### Service

service 是业务规则中心。它关心：

- 创建任务时标题不能为空。
- 新任务默认是 `todo`。
- 任务状态只能按允许的方向流转。
- 找不到任务时返回稳定错误。

### Repository

repository 是数据边界。当前课程用内存 map 实现，后续课程会替换成 PostgreSQL。service 依赖 repository 接口，而不是依赖具体 map。

### 状态流转

本课任务状态包括：

```text
todo -> doing -> done
todo -> done
```

已完成任务不能回退到 `todo`。这个规则放在 service 层。

## 代码结构

```text
examples/lesson07-layering-and-in-memory-task-api/
  go.mod
  go.sum
  README.md
  main.go
  internal/
    task/
      handler.go
      handler_test.go
      model.go
      repository.go
      service.go
      service_test.go
```

- `model.go`：任务模型和状态规则。
- `repository.go`：repository 接口和内存实现。
- `service.go`：业务用例和错误定义。
- `handler.go`：HTTP 路由和 JSON 响应。
- `*_test.go`：覆盖 service 规则和 handler 行为。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson07-layering-and-in-memory-task-api
```

安装依赖：

```bash
go mod tidy
```

运行服务：

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

运行测试：

```bash
go test ./...
```

## 关键代码讲解

service 依赖接口：

```go
type Repository interface {
	Create(context.Context, Task) (Task, error)
	Find(context.Context, string) (Task, error)
	List(context.Context) ([]Task, error)
	Update(context.Context, Task) (Task, error)
}
```

这样后续把内存 repository 换成 PostgreSQL repository 时，service 的业务规则不需要重写。

创建任务在 service 层校验标题：

```go
title = strings.TrimSpace(title)
if title == "" {
	return Task{}, fmt.Errorf("%w: title is required", ErrValidation)
}
```

状态流转也在 service 层：

```go
if !CanTransition(task.Status, next) {
	return Task{}, fmt.Errorf("%w: %s to %s", ErrInvalidTransition, task.Status, next)
}
```

handler 只负责协议转换：

```go
task, err := h.service.CreateTask(r.Context(), req.Title)
if err != nil {
	h.writeError(w, err)
	return
}
```

这种分层让测试更聚焦。service 测业务规则，handler 测 HTTP 协议表现。

## 练习

1. 新增 `DELETE /api/v1/tasks/{taskID}`，实现删除任务。
2. 给任务增加 `Description string` 字段。
3. 给 `List` 增加按状态过滤，例如 `/api/v1/tasks?status=todo`。
4. 给状态流转补充更多测试 case。

## 常见坑

- 不要让 handler 直接操作 repository，这会让 HTTP 和业务规则耦合。
- 不要让 repository 决定业务规则，它只负责数据读写。
- 内存 map 在并发请求下要加锁。
- 错误要有稳定类别，handler 才能正确映射状态码。
- 业务测试不应该依赖真实 HTTP，HTTP 测试也不应该验证所有业务细节。

## 下一课

下一课会把内存 repository 替换为 PostgreSQL，并加入 Docker Compose、本地 schema 和数据库边界测试。
