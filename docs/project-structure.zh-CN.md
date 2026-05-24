# 项目结构

当前仓库是拆分后的服务器端 monorepo：教师后台继续走 npm workspace，服务器 API 作为并行的 Python 工程维护。

## 顶层目录

- `apps/desktop-companion`
  - Electron 桌面端
  - 负责 Scratch 连接、桥接、AI 调用和主界面
- `apps/server-api`
  - Python FastAPI 服务端
  - 负责老师/学生认证、发布单、进度、AI 提示与教师看板接口
- `apps/server-web`
  - Vue 教师后台
  - 负责老师登录、学生管理、发布单管理和实时看板
- `docs`
  - 开源入口文档与工程说明

## 当前产品边界

- 维护中的产品是 `Scratch AI 教练` 服务器端教学版
- 服务器端当前采用 `Python FastAPI + Vue`
- 本仓库不包含独立本地客户端和联机桌面客户端代码

## 阅读建议

- 第一次访问仓库：先看 [`../README.zh-CN.md`](../README.zh-CN.md)
- 想参与协作：看 [`../CONTRIBUTING.zh-CN.md`](../CONTRIBUTING.zh-CN.md)
- 想理解模块职责：看 [`./architecture.zh-CN.md`](./architecture.zh-CN.md)
- 想看维护约定：看 [`./maintenance.zh-CN.md`](./maintenance.zh-CN.md)
