# 第 8 课：PostgreSQL 与持久化

## 学习目标

完成这一课后，你应该能够：

- 使用 Docker Compose 启动 PostgreSQL。
- 用 `database/sql` 连接数据库。
- 理解 schema 是后端服务的长期契约。
- 把内存 repository 替换成 PostgreSQL repository。
- 编写可选的数据库集成测试。

## 前端架构师视角

前端状态通常可以重新拉取、重新计算或存在浏览器缓存里。后端数据库不同：它是系统事实来源。schema 一旦上线，就会承载真实数据和长期兼容压力。

你在 TypeScript 里定义的接口更多是编译期约束；数据库 schema 是运行期和存储层约束。字段类型、非空限制、索引、状态枚举和迁移策略都会影响系统长期演进。

本课保留第 7 课的 handler、service、repository 分层，但把 repository 实现换成 PostgreSQL。service 依赖接口，所以业务规则不用因为存储变了而重写。

## 核心概念

### database/sql

`database/sql` 是 Go 标准库里的数据库抽象层。它不包含具体数据库驱动，需要引入 PostgreSQL 驱动。本课使用 `pgx` 的 stdlib 适配：

```go
_ "github.com/jackc/pgx/v5/stdlib"
```

代码中仍然通过标准库接口使用数据库：

```go
db, err := sql.Open("pgx", databaseURL)
```

### Schema

本课的 `tasks` 表包含：

```sql
CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('todo', 'doing', 'done')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

`CHECK` 约束让数据库也能保护状态字段，不只依赖应用代码。

### Repository 替换

service 层依赖接口：

```go
type Repository interface {
	Create(context.Context, Task) (Task, error)
	Find(context.Context, string) (Task, error)
	List(context.Context) ([]Task, error)
	Update(context.Context, Task) (Task, error)
}
```

第 7 课的内存实现和本课的 PostgreSQL 实现都满足这个接口。

### 集成测试

数据库测试需要真实 PostgreSQL。本课测试默认跳过，只有设置 `DATABASE_URL` 时才运行：

```bash
DATABASE_URL='postgres://app:secret@localhost:5432/taskdb?sslmode=disable' go test ./...
```

这样没有数据库时仍然可以跑普通编译和测试，有数据库时可以验证真实 repository 行为。

## 代码结构

```text
examples/lesson08-postgresql-and-persistence/
  docker-compose.yml
  go.mod
  go.sum
  README.md
  main.go
  internal/
    task/
      handler.go
      model.go
      postgres_repository.go
      postgres_repository_test.go
      schema.go
      schema.sql
      service.go
      service_test.go
```

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson08-postgresql-and-persistence
```

启动 PostgreSQL：

```bash
docker compose up -d
```

设置连接字符串：

```bash
export DATABASE_URL='postgres://app:secret@localhost:5432/taskdb?sslmode=disable'
```

运行服务：

```bash
go run .
```

创建任务：

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Persist tasks in PostgreSQL"}'
```

查询列表：

```bash
curl -i http://localhost:8080/api/v1/tasks
```

运行测试：

```bash
go test ./...
```

运行数据库集成测试：

```bash
DATABASE_URL='postgres://app:secret@localhost:5432/taskdb?sslmode=disable' go test ./...
```

停止数据库：

```bash
docker compose down
```

## 关键代码讲解

`main.go` 从环境变量读取数据库连接：

```go
databaseURL := os.Getenv("DATABASE_URL")
if databaseURL == "" {
	log.Fatal("DATABASE_URL is required")
}
```

启动时先 ping 数据库，再执行 schema：

```go
if err := db.PingContext(ctx); err != nil {
	log.Fatal(err)
}

if err := task.Migrate(ctx, db); err != nil {
	log.Fatal(err)
}
```

PostgreSQL repository 使用 SQL 查询实现接口：

```go
row := r.db.QueryRowContext(ctx, `
	SELECT id, title, status, created_at
	FROM tasks
	WHERE id = $1
`, id)
```

找不到数据时把 `sql.ErrNoRows` 转成业务层稳定错误：

```go
if errors.Is(err, sql.ErrNoRows) {
	return Task{}, ErrNotFound
}
```

handler 和 service 与第 7 课保持同样边界。存储从 map 换成 PostgreSQL，但 API 和业务规则没有大改。

## 练习

1. 给 `tasks.status` 增加索引，并解释它适合哪些查询。
2. 新增 `updated_at` 字段，在更新状态时同步更新。
3. 给 `List` 增加分页参数 `limit` 和 `offset`。
4. 把 `schema.sql` 拆成带版本号的迁移目录。

## 常见坑

- 不要把数据库连接字符串写死在代码里。
- 不要忽略 `rows.Close()` 和 `rows.Err()`。
- `sql.ErrNoRows` 应该转换成业务错误，不应该直接泄漏到 HTTP 响应。
- 本地数据库要能被一条命令启动，否则新成员很难复现环境。
- schema 是长期契约，字段和约束修改要谨慎。

## 下一课

下一课会补齐配置、结构化日志和错误边界，让服务更接近生产环境的运行方式。
