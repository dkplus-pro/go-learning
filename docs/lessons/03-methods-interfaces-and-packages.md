# 第 3 课：方法、接口与包设计

## 学习目标

完成这一课后，你应该能够：

- 给 `struct` 定义方法。
- 区分值接收者和指针接收者。
- 理解 Go interface 是能力约束，不是数据形状声明。
- 使用小接口隔离调用方和实现方。
- 把 demo 拆成 `main` 包和业务包。

## 前端架构师视角

如果你长期写 TypeScript，很容易把 Go 的 `interface` 理解成“对象 shape 声明”。这只对了一小部分。Go 的 interface 更常用于表达“调用方需要的能力”，例如“我需要一个可以保存学习者的东西”，而不是“我要声明一个完整对象结构”。

Go 也没有 class 继承。你不会写一个庞大的 BaseService，然后让各种业务类继承它。Go 更偏向组合：数据用 `struct` 表达，行为用方法表达，依赖边界用小接口表达。

本课的 demo 会引入一个 `internal/learning` 包。它不是复杂架构，只是为了让你从第 3 课开始习惯：入口代码和业务逻辑应该分开。

## 核心概念

### 方法

方法是带接收者的函数：

```go
func (l Learner) Level() string
```

`Learner` 是接收者。你可以把它理解成“这个函数属于这个类型的行为”，但它不是 class 方法。

### 值接收者

值接收者会拿到一份值拷贝。适合不修改对象状态的方法：

```go
func (l Learner) Level() string
```

`Level` 只根据字段计算等级，不需要修改学习者，所以用值接收者。

### 指针接收者

指针接收者可以修改原对象：

```go
func (l *Learner) CompleteLesson()
```

`CompleteLesson` 会增加完成课程数，所以需要指针接收者。

### 小接口

接口应该从使用方需求出发。demo 中的 `LearnerStore` 只包含当前业务需要的两个能力：

```go
type LearnerStore interface {
	Save(Learner) error
	FindByName(string) (Learner, error)
}
```

接口越小，替换实现和测试越容易。

### internal 包

`internal` 是 Go 的特殊目录。放在 `internal` 下面的包只能被父目录内部的代码导入，适合放当前服务私有的业务代码。

## 代码结构

```text
examples/lesson03-methods-interfaces-and-packages/
  go.mod
  README.md
  main.go
  internal/
    learning/
      learner.go
      learner_test.go
      store.go
      store_test.go
```

- `main.go`：应用入口，只负责编排 demo。
- `internal/learning/learner.go`：学习者实体和方法。
- `internal/learning/store.go`：小接口和内存实现。
- `*_test.go`：分别测试实体行为和存储行为。

## 运行方式

进入 demo 目录：

```bash
cd examples/lesson03-methods-interfaces-and-packages
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

`Learner` 使用 `struct` 表达数据：

```go
type Learner struct {
	Name             string
	CompletedLessons int
}
```

`Level` 是值接收者，因为它不修改状态：

```go
func (l Learner) Level() string {
	switch {
	case l.CompletedLessons >= 8:
		return "advanced"
	case l.CompletedLessons >= 3:
		return "intermediate"
	default:
		return "beginner"
	}
}
```

`CompleteLesson` 是指针接收者，因为它会修改字段：

```go
func (l *Learner) CompleteLesson() {
	l.CompletedLessons++
}
```

存储接口只定义调用方需要的能力：

```go
type LearnerStore interface {
	Save(Learner) error
	FindByName(string) (Learner, error)
}
```

内存实现可以替换为数据库实现，而调用方不用改变。后续课程会用同样的思路把内存版 Task API 替换成 PostgreSQL 版。

## 练习

1. 给 `Learner` 增加 `CurrentTrack string` 字段，例如 `go-backend`。
2. 新增 `ResetProgress()` 方法，把完成课程数重置为 0。
3. 给 `LearnerStore` 增加 `List() []Learner` 能力，并为内存实现补测试。

## 常见坑

- 不要把 Go interface 当成 TypeScript interface 的完全等价物。
- 需要修改接收者状态时，用指针接收者。
- 不需要修改状态的小值对象，可以优先用值接收者。
- interface 不一定要放在实现方；很多时候由使用方定义更自然。
- 不要过早抽象出很大的接口，小接口更容易测试和替换。

## 下一课

下一课进入标准库 HTTP 服务。你会用 `net/http` 写第一个 JSON API，并用 `httptest` 测试 handler。
