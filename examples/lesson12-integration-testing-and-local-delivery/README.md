# 第 12 课 Demo：整合、测试与本地交付

这个 demo 整合 PostgreSQL、用户注册登录、JWT 认证、Task API、Makefile 和 smoke test。

## 快速开始

```bash
make compose-up
make integration-test
make run
```

在另一个终端：

```bash
make smoke
```

停止依赖：

```bash
make compose-down
```

## 常用命令

```bash
make test
make integration-test
make run
make smoke
```
