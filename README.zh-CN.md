# Scratch AI 教练

`Scratch AI 教练` 服务器端仓负责教学场景下的联机链路：基于 `Python FastAPI + Vue` 维护老师后台、学生账号、发布单、进度上报和 AI 提示接口。
跨仓库文档、总体架构和路线图已迁到 [`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs) 统一维护。

## 为什么做这个项目

Scratch 帮很多人第一次真正喜欢上电脑、理解程序和创作。Scratch 本身也是开源项目，所以这个工具也希望按长期可维护的开源仓库方式运营，让更多老师、学生和开发者可以直接使用、反馈、贡献和继续演进。

## 当前支持范围

- 当前仓库只维护 **服务器端教学版**
- 技术栈为 `Python FastAPI + Vue`
- 包含 `server-api` 和 `server-web`
- 当前默认面向中文用户，但开源核心文档已提供英文版本

## 当前能力

- 提供老师注册、登录
- 提供学生账号创建与登录
- 提供 `sb3` 发布单管理
- 接收学生进度上报
- 生成服务器端 AI 提示
- 提供教师实时看板

## 下载与发布

当前仓库不产出桌面安装包。发布重点是：

- `apps/server-api` 的服务部署
- `apps/server-web` 的前端构建与部署

部署说明见 [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)。

## 本地开发

```bash
git clone git@github.com:scratchai-labs/scratch-ai-server.git
cd scratch-ai-server
npm ci
npm run test
```

常用命令：

```bash
npm run build
npm run test
npm run server:web:test
npm run server:api:test
npm run server:dev
```

服务端联调：

```bash
npm run server:dev
```

## 文档导航

- 仓库结构：[`docs/project-structure.zh-CN.md`](docs/project-structure.zh-CN.md)
- 服务器端开发说明：[`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)
- 跨仓库文档与规划：[`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs)
- 开发工作流：[`scratch-ai-docs/docs/development-workflow.zh-CN.md`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/docs/development-workflow.zh-CN.md)
- 文档归属说明：[`scratch-ai-docs/docs/documentation-guide.zh-CN.md`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/docs/documentation-guide.zh-CN.md)
- 工程文档索引：[`docs/README.zh-CN.md`](docs/README.zh-CN.md)
- 服务器 API：`apps/server-api`
- 教师后台：`apps/server-web`

## 参与贡献

欢迎通过 issue、PR、文档修订和教学场景反馈参与项目。

- 提交代码前请阅读 [`CONTRIBUTING.zh-CN.md`](CONTRIBUTING.zh-CN.md)
- 社区互动请遵守 [`CODE_OF_CONDUCT.zh-CN.md`](CODE_OF_CONDUCT.zh-CN.md)
- 安全问题请不要公开提 issue，见 [`SECURITY.zh-CN.md`](SECURITY.zh-CN.md)
- 使用问题和讨论入口见 [`SUPPORT.zh-CN.md`](SUPPORT.zh-CN.md)

## 未来方向

跨仓库层面的总体规划已经转到 [`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs) 统一维护。
当前仓主要聚焦服务器 API、教师后台和联机教学链路。

## 许可证

本项目采用 [`AGPL-3.0`](LICENSE) 许可证。
