# Scratch AI 教练 / Scratch AI Coach

面向 Scratch 教学场景的开源服务器端工作区。当前正式主线已经收口为“基于 `Gin` 的 `Go` 服务端 + Teacher Web”，仓库内不再保留旧的 `Python FastAPI` 服务端原型。机器可读 API 契约以 `apps/server-api/docs/swagger.json` 为准，`docs/server-api-contract.zh-CN.md` 只保留接入指南和补充说明。
An open source server workspace for Scratch teaching. The current production track is now a `Gin`-based `Go` backend plus the Teacher Web app, and the older `Python FastAPI` server prototype has been removed from this repository. The machine-readable API contract comes from `apps/server-api/docs/swagger.json`; `docs/server-api-contract.zh-CN.md` stays as a human-readable integration guide.

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
- 部署指南：[`docs/deployment.zh-CN.md`](docs/deployment.zh-CN.md)
- API 契约真值源：[`apps/server-api/docs/swagger.json`](apps/server-api/docs/swagger.json) / [`apps/server-api/docs/swagger.yaml`](apps/server-api/docs/swagger.yaml)
- 接入指南：[`docs/server-api-contract.zh-CN.md`](docs/server-api-contract.zh-CN.md)
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

后端优先直接用 `Go` 命令：

```bash
git clone git@github.com:scratchai-labs/scratch-ai-server.git
cd scratch-ai-server
npm ci
cd apps/server-api
go test ./...
go run ./cmd/server-api
```

教师 Web 单独开发时：

```bash
cd apps/server-web
npm run test
npm run dev
```

如果需要从仓库根目录统一调度，再用这些 `npm run` 快捷命令：

- `npm run server:api:dev`
- `npm run server:api:test`
- `npm run server:web:dev`
- `npm run server:web:test`
- `npm run server:dev`
- `npm run server:build`

当前数据库口径：

- 默认本地开发使用 `SQLite`
- 配置 `DATABASE_URL` 后切到 `Postgres`
- `sb3` 原文件默认保存在 `SB3_STORAGE_DIR`

当前部署口径：

- `server-api` 部署在 `Zeabur`
- `server-web` 部署在 `Vercel`
- `staging` 和 `production` 各自使用独立 `DATABASE_URL`
- `staging` 和 `production` 各自使用独立 `SB3_STORAGE_DIR`

当前联调状态：

- 教师 Web 已完成一次真实浏览器点击验证
- 真实 API 联调已通过登录、学生列表、发布单列表、实时看板和退出登录主流程

## License

本项目采用 [`AGPL-3.0`](LICENSE) 许可证。
This project is licensed under [`AGPL-3.0`](LICENSE).
