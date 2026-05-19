# Go 后端学习路线

> 目标：从资深前端工程师转型为企业级 Go 后端，共 7 个阶段，每阶段都有可落地的项目产出。

| 阶段 | 主题 | 项目 | 预估时长 |
|------|------|------|----------|
| [1](./phase-1-go-basics.md) | Go 语言基础 | 命令行 Todo 工具 | 1-2 周 |
| [2](./phase-2-concurrency.md) | 并发与标准库 | 文件批量处理器 | 1-2 周 |
| [3](./phase-3-web-http.md) | Web 基础 (net/http) | RESTful 任务 API | 2-3 周 |
| [4](./phase-4-database.md) | 数据库与持久化 | 数据库版任务 API | 2-3 周 |
| [5](./phase-5-project-engineering.md) | 项目工程化 | 用户认证系统 | 3-4 周 |
| [6](./phase-6-advanced.md) | 进阶服务 | 实时通知服务 | 3-4 周 |
| [7](./phase-7-production.md) | 生产就绪 | 微服务化改造 | 4-6 周 |

## 使用说明

- 每个阶段的文档包含：核心知识点 → 项目实战 → 延伸阅读 → 检查清单
- 代码约定：所有示例和项目代码直接在本仓库根目录对应子目录中编写
- 前置知识：TypeScript/JavaScript 精通，HTTP 协议熟悉，数据库基础概念了解

## 三个「思维切换」

在开始之前，了解 Go 和你熟悉的 TS/Node 生态的三个根本差异：

1. **没有类继承** — Go 用组合（embedding）和接口（interface）替代 OOP
2. **错误即返回值** — 没有 try/catch/throw，error 是普通返回值，显式处理
3. **并发是语言特性** — goroutine 和 channel 内建于语言，不是库提供的
