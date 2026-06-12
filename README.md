# Go Backend Course for Frontend Architects

这是一个面向 7 年前端架构师转型 Go 后端的循序渐进课程。

课程从 Go 零基础开始，但不从编程零基础开始。默认学习者已经熟悉变量、函数、模块化、HTTP、JSON、前端工程化和 TypeScript/Node.js 的常见开发经验。

## 课程约定

- 课程语言：文档、README、代码注释使用中文；代码标识符使用英文。
- Go 版本：Go 1.24+。
- 运行环境：只提供 macOS/Linux/类 Unix 命令。
- 课程结构：每课一个独立文档，每课一个独立可运行 demo。
- Demo 结构：每课 demo 都是独立 Go module，便于单独运行和测试。
- 测试策略：每课都包含可运行测试，逐步从纯函数测试扩展到 HTTP、数据库和集成测试。
- 依赖策略：前 5 课只使用 Go 标准库，第 6 课开始引入少量真实后端项目常用依赖。
- 提交策略：课程总体规划单独提交；之后每完成一课，单独 commit 并 push 到 `main`。

## 目录结构

```text
docs/
  course-outline.md
  lessons/
    01-hello-go.md
    02-types-functions-and-errors.md
    ...
examples/
  lesson01-hello-go/
  lesson02-types-functions-and-errors/
  ...
```

## 学习路径

完整课程大纲见 [docs/course-outline.md](docs/course-outline.md)。

| 课次 | 主题 | 文档 | Demo |
| --- | --- | --- | --- |
| 第 1 课 | Hello Go 与最小工程 | [docs/lessons/01-hello-go.md](docs/lessons/01-hello-go.md) | [examples/lesson01-hello-go](examples/lesson01-hello-go) |
| 第 2 课 | 类型、函数、错误处理 | [docs/lessons/02-types-functions-and-errors.md](docs/lessons/02-types-functions-and-errors.md) | [examples/lesson02-types-functions-and-errors](examples/lesson02-types-functions-and-errors) |
| 第 3 课 | 方法、接口与包设计 | [docs/lessons/03-methods-interfaces-and-packages.md](docs/lessons/03-methods-interfaces-and-packages.md) | [examples/lesson03-methods-interfaces-and-packages](examples/lesson03-methods-interfaces-and-packages) |
| 第 4 课 | 标准库 HTTP 服务 | [docs/lessons/04-standard-library-http.md](docs/lessons/04-standard-library-http.md) | [examples/lesson04-standard-library-http](examples/lesson04-standard-library-http) |
| 第 5 课 | Context、超时与请求生命周期 | [docs/lessons/05-context-timeout-and-lifecycle.md](docs/lessons/05-context-timeout-and-lifecycle.md) | [examples/lesson05-context-timeout-and-lifecycle](examples/lesson05-context-timeout-and-lifecycle) |

每一课都会包含：

- 学习目标
- 前端架构师视角
- 核心概念
- 代码结构
- 运行方式
- 关键代码讲解
- 练习
- 常见坑
- 下一课

## 最终项目

课程最终会构建一个 Task API / 团队任务管理后端，覆盖用户注册登录、JWT 认证、任务 CRUD、任务状态流转、分页查询、PostgreSQL 持久化、中间件、日志、错误处理、测试和 Docker Compose 启动。
