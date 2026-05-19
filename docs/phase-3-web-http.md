# 阶段 3：Web 基础 (net/http)

> 项目产出：RESTful 任务 API — 纯标准库，无第三方框架，json 响应，中间件链

## 3.1 标准库 HTTP 模型

Go 的 HTTP 模型由三个核心接口构成：

```go
// Handler — 一切 HTTP 处理的抽象
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// HandlerFunc — 普通函数适配为 Handler（类似 Express 的中间件签名）
type HandlerFunc func(ResponseWriter, *Request)
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }

// ServeMux — 路由器（Go 1.22+ 支持方法路由和路径参数）
mux := http.NewServeMux()
mux.HandleFunc("GET /tasks", listTasks)       // 方法 + 路径
mux.HandleFunc("POST /tasks", createTask)
mux.HandleFunc("GET /tasks/{id}", getTask)    // 路径参数
mux.HandleFunc("DELETE /tasks/{id}", deleteTask)
```

**TS 对比：** `http.Handler` ≈ Express 的 `(req, res, next) => void`，但更简洁 — 没有 next，中间件需手动包裹。

## 3.2 从路由到 HTTP 服务

```go
package main

import (
    "net/http"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    server.ListenAndServe() // 阻塞，返回 error
}
```

### 路径参数（Go 1.22+）

```go
mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
})
```

如果 Go < 1.22（企业常见），路由参数需要第三方库如 `chi` 或 `gorilla/mux`。

## 3.3 JSON 请求/响应模式

```go
// 请求解码
func createTask(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Title string `json:"title"`
    }

    // 限制请求体大小，防止内存攻击
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json")
        return
    }

    if input.Title == "" {
        writeError(w, http.StatusBadRequest, "title is required")
        return
    }

    task := Task{ID: nextID(), Title: input.Title}

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

// 统一错误响应
func writeError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

**关键细节：**
- 始终用 `json.NewEncoder(w).Encode()` 而不是 `json.Marshal` + `w.Write()`（前者直接写流，省内存）
- 请求体限制 `http.MaxBytesReader` 是必须的（否则攻击者可以发无限大的 body）
- Content-Type header 要在 WriteHeader 之前设置（`WriteHeader` 一旦调用，header 就固定了）

## 3.4 中间件模式

```go
// 中间件类型：接收 Handler，返回 Handler
type Middleware func(http.Handler) http.Handler

// 链式包装
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// 示例：请求日志
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// 示例：恢复 panic
func Recover(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v", err)
                http.Error(w, "internal server error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// 使用
handler := Chain(mux, Logging, Recover)
```

**TS 对比：** Go 中间件 `func(http.Handler) http.Handler` ≈ Express 中间件 `(req, res, next) => void`，但 Go 的方式是函数式包装，Express 是回调链。

## 3.5 使用外部路由器（chi）

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()

// 自带常用中间件
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.RequestID)
r.Use(middleware.Timeout(30 * time.Second))

r.Route("/tasks", func(r chi.Router) {
    r.Get("/", listTasks)
    r.Post("/", createTask)
    r.Route("/{taskID}", func(r chi.Router) {
        r.Get("/", getTask)
        r.Put("/", updateTask)
        r.Delete("/", deleteTask)
    })
})
```

chi 的优势：轻量、符合标准库习惯、带着工程化标配的中间件（RequestID、Timeout 等）。

## 3.6 项目实战：RESTful 任务 API

### 需求

- `GET /tasks` — 任务列表
- `POST /tasks` — 创建任务（验证 title 必填，长度限制）
- `GET /tasks/{id}` — 单个任务
- `PUT /tasks/{id}` — 更新任务
- `DELETE /tasks/{id}` — 删除任务
- 中间件：日志、恢复、请求超时
- 数据先存内存（下阶段接入数据库）

### 项目结构

```
cmd/
  tasksrv/
    main.go
internal/
  tasksrv/
    handler.go       # HTTP handler
    middleware.go    # 中间件
    task.go          # Task 模型 + 内存存储
    errors.go        # 错误定义与响应辅助
```

### 核心代码骨架

```go
// handler.go
type Server struct {
    store *taskStore
}

func (s *Server) RegisterRoutes(r *chi.Mux) {
    r.Get("/tasks", s.listTasks)
    r.Post("/tasks", s.createTask)
    r.Get("/tasks/{id}", s.getTask)
    r.Put("/tasks/{id}", s.updateTask)
    r.Delete("/tasks/{id}", s.deleteTask)
}
```

```go
// main.go
func main() {
    store := NewTaskStore()
    srv := &Server{store: store}

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    srv.RegisterRoutes(r)

    http.ListenAndServe(":8080", r)
}
```

### 关键点

- JSON 序列化/反序列化只用 `json.NewEncoder` / `json.NewDecoder`
- 错误响应统一格式 `{"error": "message"}`
- `w.WriteHeader` 只能调用一次
- 所有 handler 里不要 panic，返回 error 让业务上层决策

### 学习检查清单

- [ ] 理解 `http.Handler` 接口和 `HandlerFunc` 适配器
- [ ] 能写出正确的请求绑定（JSON → struct）和响应编码（struct → JSON）
- [ ] 掌握中间件的包装模式
- [ ] 能正确处理 HTTP 状态码和 Content-Type
- [ ] 了解 `r.Body.Close()` 的时机和必要性（标准库自动处理，但手动读 body 要小心）
- [ ] 能解析路径参数（Go 1.22+ 或用 chi）
- [ ] 理解 `http.Server` 的字段和 graceful shutdown
- [ ] 独立完成 RESTful 任务 API（内存版）

### 延伸阅读

- [Go 1.22 Routing Enhancements](https://go.dev/blog/routing-enhancements) — 标准库路由升级
- [chi 文档](https://go-chi.io/) — 轻量路由器
- [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
