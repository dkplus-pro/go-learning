# 第 1 课：Hello Go 与最小工程

## 学习目标

完成这一课后，你应该能够：

- 说清楚一个最小 Go 程序由哪些文件组成。
- 理解 `go.mod`、`package main`、`func main` 的职责。
- 使用 `go run .` 运行当前目录下的 Go 程序。
- 使用 `go test ./...` 运行当前 module 下的测试。
- 把可测试逻辑从 `main` 函数中拆出来。

## 前端架构师视角

如果你熟悉前端项目，可以先把 Go module 粗略类比成一个 npm package。`go.mod` 类似 `package.json` 里的包名、语言版本和依赖声明，但 Go 的依赖解析、构建和测试命令由官方工具链统一提供，不需要额外选择 webpack、vite、jest 这一类基础工具。

`package main` 表示这个包会编译成可执行程序。`func main()` 是程序入口，类似一个 Node.js CLI 程序的入口文件，但 Go 会先编译再运行，类型检查和构建流程比脚本式运行更严格。

本课刻意不引入框架，也不引入 HTTP。第一步要先建立后端开发最基本的节奏：写可运行代码，拆出可测试函数，然后用命令验证它。

## 核心概念

### Go module

`go.mod` 定义当前 demo 的 module 路径和 Go 版本：

```go
module github.com/dkplus-pro/go-learning/examples/lesson01-hello-go

go 1.24
```

当前课程约定每一课都是独立 module。这样你可以进入任意课程目录，单独运行、测试和阅读，不会被其他课程影响。

### package main

Go 文件开头必须声明 package。`package main` 是一个特殊包名，它告诉 Go 工具链：这个包可以构建成可执行程序。

### func main

`func main()` 是可执行程序入口。它不接收参数，也不返回值。真实项目里，`main` 通常只做启动编排，比如读取配置、连接依赖、启动 HTTP 服务；可测试的业务逻辑应该放到普通函数或 package 中。

### 测试文件

Go 测试文件以 `_test.go` 结尾，测试函数以 `Test` 开头，并接收 `*testing.T`。Go 官方工具链会自动发现这些测试。

## 代码结构

```text
examples/lesson01-hello-go/
  go.mod
  README.md
  greeting.go
  greeting_test.go
  main.go
```

- `main.go`：程序入口，负责调用函数并输出结果。
- `greeting.go`：可测试的问候语逻辑。
- `greeting_test.go`：覆盖问候语逻辑的单元测试。
- `README.md`：demo 的最短运行说明。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson01-hello-go
```

运行程序：

```bash
go run .
```

预期输出：

```text
Hello, Go backend learner!
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

`main.go` 只保留入口职责：

```go
func main() {
	fmt.Println(Greet("Go backend learner"))
}
```

`Greet` 函数放在 `greeting.go` 中：

```go
func Greet(name string) string {
	if name == "" {
		name = "friend"
	}

	return "Hello, " + name + "!"
}
```

这里有一个很重要的后端工程习惯：不要把所有逻辑都写进 `main`。`main` 很难直接测试，而普通函数可以被单元测试稳定覆盖。

测试文件使用表格驱动测试：

```go
tests := []struct {
	name string
	in   string
	want string
}{
	{name: "with name", in: "Go", want: "Hello, Go!"},
	{name: "empty name", in: "", want: "Hello, friend!"},
}
```

表格驱动测试是 Go 社区常见写法。它和前端测试里的参数化 case 类似，适合表达“同一个函数在多种输入下应该得到什么输出”。

## 练习

1. 修改 `main.go`，让它输出你的名字。
2. 给 `Greet` 增加一个新的测试用例，例如输入 `"Frontend Architect"`。
3. 新增一个 `Shout(name string) string` 函数，返回大写问候语，并为它写测试。

## 常见坑

- `go run .` 要在包含 `go.mod` 的 demo 目录里运行，不是在仓库根目录运行。
- `func main` 只能出现在 `package main` 中。
- 测试文件必须以 `_test.go` 结尾，否则 `go test` 不会把它当测试文件。
- 测试函数必须形如 `func TestXxx(t *testing.T)`。
- Go 字符串拼接可以用 `+`，但复杂格式化时应该使用 `fmt.Sprintf`。

## 下一课

下一课会进入 Go 的类型、函数和错误处理。你会看到 Go 如何表达数据结构，为什么后端代码倾向显式返回错误，以及如何用表格驱动测试覆盖输入校验逻辑。
