# 第 10 课 Demo：用户、密码与 JWT 认证

这个 demo 实现注册、登录、bcrypt 密码哈希、JWT 签发校验和受保护 Task API。

## 安装依赖

```bash
go mod tidy
```

## 运行

```bash
JWT_SECRET=dev-secret go run .
```

注册：

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"super-secret"}'
```

登录：

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"super-secret"}'
```

访问受保护接口：

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Protect task API"}'
```

## 测试

```bash
go test ./...
```
