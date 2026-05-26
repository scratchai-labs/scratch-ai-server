# 项目结构

当前仓库是拆分后的服务器端 monorepo。正式主线已经收口为 `Go API + 教师管理 Web`；仓库里现有的 `Python FastAPI` 代码只作为过渡原型看待。

## 顶层目录

- `apps/server-api`
  - 服务器 API 主工程
  - 当前已经是 Go 实现
  - 负责认证、学生管理、任务管理、`sb3` 解析、进度上报、提示生成和教师看板接口
- `apps/server-web`
  - 教师管理 Web
  - 负责教师登录、学生管理、任务管理和实时看板
- `docs`
  - 开源入口文档与工程说明

## 当前实现补充

- `apps/server-api` 当前基于 `Gin`
- 默认数据库是 `SQLite`，配置 `DATABASE_URL` 后切到 `Postgres`
- `apps/server-web` 既支持 mock API，也支持通过 `VITE_SERVER_WEB_API_MODE=real` 联调真实后端

## 当前产品边界

- 维护中的产品是 `Scratch AI 教练` 服务器端教学版
- 核心是服务端 API
- 教师通过 Web 管理学生和任务
- 学生通过客户端登录并接收提示
- 所有 AI 处理都放在服务端

## 阅读建议

- 第一次访问仓库：先看 [`../README.zh-CN.md`](../README.zh-CN.md)
- 想看服务器端目标方案：看 [`./server-development.zh-CN.md`](./server-development.zh-CN.md)
- 想理解模块职责：看 [`./architecture.zh-CN.md`](./architecture.zh-CN.md)
- 想看维护约定：看 [`./maintenance.zh-CN.md`](./maintenance.zh-CN.md)
