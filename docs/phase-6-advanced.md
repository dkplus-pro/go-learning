# 阶段 6：进阶服务

> 项目产出：实时通知服务 — WebSocket 实时推送 + Redis 消息队列 + 缓存层

## 6.1 WebSocket

```bash
go get github.com/coder/websocket       # 推荐（现代化，context 原生支持）
# 或
go get github.com/gorilla/websocket      # 老牌（社区最常用）
```

```go
// 用 gorilla/websocket
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // 生产要限制
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    // 从认证信息获取用户 ID
    userID := r.Context().Value("user_id").(string)

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("upgrade failed", "error", err)
        return
    }
    defer conn.Close()

    hub.Register(userID, conn)
    defer hub.Unregister(userID)

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil { break }
        // 处理消息...
    }
}
```

### Hub 模式 — 连接管理

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]*websocket.Conn // userID -> conn
}

func (h *Hub) Register(userID string, conn *websocket.Conn) { }
func (h *Hub) Unregister(userID string) { }
func (h *Hub) Send(userID string, msg []byte) error { }
func (h *Hub) Broadcast(msg []byte) { } // 全体推送
```

## 6.2 消息队列 — Redis Pub/Sub

```bash
go get github.com/redis/go-redis/v9
```

```go
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 发布
ctx := context.Background()
client.Publish(ctx, "notifications", jsonData)

// 订阅
sub := client.Subscribe(ctx, "notifications")
defer sub.Close()

ch := sub.Channel()
for msg := range ch {
    var notif Notification
    json.Unmarshal([]byte(msg.Payload), &notif)
    // 通过 Hub 推送给目标用户
    hub.Send(notif.UserID, msg.Payload)
}
```

### 发布/订阅架构

```
[HTTP API]                    [WebSocket Server]
    |                                ^
    | POST /notify                   | (conn)
    v                                |
[Redis Pub/Sub] -----------> [Subscriber goroutine]
                                     |
                                     v
                                  [Hub] --> [User's WebSocket]
```

**为什么要队列：** 通知服务可能独立于主 API 部署。Redis Pub/Sub 解耦了发送者和推送者，也支持横向扩容（多个 WebSocket 实例共享同一个 Redis channel）。

**TS 对比：** Redis Pub/Sub ≈ 简单的 EventEmitter 跨进程版，适用实时通知。复杂消息用 RabbitMQ/NATS（持久化、死信、重试）。

## 6.3 缓存

### 本地缓存（单机）

```go
// go-cache — 简单 KV 缓存，有过期时间
import "github.com/patrickmn/go-cache"

c := cache.New(5*time.Minute, 10*time.Minute)
c.Set("key", value, cache.DefaultExpiration)
if val, found := c.Get("key"); found { }
```

### Redis 缓存（分布式）

```go
// 缓存模式：Cache-Aside
func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
    // 1. 查缓存
    key := fmt.Sprintf("user:%s", id)
    data, err := s.redis.Get(ctx, key).Bytes()
    if err == nil {
        var user User
        json.Unmarshal(data, &user)
        return user, nil
    }
    if err != redis.Nil {
        slog.Error("redis error", "error", err) // 降级继续查 DB
    }

    // 2. 查数据库
    user, err := s.repo.FindByID(ctx, id)
    if err != nil { return User{}, err }

    // 3. 写缓存
    encoded, _ := json.Marshal(user)
    s.redis.Set(ctx, key, encoded, 10*time.Minute)

    return user, nil
}
```

### 缓存策略要点

| 模式 | 说明 |
|------|------|
| Cache-Aside | 应用手动管理缓存，最灵活 |
| Write-Through | 写 DB 同时更新缓存 |
| Cache-Aside + TTL | 最适合大部分业务场景 |

**陷阱：** 缓存穿透（大量不存在的 key 打到 DB）、缓存雪崩（大量 key 同时过期）、缓存击穿（热点 key 过期瞬间高并发）。用 singleflight + 空值缓存解决。

```go
import "golang.org/x/sync/singleflight"

var g singleflight.Group

func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
    v, err, _ := g.Do(id, func() (interface{}, error) {
        return s.loadUser(ctx, id) // 同一 key 只执行一次
    })
    return v.(User), err
}
```

## 6.4 可观测性

### Metrics（Prometheus + Grafana）

```bash
go get github.com/prometheus/client_golang
```

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total"},
        []string{"method", "path", "status"},
    )
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "http_request_duration_seconds"},
        []string{"method", "path"},
    )
)

// 在中间件里埋点
func Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(r.Method, r.URL.Path))
        defer timer.ObserveDuration()

        // 包装 ResponseWriter 获取状态码
        wr := &responseRecorder{ResponseWriter: w, statusCode: 200}
        next.ServeHTTP(wr, r)

        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wr.statusCode)).Inc()
    })
}
```

### 分布式追踪

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

```go
// 给 HTTP client 自动注入 trace
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}
```

**TS 对比：** OpenTelemetry 是跨语言标准。Go 的 otel 和 TS 的 `@opentelemetry/sdk-node` 同一套协议。

## 6.5 项目实战：实时通知服务

### 需求

- 用户可以通过 WebSocket 建立连接（带 JWT 认证）
- REST API 发送通知给指定用户
- 通知通过 Redis Pub/Sub 传递到 WebSocket 服务
- 缓存用户在线状态
- Prometheus metrics 端点暴露（`/metrics`）

### 架构

```
客户端 1 ───> WS ─> Hub ─> Redis Sub
                                ^
客户端 2 ───> POST /notify ─> Redis Pub
                                |
客户端 3 ───> WS ─────────────> Redis Sub
```

### 建议结构

```
cmd/
  notifysrv/
    main.go
internal/
  notifysrv/
    handler.go         # HTTP handler（发送通知 + WS upgrade）
    hub.go             # WebSocket 连接管理
    subscriber.go      # Redis 订阅消费者
    metrics.go         # Prometheus metrics
```

### 学习检查清单

- [ ] 掌握 WebSocket upgrade 和连接生命周期
- [ ] 能实现 Hub 模式管理多连接
- [ ] 理解 Redis Pub/Sub 的发布订阅模型
- [ ] 能实现 Cache-Aside 缓存模式
- [ ] 理解缓存穿透、雪崩、击穿及应对
- [ ] 能用 singleflight 合并并发请求
- [ ] 能用 Prometheus client 暴露 metrics
- [ ] 了解 OpenTelemetry 的 trace/span 概念
- [ ] 独立完成实时通知服务

### 延伸阅读

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Redis Pub/Sub](https://redis.io/docs/latest/develop/interact/pubsub/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Go singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)
