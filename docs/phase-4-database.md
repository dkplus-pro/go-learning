# 阶段 4：数据库与持久化

> 项目产出：数据库版任务 API — PostgreSQL + 迁移 + 事务 + 连接池

## 4.1 数据库驱动选择

Go 的 `database/sql` 只有接口，实际驱动由第三方提供。

```bash
go get github.com/lib/pq           # PostgreSQL 驱动（老牌）
# 或
go get github.com/jackc/pgx/v5    # PostgreSQL 驱动（更高性能，推荐）
go get github.com/golang-migrate/migrate/v4  # 迁移工具
```

**驱动对比：**

| | lib/pq | pgx/v5 |
|--|--------|--------|
| 接口 | database/sql | database/sql + 原生接口 |
| 性能 | 中等 | 高（零分配、二进制协议） |
| PostgreSQL 特性 | 基础 | 完整（COPY、LISTEN/NOTIFY 等） |
| 推荐场景 | 简单项目 | 生产项目 |

## 4.2 database/sql 核心模式

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib" // 仅注册驱动，不直接使用
)

// 打开连接池（不会立即连接）
db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 验证连接
if err := db.PingContext(ctx); err != nil {
    log.Fatal(err)
}
```

### CRUD 操作

```go
// 查询单行
var task Task
err := db.QueryRowContext(ctx,
    "SELECT id, title, done, created_at FROM tasks WHERE id = $1", id).
    Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)
if err == sql.ErrNoRows {
    // 没找到
}

// 查询多行
rows, err := db.QueryContext(ctx,
    "SELECT id, title, done, created_at FROM tasks ORDER BY created_at DESC")
if err != nil { return err }
defer rows.Close() // 必须关闭！

var tasks []Task
for rows.Next() {
    var t Task
    if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
        return err
    }
    tasks = append(tasks, t)
}
if err := rows.Err(); err != nil { return err } // 检查遍历中的错误

// 写入（用 $1, $2 参数化，防止 SQL 注入）
result, err := db.ExecContext(ctx,
    "INSERT INTO tasks (title, done) VALUES ($1, $2)", title, false)
// lastInsertID, rowsAffected := result.LastInsertId(), result.RowsAffected()

// 更新
_, err := db.ExecContext(ctx,
    "UPDATE tasks SET done = $1 WHERE id = $2", true, id)

// 删除
_, err := db.ExecContext(ctx,
    "DELETE FROM tasks WHERE id = $1", id)
```

**TS 对比：** `db.QueryRowContext` ≈ Prisma 的 `findUnique`，`db.QueryContext` ≈ `findMany`。但 Go 没有 ORM 的魔法 — SQL 你自己写，Scan 你自己映射。

### 事务

```go
// 正确的事务模式
func transferMoney(ctx context.Context, db *sql.DB, from, to int, amount float64) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil { return err }
    defer tx.Rollback() // 如果已 Commit，Rollback 是 no-op

    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    if err != nil { return err }

    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    if err != nil { return err }

    return tx.Commit() // Commit 后，defer Rollback 不会生效
}
```

**注意：** `defer tx.Rollback()` 是标准模式。Commit 成功后 Rollback 返回 `sql.ErrTxDone`，无害。

## 4.3 数据库迁移

```bash
# 安装 CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移文件
migrate create -ext sql -dir migrations -seq create_tasks_table
```

```sql
-- migrations/000001_create_tasks_table.up.sql
CREATE TABLE tasks (
    id         SERIAL PRIMARY KEY,
    title      TEXT NOT NULL CHECK (title <> ''),
    done       BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_done ON tasks(done);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

```sql
-- migrations/000001_create_tasks_table.down.sql
DROP TABLE IF EXISTS tasks;
```

```go
// 程序内运行迁移
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dbURL string) error {
    m, err := migrate.New("file://migrations", dbURL)
    if err != nil { return err }
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    return nil
}
```

## 4.4 连接池配置

```go
db.SetMaxOpenConns(25)                 // 最大打开连接数
db.SetMaxIdleConns(10)                 // 最大空闲连接数
db.SetConnMaxLifetime(5 * time.Minute) // 连接最大存活时间
db.SetConnMaxIdleTime(1 * time.Minute) // 空闲连接最大存活时间
```

**经验值：** 从 `MaxOpenConns` = CPU 核数 * 2 开始，根据数据库的 `max_connections` 和实际负载调。

## 4.5 sqlx — 减少样板代码

```go
import "github.com/jmoiron/sqlx"

// sqlx 包装了 database/sql，用 struct tag 自动 Scan
type Task struct {
    ID        int       `db:"id"`
    Title     string    `db:"title"`
    Done      bool      `db:"done"`
    CreatedAt time.Time `db:"created_at"`
}

// 再也不用写冗长的 Scan 了
tasks := []Task{}
err := db.SelectContext(ctx, &tasks, "SELECT * FROM tasks WHERE done = $1", false)

// 命名查询
task := Task{}
err := db.GetContext(ctx, &task,
    "SELECT * FROM tasks WHERE id = $1", id)
```

## 4.6 项目实战：数据库版任务 API

### 需求（在阶段 3 基础上）

- 用 PostgreSQL 替代内存存储
- 实现数据库迁移
- 所有操作走事务上下文
- 输入验证（title 不可为空、最大 200 字符）
- 优雅关闭时等待数据库 close

### 建议结构

```
cmd/
  tasksrv/
    main.go
internal/
  tasksrv/
    handler.go
    store.go         # 数据库操作层（sqlx 实现）
    task.go          # Task 模型
    errors.go
migrations/
  000001_create_tasks_table.up.sql
  000001_create_tasks_table.down.sql
```

### 关键点

```go
// store.go — 数据操作的封装
type Store struct {
    db *sqlx.DB
}

func (s *Store) Create(ctx context.Context, title string) (Task, error) {
    var t Task
    err := s.db.GetContext(ctx, &t,
        "INSERT INTO tasks (title) VALUES ($1) RETURNING *", title)
    return t, err
}

func (s *Store) List(ctx context.Context) ([]Task, error) {
    var tasks []Task
    err := s.db.SelectContext(ctx, &tasks,
        "SELECT * FROM tasks ORDER BY created_at DESC")
    return tasks, err
}

func (s *Store) Update(ctx context.Context, id int, title string) (Task, error) {
    var t Task
    err := s.db.GetContext(ctx, &t,
        "UPDATE tasks SET title=$1, updated_at=now() WHERE id=$2 RETURNING *", title, id)
    return t, err
}
```

### 学习检查清单

- [ ] 理解 `database/sql` 和驱动的注册机制（`import _ "driver"`）
- [ ] 能用迁移工具管理数据库 schema
- [ ] 掌握 QueryRow / Query / Exec 的使用场景
- [ ] 正确使用 `rows.Close()` 和 `rows.Err()`
- [ ] 理解事务的 ACID 和 Go 事务模式
- [ ] 能配置连接池参数并根据负载调整
- [ ] 了解 SQL 注入原理和参数化查询的防护
- [ ] 使用 sqlx 简化数据访问
- [ ] 正确处理 `sql.ErrNoRows`
- [ ] 独立完成数据库版任务 API

### 延伸阅读

- [database/sql tutorial](https://go.dev/doc/tutorial/database-access)
- [sqlx illustrated](https://jmoiron.github.io/sqlx/)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/current/index.html)
