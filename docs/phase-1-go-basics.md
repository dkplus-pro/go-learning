# 阶段 1：Go 语言基础

> 项目产出：命令行 Todo 管理工具（增删改查、持久化到文件）

## 1.1 环境准备

```bash
# 安装 Go（推荐 1.23+）
brew install go

# 验证
go version

# 在本仓库初始化模块
go mod init github.com/<your-username>/go-learning
```

**Go 工具链概览：**

| 命令            | 作用     | 类比 npm/pnpm       |
| ------------- | ------ | ----------------- |
| `go mod init` | 初始化模块  | `pnpm init`       |
| `go mod tidy` | 整理依赖   | `pnpm install`    |
| `go build`    | 编译为二进制 | `tsc` + `esbuild` |
| `go run`      | 编译并运行  | `npx tsx`         |
| `go fmt`      | 格式化代码  | `prettier`        |
| `go vet`      | 静态分析   | `eslint`          |
| `go test`     | 运行测试   | `vitest`          |

## 1.2 核心语法速览

### 变量与类型

```go
// 显式声明
var name string = "Go"

// 短声明（最常用，仅函数内可用）
age := 10

// 常量
const MaxRetry = 3

// 基本类型
// bool, string, int, int8, int16, int32, int64,
// uint, uint8, ..., float32, float64,
// complex64, complex128, byte (uint8 别名), rune (int32 别名)
```

**TS 对比：** `const x = 1` 在 TS 里是常量引用，在 Go 里 `x := 1` 是变量，`const x = 1` 才是编译期常量。

### 函数

```go
// 签名：func 函数名(参数) 返回值
func add(a, b int) int {
    return a + b
}

// 多返回值（Go 的招牌特性）
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// 命名返回值
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return // naked return，只在短函数里用
}
```

**TS 对比：** `const [data, err] = await fetch()` 的模式灵感就来自 Go 的 `result, err := doSomething()`。

### 控制流

```go
// for 是唯一的循环关键字（没有 while, do-while）
for i := 0; i < 10; i++ { }

for condition { }     // 当 while 用

for { }               // 无限循环

// if 可带短声明
if err := doSomething(); err != nil {
    return err
}

// switch 不需要 break（默认 break，fallthrough 显式穿透）
switch os := runtime.GOOS; os {
case "darwin":
    fmt.Println("macOS")
default:
    fmt.Println(os)
}

// 无条件的 switch 替代 else-if
switch {
case score >= 90:
    grade = "A"
case score >= 80:
    grade = "B"
default:
    grade = "C"
}
```

### 复合类型

```go
// 数组 — 定长，编译期确定
arr := [3]int{1, 2, 3}

// 切片 — 动态数组，99% 时间用这个
s := []int{1, 2, 3}
s = append(s, 4)        // append 返回新切片
fmt.Println(s[1:3])     // 切片操作 => [2, 3]
fmt.Println(len(s))     // 长度
fmt.Println(cap(s))     // 容量

// map — 无序哈希表
m := map[string]int{"a": 1, "b": 2}
v, ok := m["c"]         // 两个值：值 + 是否存在
delete(m, "a")

// struct — 数据载体
type Person struct {
    Name string
    Age  int
}
p := Person{Name: "Alice", Age: 30}

// 循环：range 关键字
for i, v := range s { }     // 切片：索引 + 值
for k, v := range m { }     // map：键 + 值
```

### 方法 vs 函数

```go
// 普通函数
func (p Person) Greet() string {
    return "Hello, " + p.Name
}

// 值接收者 — 不会修改原值
func (p Person) SetName(name string) {
    p.Name = name // 无效！改的是副本
}

// 指针接收者 — 会修改原值
func (p *Person) SetName(name string) {
    p.Name = name // 有效
}
```

**TS 对比：** 方法定义在 struct 外部，挂载方式类似 `Person.prototype.greet = function() {}`。指针接收者 ≈ 方法里拿到了对象引用可以改原对象。

### 接口 — 隐式实现

```go
// 定义接口（用消费者视角命名）
type Reader interface {
    Read(p []byte) (n int, err error)
}

// 任何有 Read([]byte) (int, error) 方法的类型都自动实现了 Reader
// 不需要关键字声明 implements
type MyReader struct{}

func (m MyReader) Read(p []byte) (int, error) {
    // 实现细节
}

// 使用
var r Reader = MyReader{}  // 隐式接口，编译时检查
```

**TS 对比：** Go 接口是 structural typing，类似 TypeScript 的 structural type system — "鸭子类型" 在编译期就被检查了。TS 里 `interface X { read(buf: Uint8Array): number }` 也是结构化匹配，思想一致。

### 包与可见性

```go
// 大写字母开头 = 导出（exported），类似 TS 的 export
// 小写字母开头 = 包内私有
package todo

func AddTask(title string) Task { }  // 导出
func validate(id int) error { }      // 私有
```

## 1.3 错误处理（早适应，早解脱）

```go
// Go 的 err != nil 是语言文化，不是权宜之计
data, err := os.ReadFile("data.json")
if err != nil {
    return fmt.Errorf("read config: %w", err)  // %w 包装错误链
}

// error 是接口，只有一个方法
type error interface {
    Error() string
}

// 你可以用 errors.New() 或 fmt.Errorf() 创建 error
```

不要试图用 panic/recover 替代 error — panic 是真正不可恢复的场景用的（类似服务器启动失败），不用于业务流程。

## 1.4 项目实战：命令行 Todo 工具

### 需求

- 支持 `add` / `list` / `done` / `delete` 子命令
- 数据持久化到 `~/.todos.json`
- 使用标准库 `os`、`encoding/json`、`flag` 或 `os.Args`
- 良好的错误处理

### 项目结构

```
.
├── go.mod
└── cmd/
    └── todo/
        └── main.go
└── internal/
    └── todo/
        ├── todo.go       # Task 定义 + Manager
        └── todo_test.go  # 单元测试
```

### 核心代码骨架

**Task 模型：**

```go
package todo

import "time"

type Task struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Done      bool      `json:"done"`
    CreatedAt time.Time `json:"created_at"`
}
```

**Manager：**

```go
package todo

type Manager struct {
    filePath string
    tasks    []Task
    nextID   int
}

func NewManager(filePath string) (*Manager, error) {
    m := &Manager{filePath: filePath, nextID: 1}
    if err := m.load(); err != nil {
        return nil, err
    }
    return m, nil
}

func (m *Manager) Add(title string) error { /* ... */ }
func (m *Manager) List() []Task            { /* ... */ }
func (m *Manager) Done(id int) error       { /* ... */ }
func (m *Manager) Delete(id int) error     { /* ... */ }
```

### 关键点

- 使用 `os.ReadFile` / `os.WriteFile` 读写文件
- 使用 `json.Marshal` / `json.Unmarshal` 序列化
- 使用 struct tag `` `json:"field_name"` `` 控制字段映射
- 单元测试用 `_test.go` 后缀，`go test ./...` 运行

### 学习检查清单

- [ ] 能用 `:=` 和 `var` 正确声明变量，理解区别
- [ ] 能写出带错误返回值的函数
- [ ] 能用 `for range` 遍历切片和 map
- [ ] 理解值类型 vs 指针类型的区别
- [ ] 能定义 struct 和方法
- [ ] 能定义和使用接口
- [ ] 理解包的导出规则（大小写）
- [ ] 能写表驱动测试（table-driven tests）
- [ ] 独立完成 Todo 命令行工具

### 延伸阅读

- [A Tour of Go](https://go.dev/tour/) — Go 官方交互式教程
- [Effective Go](https://go.dev/doc/effective_go) — Go 编程风格指南
- [Go by Example](https://gobyexample.com/) — 常见模式速查

