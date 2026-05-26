# Scratch AI 教练 / Scratch AI Coach

面向 Scratch 教学场景的开源服务器端工作区。当前正式主线已经收口为 `Go API + Teacher Web`，仓库中的 `Python FastAPI` 代码只作为过渡原型保留。
An open source server workspace for Scratch teaching. The current production track is `Go API + Teacher Web`, while the existing `Python FastAPI` code is kept only as a transition prototype.

## Language / 语言

- 中文：[`README.zh-CN.md`](README.zh-CN.md)
- English: [`README.en.md`](README.en.md)

## Current Scope

- 当前主线只维护 **服务器端教学版**
- 核心是服务端 API
- 教师通过 Web 管理学生和任务
- 学生通过客户端登录和接收提示
- 所有 AI 处理都放在服务端

## Quick Links

- 中文总览：[`README.zh-CN.md`](README.zh-CN.md)
- English overview: [`README.en.md`](README.en.md)
- 仓库结构：[`docs/project-structure.zh-CN.md`](docs/project-structure.zh-CN.md)
- 架构说明：[`docs/architecture.zh-CN.md`](docs/architecture.zh-CN.md)
- 开发说明：[`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)
- 贡献指南：[`CONTRIBUTING.zh-CN.md`](CONTRIBUTING.zh-CN.md) / [`CONTRIBUTING.en.md`](CONTRIBUTING.en.md)
- 行为准则：[`CODE_OF_CONDUCT.zh-CN.md`](CODE_OF_CONDUCT.zh-CN.md) / [`CODE_OF_CONDUCT.en.md`](CODE_OF_CONDUCT.en.md)
- 安全说明：[`SECURITY.zh-CN.md`](SECURITY.zh-CN.md) / [`SECURITY.en.md`](SECURITY.en.md)
- 支持与提问：[`SUPPORT.zh-CN.md`](SUPPORT.zh-CN.md) / [`SUPPORT.en.md`](SUPPORT.en.md)
- 跨仓库文档与规划 / Cross-repo docs and planning: [`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs)

## Current Direction

- 教师注册、登录
- 教师批量创建学生账号和密码
- 教师上传并分析参考 `sb3`
- 学生客户端登录与进度上报
- 服务端调用 DeepSeek 生成下一步提示
- 教师查看实时进度与提示

## Local Development

当前根目录命令已经对齐到 Go 服务端与教师 Web：

```bash
git clone git@github.com:scratchai-labs/scratch-ai-server.git
cd scratch-ai-server
npm ci
npm run server:api:test
npm run server:web:test
npm run server:dev
```

当前数据库口径：

- 默认本地开发使用 `SQLite`
- 配置 `DATABASE_URL` 后切到 `Postgres`
- `sb3` 原文件默认保存在 `SB3_STORAGE_DIR`

当前联调状态：

- 教师 Web 已完成一次真实浏览器点击验证
- 真实 API 联调已通过登录、学生列表、发布单列表、实时看板和退出登录主流程

## License

本项目采用 [`AGPL-3.0`](LICENSE) 许可证。
This project is licensed under [`AGPL-3.0`](LICENSE).
