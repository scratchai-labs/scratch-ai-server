# 贡献指南

感谢你关注 `Scratch AI 教练`。

这个仓库当前只维护一个明确目标：`apps/server-api` + `apps/server-web` 的服务器端 monorepo。提交 PR 前，请先对齐下面这些规则。

## 开始前

- 先阅读 [`README.zh-CN.md`](README.zh-CN.md)
- 看仓库结构：[`docs/project-structure.zh-CN.md`](docs/project-structure.zh-CN.md)
- 看服务器端开发说明：[`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)
- 大改动、新功能或路线调整，请先提 issue 讨论

## 环境要求

- Node.js `22`
- Go `1.26`
- npm workspace
- Windows / macOS / Linux 均可

初始化：

```bash
npm ci
npm run test
```

## 推荐贡献方式

- 修文档：README、仓库结构、服务器端开发说明、支持文档
- 修缺陷：优先附复现步骤、接口信息和日志
- 补测试：特别欢迎为 API 契约、前后端联调和 `server-web` smoke test 补回归测试
- 提功能建议：请先说明课堂场景、用户角色和成功标准

## 提交前要求

- 尽量先写失败测试，再补实现
- 改动命令、目录、发布口径时，同步更新相关文档
- 运行与你改动范围匹配的测试；不确定时至少执行 `npm run test`
- 不要提交 `node_modules/`、`dist/`、`.venv/`、`.pytest_cache/`、临时截图或本地调试垃圾文件

## 提交信息

仓库当前推荐使用中文提交说明，并保留简洁的类型前缀：

- `feat:` 新功能
- `fix:` 缺陷修复
- `improve:` 结构、文档、工程整理

推荐格式：

```text
improve: 整理服务器端仓库基础文档与清理脚本
问题或需求描述：仓库里仍残留旧 workspace 引用
修复或实现思路：收口 README、支持文档、清理脚本和忽略规则
```

## Pull Request 期望

- 说明改动动机
- 说明影响范围
- 列出验证命令
- 如果有 API 或 Web 入口变化，附关键说明或截图
- 如果是 roadmap 级别改动，请明确哪些内容是“当前实现”，哪些只是“未来计划”

## 当前不建议的改动

- 未经讨论就引入新的运行时、后端语言或重型基础设施
- 在一个 PR 中同时做产品重构、文档重写和发布流程重做
- 未验证就修改 API 契约或本地启动链路
- 把未实现的桌面或单机方案直接并入当前主线

## 行为与安全

- 社区互动请遵守 [`CODE_OF_CONDUCT.zh-CN.md`](CODE_OF_CONDUCT.zh-CN.md)
- 安全问题请走 [`SECURITY.zh-CN.md`](SECURITY.zh-CN.md) 中的私下披露方式
- 使用问题与讨论入口见 [`SUPPORT.zh-CN.md`](SUPPORT.zh-CN.md)
