# Go 后端课程大纲

## 课程定位

这套课程服务于有 7 年左右前端架构经验、希望系统转型 Go 后端开发的学习者。

课程不假设学习者懂 Go，但假设学习者具备成熟的工程经验，理解 HTTP、JSON、模块化、测试、构建、部署和前端应用架构。每一课都会用前端/TypeScript/Node.js 的经验做对照，帮助学习者判断哪些经验可以迁移，哪些习惯需要调整。

## 技术栈

- Go 1.24+
- 标准库优先：`net/http`、`context`、`testing`、`database/sql`、`log/slog`
- 路由：`chi`
- 数据库：PostgreSQL
- SQL 访问：前期手写 `database/sql`，后期视复杂度引入更清晰的组织方式
- 配置：环境变量优先，必要时引入轻量辅助库
- 认证：JWT
- 本地依赖编排：Docker Compose

## 课程产物约定

每一课包含一个独立文档和一个独立 demo：

```text
docs/lessons/01-hello-go.md
examples/lesson01-hello-go/
```

每个 demo 都是独立 Go module，包含：

```text
go.mod
README.md
main.go 或 cmd/server/main.go
必要的包目录
必要的测试文件
```

每课 demo 至少支持：

```bash
go run .
go test ./...
```

涉及 HTTP 服务时，额外提供 `curl` 验证命令。涉及数据库时，额外提供 Docker Compose 启动命令。

## 文档模板

每课文档使用统一结构：

```text
# 第 X 课：标题

## 学习目标
## 前端架构师视角
## 核心概念
## 代码结构
## 运行方式
## 关键代码讲解
## 练习
## 常见坑
## 下一课
```

## 12 课路线

### 第 1 课：Hello Go 与最小工程

文档路径：`docs/lessons/01-hello-go.md`

Demo 路径：`examples/lesson01-hello-go/`

目标：

- 安装和认识 Go 工具链的基本命令。
- 理解 `go.mod`、`package main`、`func main`。
- 写第一个可运行 CLI demo。
- 写第一个纯函数和单元测试。
- 建立 `go run .` 与 `go test ./...` 的习惯。

前端迁移重点：

- Go module 与 npm package 的差异。
- `main` 包和应用入口与前端入口文件的差异。
- Go 的测试文件命名和运行方式。

### 第 2 课：类型、函数、错误处理

文档路径：`docs/lessons/02-types-functions-and-errors.md`

Demo 路径：`examples/lesson02-types-functions-and-errors/`

目标：

- 掌握基础类型、切片、map、struct。
- 理解多返回值和显式错误处理。
- 用表格驱动测试覆盖核心逻辑。
- 通过一个小型输入校验器理解 Go 的数据建模。

前端迁移重点：

- `struct` 与 TypeScript `type/interface` 的边界。
- 显式错误返回与 `throw/catch` 的差异。
- 表格驱动测试与前端参数化测试的相似点。

### 第 3 课：方法、接口与包设计

文档路径：`docs/lessons/03-methods-interfaces-and-packages.md`

Demo 路径：`examples/lesson03-methods-interfaces-and-packages/`

目标：

- 理解方法接收者、值接收者和指针接收者。
- 理解小接口的设计方式。
- 拆分 package，建立最小业务分层。
- 为接口行为编写测试。

前端迁移重点：

- Go interface 是能力约束，不是数据形状声明。
- Go 没有 class 继承，组合优先。
- 包边界比“大而全的 service 对象”更重要。

### 第 4 课：标准库 HTTP 服务

文档路径：`docs/lessons/04-standard-library-http.md`

Demo 路径：`examples/lesson04-standard-library-http/`

目标：

- 使用 `net/http` 编写第一个 HTTP API。
- 理解 handler、request、response。
- 处理 JSON 请求和响应。
- 使用 `httptest` 测试 handler。

前端迁移重点：

- handler 与前端路由组件/服务端路由的区别。
- JSON 编解码与 TypeScript 类型校验的差异。
- HTTP 测试不等于只用浏览器手动点。

### 第 5 课：Context、超时与请求生命周期

文档路径：`docs/lessons/05-context-timeout-and-lifecycle.md`

Demo 路径：`examples/lesson05-context-timeout-and-lifecycle/`

目标：

- 理解 `context.Context` 的用途。
- 实现请求超时、取消和跨层传递。
- 建立请求级日志字段。
- 测试超时和取消行为。

前端迁移重点：

- `context` 不是全局状态，也不是 React Context。
- 请求取消与 `AbortController` 的相似点和差异。
- 后端要主动控制资源生命周期。

### 第 6 课：路由、中间件与 API 组织

文档路径：`docs/lessons/06-routing-middleware-and-api-shape.md`

Demo 路径：`examples/lesson06-routing-middleware-and-api-shape/`

目标：

- 引入 `chi` 路由。
- 组织 REST API 路径。
- 编写日志、恢复、请求 ID 中间件。
- 统一错误响应格式。

前端迁移重点：

- 中间件与前端拦截器、插件机制的对应关系。
- 路由树与页面路由树的差异。
- API 响应结构要服务于前端消费，但不被前端状态结构绑架。

### 第 7 课：后端分层与内存版 Task API

文档路径：`docs/lessons/07-layering-and-in-memory-task-api.md`

Demo 路径：`examples/lesson07-layering-and-in-memory-task-api/`

目标：

- 建立 handler、service、repository 的最小分层。
- 实现内存版 Task CRUD。
- 设计 Task 状态流转。
- 为 service 和 handler 分别编写测试。

前端迁移重点：

- 后端分层不是照搬前端目录结构。
- repository 是数据边界，不是任意工具类集合。
- 业务规则应该集中在 service，而不是散落在 handler。

### 第 8 课：PostgreSQL 与持久化

文档路径：`docs/lessons/08-postgresql-and-persistence.md`

Demo 路径：`examples/lesson08-postgresql-and-persistence/`

目标：

- 使用 Docker Compose 启动 PostgreSQL。
- 建立 schema 和迁移脚本。
- 使用 Go 连接数据库。
- 将 Task API 从内存存储切换到 PostgreSQL。
- 测试 repository 边界。

前端迁移重点：

- 数据库 schema 是长期契约，不是临时 JSON shape。
- SQL 查询性能和索引需要提前进入设计视野。
- 本地数据库环境应该可重复启动。

### 第 9 课：配置、日志与错误边界

文档路径：`docs/lessons/09-config-logging-and-error-boundaries.md`

Demo 路径：`examples/lesson09-config-logging-and-error-boundaries/`

目标：

- 使用环境变量管理配置。
- 使用 `log/slog` 输出结构化日志。
- 区分业务错误、输入错误和系统错误。
- 统一 HTTP 错误映射。

前端迁移重点：

- 后端错误信息需要兼顾用户、调用方和排障。
- 结构化日志比字符串拼接更适合生产环境。
- 配置应该由运行环境注入，而不是写死在代码里。

### 第 10 课：用户、密码与 JWT 认证

文档路径：`docs/lessons/10-users-passwords-and-jwt.md`

Demo 路径：`examples/lesson10-users-passwords-and-jwt/`

目标：

- 实现用户注册和登录。
- 使用安全方式存储密码哈希。
- 签发和验证 JWT。
- 编写认证中间件保护 Task API。
- 测试认证成功和失败路径。

前端迁移重点：

- JWT 对前端友好，但安全边界在后端。
- 密码永远不应该明文存储或出现在日志中。
- 认证状态和授权判断不能只依赖前端。

### 第 11 课：并发、后台任务与优雅关闭

文档路径：`docs/lessons/11-concurrency-background-jobs-and-shutdown.md`

Demo 路径：`examples/lesson11-concurrency-background-jobs-and-shutdown/`

目标：

- 理解 goroutine 和 channel 的基础用法。
- 实现一个简单后台任务。
- 使用 context 控制后台任务生命周期。
- 实现 HTTP 服务优雅关闭。
- 测试关键并发逻辑。

前端迁移重点：

- goroutine 不是 Promise。
- channel 不是事件总线的万能替代品。
- 后端服务要能正确处理启动、运行和退出。

### 第 12 课：整合、测试与本地交付

文档路径：`docs/lessons/12-integration-testing-and-local-delivery.md`

Demo 路径：`examples/lesson12-integration-testing-and-local-delivery/`

目标：

- 整合 Task API 的核心能力。
- 使用 Docker Compose 一键启动服务和 PostgreSQL。
- 编写 API smoke test 和集成测试。
- 补充 Makefile 或等价命令入口。
- 梳理从本地开发到部署前检查的流程。

前端迁移重点：

- 后端交付物不只是源码，还包括配置、数据库、启动方式和验证方式。
- 集成测试用于验证关键链路，不替代所有单元测试。
- 一个可维护服务需要清晰的运行手册。

## 依赖引入节奏

```text
第 1-5 课：只使用 Go 标准库
第 6 课：引入 chi
第 8 课：引入 PostgreSQL 驱动
第 10 课：引入 JWT 相关库和密码哈希能力
第 11-12 课：不再引入非必要新概念，重点做工程整合
```

## 提交节奏

第一阶段提交课程总体规划。

之后每课一个提交，提交范围包含：

- 当前课程文档
- 当前课程 demo
- 必要的仓库首页或索引更新

原则上不在后续课程提交中回头大改早期课程，除非发现明确错误。
