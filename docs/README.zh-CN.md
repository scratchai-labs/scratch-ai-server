# 工程文档索引

`docs/` 现在主要承载两类内容：

- 面向开源协作的入口文档
- 面向维护者的工程说明

## 开源入口

- 项目总览：[`../README.zh-CN.md`](../README.zh-CN.md)
- English overview: [`../README.en.md`](../README.en.md)
- 贡献指南：[`../CONTRIBUTING.zh-CN.md`](../CONTRIBUTING.zh-CN.md)
- 跨仓库文档与规划：[`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs)
- 跨仓开发工作流：[`scratch-ai-docs/docs/development-workflow.zh-CN.md`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/docs/development-workflow.zh-CN.md)
- 文档归属说明：[`scratch-ai-docs/docs/documentation-guide.zh-CN.md`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/docs/documentation-guide.zh-CN.md)
- 仓库结构：[`./project-structure.zh-CN.md`](./project-structure.zh-CN.md)

## 工程文档

- 架构说明：[`./architecture.zh-CN.md`](./architecture.zh-CN.md)
- 部署指南：[`./deployment.zh-CN.md`](./deployment.zh-CN.md)
- 服务器端开发说明：[`./server-development.zh-CN.md`](./server-development.zh-CN.md)
- 数据库结构变更默认走 `server-api` 启动时的内置自动迁移，当前真值口径见部署指南与 `apps/server-api/README.md`
- 教师后台 mock 浏览器验证与自动化脚本：[`./server-web-mock-smoke-test.zh-CN.md`](./server-web-mock-smoke-test.zh-CN.md)
- 教师后台 real-mode 浏览器验证与自动化脚本：[`./server-web-real-smoke-test.zh-CN.md`](./server-web-real-smoke-test.zh-CN.md)
- 文档维护约定：[`./maintenance.zh-CN.md`](./maintenance.zh-CN.md)
- 服务器 API：`../apps/server-api`
- 教师后台：`../apps/server-web`

## 目录约定

- `./assets/screenshots/`
  - 文档截图目录

## 清理入口

```bash
npm run clean:dry-run
npm run clean
```
