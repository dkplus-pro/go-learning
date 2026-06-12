# 第 10 课：用户、密码与 JWT 认证

## 学习目标

完成这一课后，你应该能够：

- 实现用户注册和登录。
- 使用 bcrypt 存储密码哈希。
- 签发和校验 JWT。
- 编写认证中间件保护 API。
- 测试认证成功和失败路径。

## 前端架构师视角

前端经常处理“登录态”，例如保存 token、刷新页面后恢复用户信息、在请求头里带上 `Authorization`。但真正的安全边界在后端。前端可以改善体验，不能作为权限判断的可信来源。

密码永远不应该明文存储，也不应该出现在日志里。后端只保存密码哈希。登录时用用户提交的密码和哈希做校验。

JWT 对前端很友好，因为它是一个可以放进请求头的字符串。但后端必须验证签名、过期时间和声明内容。不要只把 token 解码出来就信任它。

## 核心概念

### bcrypt

bcrypt 是专门用于密码哈希的算法。它会自动加入 salt，并且计算成本可配置：

```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

验证密码时不要自己比较字符串，而是使用：

```go
bcrypt.CompareHashAndPassword(hash, []byte(password))
```

### JWT

JWT 由 header、payload、signature 三部分组成。payload 里可以放用户 ID、邮箱、过期时间等声明。签名用于防篡改。

本课使用 HMAC SHA-256：

```go
jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
```

### Authorization Header

受保护接口要求请求头：

```text
Authorization: Bearer <token>
```

认证中间件负责解析 header、验证 token，并把用户信息放入 request context。

### 认证和授权

认证回答“你是谁”。授权回答“你能做什么”。本课重点是认证，受保护的 Task API 会把当前用户 ID 写入任务 owner 字段，为后续授权打基础。

## 代码结构

```text
examples/lesson10-users-passwords-and-jwt/
  go.mod
  go.sum
  README.md
  main.go
  internal/
    auth/
      errors.go
      handler.go
      handler_test.go
      jwt.go
      jwt_test.go
      service.go
      service_test.go
```

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson10-users-passwords-and-jwt
```

安装依赖：

```bash
go mod tidy
```

运行服务：

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

把登录响应里的 token 放进环境变量：

```bash
export TOKEN='paste-token-here'
```

访问受保护接口：

```bash
curl -i -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Protect task API"}'
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

注册时只保存密码哈希：

```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

登录时校验密码：

```go
if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) != nil {
	return "", ErrUnauthorized
}
```

JWT claims 保存用户身份和过期时间：

```go
claims := UserClaims{
	UserID: user.ID,
	Email:  user.Email,
	RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
	},
}
```

认证中间件把用户信息写入 context：

```go
ctx := context.WithValue(r.Context(), userContextKey{}, claims.User())
next.ServeHTTP(w, r.WithContext(ctx))
```

handler 从 context 读取当前用户：

```go
user, ok := UserFromContext(r.Context())
```

## 练习

1. 增加 `GET /api/v1/me`，返回当前用户信息。
2. 给 JWT 增加 issuer，并在验证时检查 issuer。
3. 给注册接口增加邮箱格式校验。
4. 实现“只有任务 owner 可以读取任务”的授权规则。

## 常见坑

- 不要保存明文密码。
- 不要把密码或 token 写进日志。
- JWT 必须验证签名和过期时间，不能只 decode。
- `JWT_SECRET` 不能使用默认值上生产。
- 前端持有 token 不代表前端可以决定权限。

## 下一课

下一课会讲 goroutine、channel、后台任务和 HTTP 服务优雅关闭。
