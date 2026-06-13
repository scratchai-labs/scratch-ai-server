# Scratch AI 教练

`Scratch AI 教练` 服务器端仓面向 Scratch 教学场景，当前主线已经落到“基于 `Gin` 的 `Go` 服务端 + 教师管理 Web”。仓库内已经清理掉旧的 `Python FastAPI` 服务端原型。

跨仓库文档、总体架构和路线图已迁到 [`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs) 统一维护。

## 为什么做这个项目

Scratch 帮很多人第一次真正喜欢上电脑、理解程序和创作。这个项目希望把课堂里的“老师示例作品、学生当前进度、下一步提示”收口成一条长期可维护的开源服务链路。

## 当前支持范围

- 当前仓库只维护 **服务器端教学版**
- 目标技术栈为“基于 `Gin` 的 `Go` 服务端 + Web”
- 包含 `server-api` 和 `server-web`
- 中文是当前主语言

## 目标能力

- 教师注册、登录
- 管理员创建、禁用、启用、重置教师和学生账号
- 教师批量创建学生账号和密码
- 教师上传参考 `sb3`
- 服务端分析参考 `sb3`
- 学生客户端登录
- 学生进度上报
- 服务端调用 DeepSeek 生成下一步提示
- 教师查看实时进度和提示

## 当前开发口径

- 核心是 API
- 管理员通过 Web 管理教师和学生账号，默认入口为 `/admin`
- 教师通过 Web 管理学生和任务
- 学生只通过客户端登录和接收提示
- 所有 AI 处理都放在服务端

机器可读 API 契约以 [`apps/server-api/docs/swagger.json`](apps/server-api/docs/swagger.json) 和 [`apps/server-api/docs/swagger.yaml`](apps/server-api/docs/swagger.yaml) 为准。
客户端对接优先看 [`docs/server-api-contract.zh-CN.md`](docs/server-api-contract.zh-CN.md) 里的调用顺序和示例。
详细开发背景再看 [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)。
部署落地优先看 [`docs/deployment.zh-CN.md`](docs/deployment.zh-CN.md)。

## 本地开发

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
- `sb3` 原文件默认落到本地目录 `SB3_STORAGE_DIR`

当前部署口径：

- 线上测试环境和正式环境分开
- `server-api` 部署在 `Zeabur`
- `server-web` 部署在 `Vercel`
- `staging` 和 `production` 各自使用独立 `DATABASE_URL`
- `staging` 和 `production` 各自使用独立 `SB3_STORAGE_DIR`

当前联调状态：

- 教师 Web 已完成一次真实浏览器点击验证
- 教师端真实 API 联调已通过登录、学生列表、发布单列表、实时看板和退出登录主流程
- 当前后端已放开教师 Web 本地联调所需的 `CORS` 预检请求
- 管理员后台已落地，可通过 `ADMIN_BOOTSTRAP_USERNAME` / `ADMIN_BOOTSTRAP_PASSWORD` 自举首个管理员账号并登录 `/admin`

## 文档导航

- 仓库结构：[`docs/project-structure.zh-CN.md`](docs/project-structure.zh-CN.md)
- 架构说明：[`docs/architecture.zh-CN.md`](docs/architecture.zh-CN.md)
- 部署指南：[`docs/deployment.zh-CN.md`](docs/deployment.zh-CN.md)
- API 契约真值源：[`apps/server-api/docs/swagger.json`](apps/server-api/docs/swagger.json) / [`apps/server-api/docs/swagger.yaml`](apps/server-api/docs/swagger.yaml)
- 接入指南：[`docs/server-api-contract.zh-CN.md`](docs/server-api-contract.zh-CN.md)
- 服务器端开发说明：[`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)
- 跨仓库文档与规划：[`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs)
- 开发工作流：[`scratch-ai-docs/docs/development-workflow.zh-CN.md`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/docs/development-workflow.zh-CN.md)
- 文档归属说明：[`scratch-ai-docs/docs/documentation-guide.zh-CN.md`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/docs/documentation-guide.zh-CN.md)

## 参与贡献

欢迎通过 issue、PR、文档修订和教学场景反馈参与项目。

- 提交代码前请阅读 [`CONTRIBUTING.zh-CN.md`](CONTRIBUTING.zh-CN.md)
- 社区互动请遵守 [`CODE_OF_CONDUCT.zh-CN.md`](CODE_OF_CONDUCT.zh-CN.md)
- 安全问题请不要公开提 issue，见 [`SECURITY.zh-CN.md`](SECURITY.zh-CN.md)
- 使用问题和讨论入口见 [`SUPPORT.zh-CN.md`](SUPPORT.zh-CN.md)

## 许可证

本项目采用 [`AGPL-3.0`](LICENSE) 许可证。
