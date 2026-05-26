# Scratch AI Coach 主工程架构说明

本文描述服务器端仓当前维护中的主线架构。正式方向已经收口为“基于 `Gin` 的 `Go` 服务端 + 教师管理 Web”，旧的 `Python FastAPI` 原型已经从仓库清理。

如果只想看需求与接口范围，优先阅读 [`./server-development.zh-CN.md`](./server-development.zh-CN.md)。

## 1. 工作区边界

当前仓库继续保持服务器端 monorepo：

- `apps/server-api`
  - Go 服务端主工程
- `apps/server-web`
  - 教师管理 Web

仓库职责只有一条：

1. 维护面向课堂场景的联机教学服务端

## 2. 组件职责

### `apps/server-api`

Go API，负责：

- 教师注册、登录、退出
- 教师单个创建和批量创建学生
- 学生客户端登录
- 教师上传参考 `sb3`
- `sb3` 解析与分析
- 学生进度上报
- DeepSeek 调用
- 提示日志保存
- 教师实时看板数据聚合
- OpenAPI 生成物是机器可读契约真值源，`docs/server-api-contract.zh-CN.md` 只保留接入说明

### `apps/server-web`

教师管理 Web，负责：

- 教师登录
- 学生管理
- 任务管理
- `sb3` 上传与分析查看
- 学生实时进度查看

### 学生客户端

学生客户端不在本仓库里实现，但它是服务端主链路的一部分，负责：

- 学生登录
- 读取本地 Scratch 项目状态
- 上报结构化进度
- 接收服务端返回的提示

## 3. 主数据流

```text
Teacher Web -> Go API -> SQLite / Postgres
Student Client -> Go API -> SQLite / Postgres
Teacher Web -> Go API -> sb3 Storage
Go API -> DeepSeek API
Teacher Web <- Go API <- 聚合后的进度与提示
```

主链路：

1. 教师在 Web 端登录。
2. 教师创建学生账号，并上传参考 `sb3`。
3. Go API 保存参考 `sb3`，并在后台异步分析。
4. 教师等待分析完成后，把任务分配给学生。
5. 学生在客户端登录并读取自己的任务。
6. 学生客户端从本地 Scratch 项目提取结构化快照并上报。
7. Go API 结合参考 `sb3` 分析结果与学生快照调用 DeepSeek。
8. 学生客户端接收下一步提示。
9. 教师 Web 轮询查看实时进度和提示记录。

## 4.1 当前实现状态

- API 路由与中间件基于 `Gin`
- 数据存储默认使用 `SQLite` 文件库，生产可切到 `Postgres`
- `apps/server-api/docs/swagger.json` 和 `swagger.yaml` 是当前接口契约的机器可读真值源
- `sb3` 分析采用异步处理，并在服务启动时恢复 `pending / processing` 任务
- 提示链路优先调用真实 `DeepSeek`，失败时回退到本地规则提示
- 教师 Web 与真实 API 的本地联调已经通过浏览器点击验证，API 侧已补 `CORS` 预检支持

## 4. 当前架构约束

- 第一阶段只做“教师 / 学生 / 任务”三类核心对象
- 所有 AI 调用只放在服务端
- 学生只能通过客户端登录，不开放学生 Web 入口
- 教师看板第一阶段统一使用轮询
- 学生端优先上传结构化快照，不把完整本地 `sb3` 高频上传到服务端

## 5. 当前风险点

- `OpenAPI` 生成物需要和 `internal/http` 路由、注释保持同步，否则接入文档会失真
- `sb3` 上传后会引入文件存储与分析耗时，接口要明确同步和异步边界
- DeepSeek 是外部依赖，必须准备失败兜底，不能把学生端体验完全绑死在第三方响应上
