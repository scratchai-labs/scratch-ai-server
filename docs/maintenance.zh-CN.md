# 主工程文档维护约定

这份文档只约束当前主线文档：

- `README.md`
- `README.zh-CN.md`
- `README.en.md`
- `CONTRIBUTING*.md`
- `CODE_OF_CONDUCT*.md`
- `SECURITY*.md`
- `SUPPORT*.md`
- `docs/*.md`
- `apps/server-api/*.md`
- `apps/server-web/*.md`

## 1. 文档层级

默认按下面顺序维护：

1. 根 `README.md`
2. `README.zh-CN.md` / `README.en.md`
3. `CONTRIBUTING*` / `CODE_OF_CONDUCT*` / `SECURITY*` / `SUPPORT*`
4. `docs/README.zh-CN.md`
5. `docs/project-structure*` / `docs/roadmap*`
6. `docs/architecture.zh-CN.md`
7. `docs/maintenance.zh-CN.md`
8. `apps/server-api` 下的服务端开发和接口文档
9. `apps/server-web` 下的教师 Web 说明和测试文档

## 2. 每份文档负责什么

### `README.md`

负责：

- 作为 GitHub 首页的双语入口
- 项目定位
- 当前支持范围
- 开源核心文档导航

这些变化后必须更新：

- 项目定位变化
- 语言入口变化
- 许可证或贡献入口变化

### `README.zh-CN.md` / `README.en.md`

负责：

- 中文 / 英文项目总览
- 开发与联调入口
- 对外文档导航

这些变化后必须更新：

- 产品定位变化
- 入口命令变化
- 对外贡献入口变化

### `CONTRIBUTING*` / `CODE_OF_CONDUCT*` / `SECURITY*` / `SUPPORT*`

负责：

- 开源协作规则
- 社区行为边界
- 安全披露路径
- 提问与支持入口

这些变化后必须更新：

- Issue / PR 流程变化
- 安全联系路径变化
- 支持入口变化

### `docs/README.zh-CN.md`

负责：

- 工程文档导航
- 目录收口
- 清理入口

这些变化后必须更新：

- 文档路径变化
- 清理脚本覆盖范围变化
- API 文档或服务器端说明入口变化

### `docs/architecture.zh-CN.md`

负责：

- 组件职责
- 主数据流
- 当前风险点

这些变化后必须更新：

- `server-api` / `server-web` 结构变化
- API 契约或数据流变化
- AI 调用链路变化

### `apps/server-api/*.md` / `apps/server-web/*.md`

负责：

- 服务器 API、教师 Web 的开发、联调和部署说明

这些变化后必须更新：

- 鉴权或接口契约变化
- 本地启动命令变化
- 前后端目录结构变化

## 3. 维护检查清单

改完代码后，至少检查下面几项：

- 根 README 与中英文 README 的入口都可用
- 文档里的路径与真实目录一致
- 文档里的接口契约、脚本入口和本地启动命令与当前配置一致
- 文档是否还提到旧 workspace 路径、已删除目录或其他过时命名
- `npm run clean:dry-run` 的描述是否和脚本输出一致
