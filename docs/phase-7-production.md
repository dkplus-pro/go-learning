# 阶段 7：生产就绪

> 项目产出：将认证系统微服务化 — Docker 多阶段构建 + CI/CD + 性能调优

## 7.1 Docker 多阶段构建

```dockerfile
# 阶段 1：编译
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /apiserver ./cmd/apiserver

# 阶段 2：最小运行时
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /apiserver /usr/local/bin/apiserver
USER 1000:1000
EXPOSE 8080
ENTRYPOINT ["apiserver"]
```

**关键点：**
- `CGO_ENABLED=0` 编译出纯静态二进制，不需要 glibc
- `-ldflags="-w -s"` 去掉调试信息，减小体积
- `USER 1000` 不以 root 运行
- 最终镜像 ~15MB

## 7.2 配置与环境

```go
// 十二因子 App 风格 — 配置从环境变量来
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Auth     AuthConfig
}

type DatabaseConfig struct {
    Host     string `env:"DB_HOST,required"`
    Port     int    `env:"DB_PORT" envDefault:"5432"`
    Name     string `env:"DB_NAME,required"`
    User     string `env:"DB_USER,required"`
    Password string `env:"DB_PASSWORD,required"`

    MaxOpenConns    int           `env:"DB_MAX_OPEN" envDefault:"25"`
    MaxIdleConns    int           `env:"DB_MAX_IDLE" envDefault:"10"`
    ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFE" envDefault:"5m"`
}

func (c DatabaseConfig) DSN() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        c.User, c.Password, c.Host, c.Port, c.Name)
}
```

### docker-compose 本地开发

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: myapp
      POSTGRES_PASSWORD: devonly
    ports: ["5432:5432"]
    volumes: ["pgdata:/var/lib/postgresql/data"]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  app:
    build: .
    ports: ["8080:8080"]
    environment:
      DB_HOST: db
      DB_NAME: myapp
      DB_USER: myapp
      DB_PASSWORD: devonly
      REDIS_ADDR: redis:6379
    depends_on: [db, redis]
```

## 7.3 性能调优

### pprof — Go 内置性能分析器

```go
import _ "net/http/pprof"

// 在 main.go 里加
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
# 跑一次 30 秒的 CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 分析内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 分析 goroutine 泄漏
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 火焰图
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30
```

### 压测工具

```bash
# vegeta — 用 Go 写的 HTTP 压测
echo "GET http://localhost:8080/tasks" | vegeta attack -duration=30s -rate=1000 | vegeta report

# wrk — 高性能
wrk -t4 -c100 -d30s http://localhost:8080/tasks
```

### 常见性能陷阱

| 陷阱 | 说明 | 修复 |
|------|------|------|
| 循环中的字符串拼接 | `+=` 每次都分配新内存 | 用 `strings.Builder` |
| 未预分配切片 | `append` 多次触发扩容 | `make([]T, 0, capacity)` |
| defer 在循环中 | defer 在函数结束才执行 | 放到循环外的函数里 |
| 大 struct 作为函数参数 | 值拷贝开销大 | 传指针 |
| channel 未关闭 | goroutine 泄漏（永远阻塞） | 确保发送方关闭或接收方有退出条件 |
| sync.Mutex 复制 | Lock/Unlock 失效 | `go vet` 能检测 |

### 并发安全清单

```go
// 全局变量 → 用 sync.Once
var instance *Service
var once sync.Once

func GetService() *Service {
    once.Do(func() { instance = &Service{} })
    return instance
}

// map 并发读写 → sync.Map 或 mutex + map
// 但 sync.Map 只适合读多写少的特殊场景，大部分用 map + RWMutex

// goroutine 泄漏 → 永远确保 goroutine 有退出路径
// 用 context.WithCancel 或 <-done channel
```

## 7.4 CI/CD

### GitHub Actions 示例

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env: { POSTGRES_PASSWORD: test }
        ports: ["5432:5432"]

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.23" }

      - run: go mod download
      - run: go vet ./...
      - run: go test -race -count=1 ./...

      - run: go test -tags=integration -count=1 ./...
        env: { DATABASE_URL: "postgres://postgres:test@localhost/test?sslmode=disable" }

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6
```

**关键：** `-race` 启用数据竞争检测（生产 CI 必须开）；`-count=1` 禁用测试缓存。

### 部署策略

| 策略 | 工具 | 适用 |
|------|------|------|
| 单机部署 | systemd + rsync | 小项目 |
| 容器化 | Docker + docker-compose | 中型项目 |
| 编排 | Kubernetes + Helm | 企业级 |
| Serverless | Google Cloud Run / AWS Lambda | 事件驱动 |

Go 二进制是静态编译的单个文件，部署极其简单 — 拷贝到服务器直接运行。

## 7.5 安全清单

```go
// ✅ 始终做
r.Body = http.MaxBytesReader(w, r.Body, 1<<20)  // 限制请求体
db.Query("SELECT ... WHERE id = $1", id)          // 参数化查询
bcrypt.GenerateFromPassword([]byte(pw), 12)        // bcrypt cost >= 12
token := jwt.SigningMethodES256                     // JWT 用非对称算法

// ❌ 永远不要
db.Query("SELECT ... WHERE id = " + id)             // SQL 注入
fmt.Sprintf("rm -rf %s", userInput)                 // 命令注入
template.HTML(userInput)                            // XSS（如果渲染 HTML）
```

## 7.6 项目实战：微服务化改造

### 目标

将阶段 5 的认证系统拆成两个服务，并通过 Docker Compose 编排部署。

### 服务拆分

```
gateway (HTTP)         ──> auth-service (gRPC)
   │                           │
   │ REST API                  │ DB + Redis
   │ JWT 验证                  │
   │ 路由转发                  │
```

### 建议结构

```
cmd/
  gateway/
    main.go
  authsvc/
    main.go
internal/
  gateway/
    handler.go       # 对外 REST API
  authsvc/
    handler.go       # gRPC handler
    service.go       # 认证逻辑
api/
  auth/v1/
    auth.proto       # protobuf 定义
deploy/
  docker-compose.yml
  Dockerfile.gateway
  Dockerfile.authsvc
```

### gRPC 快速入门

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

```protobuf
// api/auth/v1/auth.proto
service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}
```

```go
// gateway 里调用 auth-service
conn, _ := grpc.Dial("authsvc:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewAuthServiceClient(conn)

resp, err := client.Login(ctx, &pb.LoginRequest{Email: email, Password: password})
```

### 学习检查清单

- [ ] 能写 Docker 多阶段构建文件
- [ ] 能用 docker-compose 编排本地开发环境
- [ ] 会用 pprof 分析 CPU/内存/goroutine
- [ ] 能使用压测工具发现性能瓶颈
- [ ] 了解 Go 常见性能陷阱和修复方法
- [ ] 能搭建 CI/CD pipeline（lint + test + build）
- [ ] 理解十二因子 App 原则
- [ ] 掌握 Go 安全编码基础
- [ ] 了解 gRPC 和 protobuf 基础
- [ ] 独立完成微服务化改造

### 延伸阅读

- [Docker 官方 Go 指南](https://docs.docker.com/language/golang/)
- [Go pprof 文档](https://pkg.go.dev/runtime/pprof)
- [gRPC Go Quickstart](https://grpc.io/docs/languages/go/quickstart/)
- [Go Security Checklist](https://github.com/Checkmarx/Go-SCP)
