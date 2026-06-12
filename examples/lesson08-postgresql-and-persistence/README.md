# 第 8 课 Demo：PostgreSQL 与持久化

这个 demo 把 Task API 的 repository 从内存实现替换为 PostgreSQL 实现。

## 启动数据库

```bash
docker compose up -d
```

## 设置环境变量

```bash
export DATABASE_URL='postgres://app:secret@localhost:5432/taskdb?sslmode=disable'
```

## 运行服务

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

## 测试

不连接数据库时，数据库集成测试会自动跳过：

```bash
go test ./...
```

连接本地 PostgreSQL 后运行完整测试：

```bash
DATABASE_URL='postgres://app:secret@localhost:5432/taskdb?sslmode=disable' go test ./...
```
