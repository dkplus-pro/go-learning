# 第 12 课：整合、测试与本地交付

## 学习目标

完成这一课后，你应该能够：

- 把注册登录、JWT、Task API 和 PostgreSQL 整合成一个可运行服务。
- 使用 Docker Compose 启动本地依赖。
- 使用 Makefile 固化常用命令。
- 编写数据库集成测试和 API smoke test。
- 输出一份足够清晰的本地运行手册。

## 前端架构师视角

前端交付物通常包含静态资源、环境变量和部署配置。后端交付物还必须包含运行依赖、数据库 schema、启动顺序、健康检查、迁移方式和验证命令。

一个后端 demo 不能只说“代码在这里”。它应该能回答：

- 依赖怎么启动？
- 环境变量怎么设置？
- 服务怎么跑？
- 怎么验证主链路？
- 测试怎么跑？
- 失败时先看哪里？

本课把前面课程里的能力合并成一个小型 Task API，并用 Makefile 和 smoke script 固化本地交付流程。

## 核心概念

### 本地交付

本地交付不是生产部署，但它应该稳定复现生产关键依赖。本课使用：

```text
docker-compose.yml
Makefile
scripts/smoke.sh
```

### 集成测试

单元测试验证函数和小边界。集成测试验证应用和真实数据库之间的契约。本课数据库测试默认跳过，设置 `DATABASE_URL` 后运行。

### Smoke Test

Smoke test 验证最关键链路是否可用：

```text
healthz -> register -> login -> create task -> list tasks -> update status
```

它不替代完整测试，但能快速证明服务基本可运行。

## 代码结构

```text
examples/lesson12-integration-testing-and-local-delivery/
  docker-compose.yml
  Makefile
  README.md
  go.mod
  go.sum
  main.go
  scripts/
    smoke.sh
  internal/
    app/
      app.go
      app_test.go
      handler.go
      schema.go
      schema.sql
      token.go
      token_test.go
```

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson12-integration-testing-and-local-delivery
```

启动 PostgreSQL：

```bash
make compose-up
```

运行测试：

```bash
make test
```

运行包含数据库的集成测试：

```bash
make integration-test
```

启动服务：

```bash
make run
```

在另一个终端运行 smoke test：

```bash
make smoke
```

停止依赖：

```bash
make compose-down
```

## 关键代码讲解

`main.go` 负责启动编排：

```go
db, err := sql.Open("pgx", databaseURL)
app.Migrate(ctx, db)
application := app.New(db, app.NewTokenManager(secret, 2*time.Hour))
```

schema 由 Go embed 进入二进制：

```go
//go:embed schema.sql
var schemaSQL string
```

认证中间件验证 `Authorization: Bearer <token>`，并把用户放进 context：

```go
user, err := a.tokens.Verify(token)
ctx := context.WithValue(r.Context(), userContextKey{}, user)
```

Makefile 固化常用命令，减少文档和实际操作漂移：

```makefile
test:
	go test ./...
```

smoke 脚本用 curl 走一遍主链路，适合本地改动后的快速确认。

## 练习

1. 给 `/api/v1/tasks` 增加分页。
2. 给任务增加 `description` 字段并编写迁移。
3. 把 JWT secret 改成必须从安全配置系统注入。
4. 增加 Dockerfile，把服务也纳入 Docker Compose。

## 常见坑

- Makefile 命令要和 README 保持一致。
- smoke test 不应该依赖手动复制 token。
- 集成测试需要可重复执行，测试数据要避免固定冲突。
- 本地 schema 自动执行方便教学，但生产环境应该使用正式迁移工具。
- 健康检查应该只验证服务能响应，不要做昂贵业务操作。

## 课程收束

到这里，你已经从最小 Go 程序走到一个可本地交付的 Go 后端服务。下一步可以沿三个方向继续加深：

- 工程化：迁移工具、CI、Dockerfile、部署配置。
- 后端能力：分页、授权、事务、缓存、消息队列。
- 生产质量：可观测性、压测、限流、安全审计。
