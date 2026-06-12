# 第 2 课：类型、函数、错误处理

## 学习目标

完成这一课后，你应该能够：

- 使用 `struct` 表达输入数据和业务数据。
- 理解函数参数、返回值和多返回值。
- 使用 `error` 显式表达失败路径。
- 使用 `strings` 包处理常见字符串规则。
- 编写表格驱动测试覆盖成功和失败场景。

## 前端架构师视角

前端/TypeScript 项目里，你可能习惯用 `type` 或 `interface` 描述数据形状，再用运行时校验库处理用户输入。Go 的 `struct` 同时承担一部分“数据形状”和“运行期数据结构”的职责，但它不会自动校验字段。校验规则仍然需要你明确写出来。

Go 的错误处理也和 JavaScript/TypeScript 很不一样。Go 通常不会用 `throw/catch` 组织常规业务错误，而是把错误作为最后一个返回值显式返回：

```go
profile, err := BuildProfile(input)
if err != nil {
	return err
}
```

这会让失败路径更啰嗦，但也更可见。后端服务里，输入错误、权限错误、数据库错误和系统错误都应该被清楚地建模，而不是被一个全局异常处理器模糊吞掉。

## 核心概念

### struct

`struct` 是字段的集合，适合表达明确的数据结构：

```go
type SignupInput struct {
	Name  string
	Email string
	Age   int
}
```

字段名首字母大写表示可被其他 package 访问。当前 demo 只有一个 package，但课程从一开始就使用导出字段，方便后续过渡到 JSON API。

### 多返回值

Go 函数可以返回多个值。常见约定是最后一个返回值为 `error`：

```go
func BuildProfile(input SignupInput) (UserProfile, error)
```

成功时返回业务结果和 `nil` 错误；失败时返回零值结果和非空错误。

### error

`error` 是一个接口：

```go
type error interface {
	Error() string
}
```

本课使用 `errors.New` 定义哨兵错误，再用 `fmt.Errorf("%w: ...", err)` 包装上下文。这样调用方可以用 `errors.Is` 判断错误类别。

### 表格驱动测试

表格驱动测试把多个输入输出 case 放在一个切片里，减少重复代码，并让测试覆盖范围更清晰。

## 代码结构

```text
examples/lesson02-types-functions-and-errors/
  go.mod
  README.md
  main.go
  profile.go
  profile_test.go
```

- `profile.go`：定义输入类型、输出类型和校验函数。
- `profile_test.go`：测试成功路径和多个错误路径。
- `main.go`：构造一个输入并打印结果。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson02-types-functions-and-errors
```

运行程序：

```bash
go run .
```

运行测试：

```bash
go test ./...
```

## 关键代码讲解

输入类型和输出类型分开：

```go
type SignupInput struct {
	Name  string
	Email string
	Age   int
}

type UserProfile struct {
	DisplayName string
	Email       string
	IsAdult     bool
}
```

这和前端表单模型、后端领域模型分开是同一个原则：输入数据通常不等于系统内部真正要使用的数据。

构建函数返回两个值：

```go
func BuildProfile(input SignupInput) (UserProfile, error)
```

校验失败时返回零值和错误：

```go
if name == "" {
	return UserProfile{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
}
```

调用方不需要猜测函数是否会抛异常，只需要检查 `err`：

```go
profile, err := BuildProfile(input)
if err != nil {
	fmt.Println("build profile failed:", err)
	return
}
```

测试里使用 `errors.Is` 判断错误类别，而不是完全依赖错误字符串：

```go
if !errors.Is(err, ErrInvalidInput) {
	t.Fatalf("want ErrInvalidInput, got %v", err)
}
```

错误字符串可以变化，但错误类别应该稳定。

## 练习

1. 给 `SignupInput` 增加 `Role string` 字段，只允许 `frontend`、`backend`、`fullstack`。
2. 把最低年龄限制从 13 改成 18，并更新测试。
3. 新增一个 `NormalizeEmail(email string) string` 函数，把邮箱统一转成小写并写测试。

## 常见坑

- Go 不会自动 trim 字符串，用户输入进入业务逻辑前要显式处理。
- `err != nil` 是 Go 中非常常见的控制流，不要为了少写几行而隐藏它。
- 不要只测试成功路径，输入校验最需要覆盖失败路径。
- `struct` 字段首字母小写时，跨 package 和 JSON 编解码都会受到影响。
- 错误字符串适合给开发者排障，不适合作为稳定的业务判断依据。

## 下一课

下一课会讲方法、接口和包设计。你会把函数挂到类型上，理解 Go 的接口为什么更像“能力约束”，并开始拆分 package 边界。
