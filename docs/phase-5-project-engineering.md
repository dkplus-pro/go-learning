# 阶段 5：项目工程化

> 项目产出：用户认证系统 — 注册/登录/JWT/刷新令牌/权限中间件，具备生产级项目骨架

## 5.1 Go 项目布局

Go 社区没有官方标准，但有一套被广泛接受的惯例：

```
.
├── cmd/                    # 入口程序（每个子目录是一个可执行文件）
│   └── apiserver/
│       └── main.go
├── internal/               # 私有代码（编译器强制外部不能 import）
│   ├── auth/               # 认证领域
│   ├── user/               # 用户领域
│   └── middleware/
├── pkg/                    # 可被外部引用的公共库（谨慎使用）
│   └── jwtutil/
├── migrations/             # 数据库迁移文件
├── config/                 # 配置文件（yaml/toml）
├── api/                    # OpenAPI/Swagger 定义
├── scripts/                # 构建/部署脚本
├── go.mod
├── go.sum
└── Makefile
```

**核心理念：** 按领域/功能组织，不是按技术层（MVC 的 controller/model 就是按层分的反面例子）。

**TS 对比：** `internal/` ≈ 没有 `export` 的模块，Go 编译器强制约束。`cmd/` ≈ `src/entrypoints/`。

## 5.2 配置管理

```go
// 使用环境变量 + 配置文件，推荐 caarlos0/env 或 viper
import "github.com/caarlos0/env/v11"

type Config struct {
    Port        int           `env:"PORT" envDefault:"8080"`
    DatabaseURL string        `env:"DATABASE_URL,required"`
    JWTSecret   string        `env:"JWT_SECRET,required"`
    LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
    ReadTimeout time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
}

func LoadConfig() (Config, error) {
    cfg := Config{}
    if err := env.Parse(&cfg); err != nil {
        return cfg, err
    }
    return cfg, nil
}
```

**习惯：** Go 后端偏好环境变量，相对较少用配置文件。十二因子 App 的方式天然契合。

## 5.3 结构化日志

```go
// 标准库 log 不适合生产 — 用 log/slog（Go 1.21+）或 zerolog/zap
import "log/slog"
import "os"

func setupLogger(level string) *slog.Logger {
    var lvl slog.Level
    slog.SetLogLoggerLevel(lvl) // 或直接 NewJSONHandler
    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: lvl,
    }))
}

// 使用
slog.Info("user created", "user_id", userID, "email", email)
slog.Error("failed to create user", "error", err)

// 从 context 里取带追踪信息的 logger
func WithTrace(ctx context.Context) *slog.Logger {
    return slog.With("trace_id", ctx.Value("trace_id"))
}
```

**TS 对比：** `slog` ≈ `pino`（结构化 JSON 日志），zerolog ≈ `winston`（链式 API）。

## 5.4 测试文化

```go
// 表驱动测试 — Go 社区标准
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 5, 5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}

// 集成测试 — 用 build tag 分离
//go:build integration
// +build integration

func TestUserRepository_Integration(t *testing.T) {
    // 连接真实数据库
}

// 运行：go test -tags=integration ./...
```

### 测试辅助

```go
// testify — 断言库
import "github.com/stretchr/testify/assert"

func TestSomething(t *testing.T) {
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.Contains(t, slice, element)
}

// httptest — 测试 HTTP handler
func TestHandler(t *testing.T) {
    handler := NewHandler(store)

    req := httptest.NewRequest("GET", "/tasks", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}
```

**核心测试原则：**
- 单元测试：`go test ./...` 应在秒级完成
- 集成测试：用 build tag 分离，CI 中单独阶段
- Mock 用接口而不是反射（Go 的接口是隐式的，mock 非常自然）

## 5.5 依赖注入

Go 不需要框架做 DI — 手动构造就足够清晰。

```go
// ❌ 全局变量
var db *sql.DB
func GetUser(id int) { db.Query(...) }

// ✅ 依赖注入
type UserService struct {
    repo UserRepository
    jwt  *JWTManager
    log  *slog.Logger
}

func NewUserService(repo UserRepository, jwt *JWTManager, log *slog.Logger) *UserService {
    return &UserService{repo: repo, jwt: jwt, log: log}
}

// wire-up（在 main.go 里）
func main() {
    cfg := LoadConfig()
    log := setupLogger()
    db := connectDB(cfg)
    repo := NewUserRepo(db)
    jwt := NewJWTManager(cfg.JWTSecret)
    svc := NewUserService(repo, jwt, log)

    runServer(svc)
}
```

大型项目用 `google/wire` 做编译期 DI 生成，但手动构造在中小项目足够了。

## 5.6 优雅关闭

```go
func main() {
    // ... 启动 server

    // 监听系统信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("shutting down...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 先停 HTTP（不再接受新请求）
    server.Shutdown(ctx)
    // 再关数据库
    db.Close()

    slog.Info("server stopped")
}
```

## 5.7 认证基础

### 密码哈希

```go
import "golang.org/x/crypto/bcrypt"

hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword(hash, []byte(password))
// err == nil 表示密码正确
```

### JWT

```go
import "github.com/golang-jwt/jwt/v5"

// 创建
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "sub": userID,
    "exp": time.Now().Add(1 * time.Hour).Unix(),
    "iat": time.Now().Unix(),
})
tokenString, _ := token.SignedString([]byte(secret))

// 验证
parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return []byte(secret), nil
})
```

## 5.8 项目实战：用户认证系统

### 需求

- `POST /auth/register` — 注册（email + password，返回 tokens）
- `POST /auth/login` — 登录（返回 access + refresh token）
- `POST /auth/refresh` — 刷新令牌
- `GET /me` — 获取当前用户信息（需认证）
- Access token 短期（15min），Refresh token 长期（7 days）
- 中间件：认证、日志、请求 ID、恢复

### 模型

```sql
-- 用户表
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 刷新令牌表
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 建议结构

```
cmd/
  apiserver/
    main.go
internal/
  auth/
    handler.go      # HTTP handler
    service.go      # 业务逻辑
    jwt.go          # JWT 管理
  user/
    repository.go   # 数据库操作
    model.go        # User struct
  middleware/
    auth.go         # 认证中间件
    logging.go
    requestid.go
migrations/
config/
```

### 学习检查清单

- [ ] 能组织符合 Go 惯例的项目结构
- [ ] 掌握环境变量/配置文件管理
- [ ] 能用 slog 或 zerolog 输出结构化日志
- [ ] 掌握表驱动测试和 httptest
- [ ] 理解手动 DI 和接口松耦合
- [ ] 实现优雅关闭（signal + shutdown + drain）
- [ ] 正确实现密码哈希和 JWT
- [ ] 理解 Refresh Token 的安全考量（哈希存储、轮换）
- [ ] 能写出生产可用的 CRUD handler 模板
- [ ] 独立完成用户认证系统

### 延伸阅读

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) — 参考，并非官方标准
- [Go 1.21 slog](https://go.dev/blog/slog) — 官方结构化日志
- [OWASP JWT Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
