# 阶段 2：并发与标准库

> 项目产出：文件批量处理器 — 并发扫描目录、分类整理、生成报告

## 2.1 Goroutine — 轻量级协程

```go
// 启动一个 goroutine
go func() {
    fmt.Println("运行在另一个 goroutine")
}()

// goroutine 不是 OS 线程，是 Go 运行时管理的用户态协程
// 一个 Go 程序启动时，main 函数运行在主 goroutine
// goroutine 极轻量（~2KB 栈起步），可以轻松起上万个
```

**TS 对比：**

| Go | JavaScript/TypeScript |
|----|----------------------|
| `go fn()` | `Promise.resolve().then(fn)` 或 `setImmediate(fn)` |
| goroutine 是抢占式，多核并行 | JS 是单线程事件循环（不考虑 Worker） |
| 一个 goroutine 阻塞不影响调度 | 一个 Promise 的长时间计算会阻塞主线程 |

## 2.2 Channel — goroutine 间的通信管道

```go
// 创建
ch := make(chan int)       // 无缓冲 — 同步
ch := make(chan int, 10)   // 有缓冲 — 异步（满时阻塞）

// 发送
ch <- 42

// 接收
value := <-ch

// 关闭（发送方关闭，接收方检查）
close(ch)
value, ok := <-ch  // ok == false 则 channel 已关闭且已空

// 用 range 消费 channel，直到关闭
for v := range ch {
    fmt.Println(v)
}
```

**核心哲学：** "Don't communicate by sharing memory; share memory by communicating."

```go
// ❌ 共享内存 + 锁（不 Go）
func bad() {
    var mu sync.Mutex
    var counter int
    for i := 0; i < 1000; i++ {
        go func() {
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
}

// ✅ 用 channel 通信
func good() {
    ch := make(chan int)
    for i := 0; i < 1000; i++ {
        go func() { ch <- 1 }()
    }
    counter := 0
    for i := 0; i < 1000; i++ {
        counter += <-ch
    }
}
```

### Select — 多 channel 选择器

```go
select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case ch2 <- data:
    fmt.Println("sent to ch2")
case <-time.After(1 * time.Second):
    fmt.Println("timeout")
default:
    fmt.Println("no one ready")
}
```

select 类比 JS 的 `Promise.race()`，但有更多控制。

问题：上面的 file processor 示例里用 channel 只是把锁换了个形式。真正吃场景的是 pipeline、fan-out/fan-in — 阶段 6 会展开。

## 2.3 常用标准库

### sync 包 — 同步原语

```go
// WaitGroup — 等待一组 goroutine 完成
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        process(n)
    }(i)
}
wg.Wait()

// Mutex — 互斥锁
var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()

// RWMutex — 读写锁（读多写少场景）
var rw sync.RWMutex
rw.RLock()   // 多个读可同时持有
rw.RUnlock()
rw.Lock()    // 写互斥
rw.Unlock()

// Once — 只执行一次
var once sync.Once
once.Do(func() { setup() })
```

### 文件操作 — os + path/filepath + io

```go
// 读取整个文件
data, err := os.ReadFile("input.txt")

// 写入文件
err := os.WriteFile("output.txt", data, 0644)

// 流式读取
f, _ := os.Open("large.txt")
defer f.Close()
buf := make([]byte, 4096)
for {
    n, err := f.Read(buf)
    if err == io.EOF { break }
}
```

### JSON 编码/解码

```go
// 结构体 → JSON
data, err := json.Marshal(obj)
data, err := json.MarshalIndent(obj, "", "  ")

// JSON → 结构体
err := json.Unmarshal(data, &obj)

// JSON → io.Writer（避免中间[]byte）
encoder := json.NewEncoder(w)
err := encoder.Encode(obj)

// io.Reader → 结构体（避免中间[]byte）
decoder := json.NewDecoder(r)
err := decoder.Decode(&obj)
```

### context — 传递取消信号和超时

```go
// 这是 Go 里贯穿所有异步/IO 的"公共基础设施"
ctx := context.Background()    // 根 context

// 超时
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// 手动取消
ctx, cancel := context.WithCancel(ctx)

// 传值（仅用于请求级元数据，不要传业务参数）
ctx := context.WithValue(ctx, key, value)
```

## 2.4 项目实战：文件批量处理器

### 需求

- 递归扫描指定目录下的所有文件
- 按文件类型/扩展名分类，移动到对应子目录
- 并发处理，可控制并发数
- 生成处理报告（文件数、大小、耗时）

### 建议结构

```
cmd/
  fileproc/
    main.go
internal/
  fileproc/
    scanner.go      # 扫描文件
    processor.go    # 处理/移动文件
    reporter.go     # 生成报告
    scanner_test.go
```

### 核心模式

```go
// worker pool 模式 — 固定数量的 goroutine 消费任务
type Job struct {
    Path string
    Size int64
}

func processFiles(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        // 处理 job ...
        results <- result
    }
}

// 启动 N 个 worker
const numWorkers = 8
jobs := make(chan Job, 100)
results := make(chan Result, 100)
var wg sync.WaitGroup

for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go processFiles(jobs, results, &wg)
}
```

### 学习检查清单

- [ ] 能用 goroutine + channel 完成并发任务
- [ ] 理解 select 的多路复用场景
- [ ] 掌握 WaitGroup、Mutex 的使用
- [ ] 会用 io.Reader / io.Writer 抽象做流式处理
- [ ] 能用 context 管理超时和取消
- [ ] 理解 Go 错误处理的惯例（不 panic、包装错误链）
- [ ] 能写并发安全的代码
- [ ] 独立完成文件批量处理器

### 延伸阅读

- [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide) — Rob Pike 的经典演讲
- [Go Memory Model](https://go.dev/ref/mem) — 了解 happens-before，至少知道什么不能乱
- [The Go Blog: Pipelines and cancellation](https://go.dev/blog/pipelines)
