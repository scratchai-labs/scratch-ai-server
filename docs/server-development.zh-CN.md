# 服务器端开发说明

本文收口服务器端当前实现与后续维护口径。正式 API 主线已经收口为基于 `Gin` 的 `Go` 服务端，教师管理端继续保留 `Web` 形态；旧的 `Python FastAPI` 原型已经从仓库清理。

如果只是为了对接客户端，可以先看 [`./server-api-contract.zh-CN.md`](./server-api-contract.zh-CN.md) 里的调用顺序和示例；字段、路径和响应的真值以 `apps/server-api/docs/swagger.json` 和 `apps/server-api/docs/swagger.yaml` 为准。

当前接口契约的机器可读真值源是 `apps/server-api` 生成的 OpenAPI 产物，位于：

- `apps/server-api/docs/swagger.json`
- `apps/server-api/docs/swagger.yaml`
- `apps/server-api/docs/docs.go`

本地启动服务后，可直接打开：

- `http://127.0.0.1:8000/swagger/index.html`
- `http://127.0.0.1:8000/swagger/doc.json`

更新接口后，需要同步执行：

```bash
npm run server:api:docs
```

## 0. 当前实现状态

当前代码已经完成第一阶段主链路，真实实现口径如下：

- 服务端语言：`Go`
- HTTP 框架：`Gin`
- 默认数据库：`SQLite`
- 可选数据库：`Postgres`（配置 `DATABASE_URL` 后启用）
- `sb3` 存储：本地目录 `SB3_STORAGE_DIR`
- AI 上游：`DeepSeek API`
- 教师端：`Vue 3 + Vite`

已落地能力：

- 管理员自举登录、后台总览、教师账号列表、新建、角色切换、重置密码、启停，学生账号列表、新建、重置密码、启停，以及管理员操作日志查询
- 教师注册、登录、退出、`me`
- 教师单个创建学生、批量创建学生、重置学生密码
- 学生客户端登录、退出、`me`
- 教师上传参考 `sb3`
- `sb3` 异步分析、失败记录、启动恢复、有限重试
- 教学任务分配、发布、归档
- 学生任务列表、任务详情、进度上报、提示请求
- DeepSeek 提示生成与 fallback
- 教师实时看板、学生历史查询
- 教师 Web 真实 API 联调与浏览器点击验证

## 0.1 后台选型建议（2026-06）

当前仓库不是从零开始选后台，而是在已经跑通的 `Gin + 自定义 Go API + Vue 教师 Web` 上继续扩能力。现状已经包含：

- 教师 / 管理员 / 学生三套角色边界
- 教师业务接口、学生业务接口、管理员总览 / 教师 / 学生管理接口
- 管理员登录后进入 `/admin`，教师登录后进入原有教学后台

因此，当前阶段的默认建议是：继续沿现有仓库自研扩展，不建议现在切整套开源管理系统。

在“1 名熟悉当前仓库的 Go/Vue 工程师”前提下，可粗略按下面评估：

| 方案 | 如何落到当前仓库 | 一次性工程量 | 优点 | 主要成本 / 风险 | 当前建议 |
| --- | --- | --- | --- | --- | --- |
| 继续自研扩展 | 继续在 `apps/server-api/internal/admin` 与 `apps/server-web` 上补人员管理、角色、审计 | 人员管理基础版约 `3-7` 人日；若补 `RBAC / 审计 / 系统配置`，约 `2-5` 周 | 与现有教师/学生/任务/`sb3`/提示链路完全贴合；改动可控；不需要迁移登录和路由 | 通用后台能力要自己做，代码生成、菜单、审计等需要逐步补齐 | `强烈推荐` |
| 接入 `go-admin-team/go-admin` | 更适合作为“独立管理中心”挂到现有库表或新管理 API 上 | 最小接入约 `1-2` 周；若要统一登录、权限、菜单和现有课堂业务，约 `3-6` 周 | 自带用户、角色、菜单、日志、代码生成等通用基建，适合做完整中后台 | 需要对接现有认证、权限、菜单、前端壳层、数据模型；容易出现两套后台并存 | `仅在需求升级时考虑` |
| 接入 `gin-vue-admin` | 更像替换成另一套完整全栈脚手架 | 通常不低于 `3-6` 周 | 功能更全，生态和脚手架能力强 | 迁移最重；现有业务要适配其工程结构；官方 README 标注 `BSL 1.1`，商用前必须单独核对授权边界 | `当前不建议` |
| 接入 `GoAdminGroup/go-admin` | 更适合作为独立 CRUD / 运维数据面板 | 约 `1-2` 周可出基础数据页 | 轻，适合快速挂简单管理页 | 更偏“数据面板框架”，不适合直接承接你们已有教师 SPA 和课堂业务工作流 | `可作轻量备选` |

继续自研并不意味着“不能做人员管理”。按当前结构，完全可以继续补：

- 管理员管理教师、学生、管理员账号
- 新建教师、禁用教师、重置密码、角色切换
- 后续增加组织、班级、菜单权限

只有在目标明显升级为“完整通用中后台”时，接入开源才更划算。典型信号包括：

- 除教师外，还要长期维护运营、教务、审核、客服等多角色后台
- 需要成熟的 `RBAC`、菜单、部门、岗位、审计日志、系统参数
- 需要大量通用 CRUD 页面，希望依赖代码生成器提速
- 可以接受把管理后台独立成第二个产品，而不是继续塞进当前教师 Web

如果后续确定要切开源，当前仓库更建议优先评估 `go-admin-team/go-admin`，并把它定位为“独立管理中心”，不要直接替换现有教师业务后台。

后端日常开发优先直接用 `Go` 命令：

```bash
cd apps/server-api
go test ./...
go run ./cmd/server-api
```

从仓库根目录统一调度时，再用这些 `npm run` 快捷命令：

```bash
npm run server:api:test
npm run server:web:test
npm run server:dev
```

## 1. 文档定位

这份文档解决 4 个问题：

- 服务器端到底做什么，不做什么
- 教师端、学生端、AI 处理分别放在哪里
- 第一阶段 API 应该先做哪些接口
- Go 服务端应该按什么模块拆

当前结论：

- 核心是基于 `Gin` 的 `Go API`
- 教师需要一个 `Web` 管理端
- 学生端只负责登录、读取本地 Scratch 状态、接收提示、上报进度
- 所有 AI 调用都放在服务端，不放在学生客户端
- OpenAPI 规格从 `Go` 代码生成；手写 `server-api-contract.zh-CN.md` 只保留接入指南和补充说明，避免与实现漂移

## 2. 产品目标

服务器端第一阶段只做课堂教学主链路：

1. 教师注册、登录
2. 教师批量创建学生账号和初始密码
3. 教师上传参考 `sb3` 文件
4. 服务端解析并分析参考 `sb3`
5. 学生客户端登录后读取自己的任务
6. 学生客户端读取本地 Scratch 项目状态并持续上报
7. 服务端结合“教师参考 `sb3` + 学生当前进度”调用 `DeepSeek API`
8. 学生客户端接收下一步提示
9. 教师在 Web 端查看学生最新进度与最新提示

## 3. 明确边界

第一阶段明确不做：

- 学生自助注册
- 班级 / 课程 / 学期多层模型
- 浏览器端直接调用 DeepSeek
- 学生客户端直接持有 DeepSeek API Key
- WebSocket / SSE 强推送
- 在线编辑 Scratch 项目
- 学生频繁上传完整 `sb3` 原文件

第一阶段默认这样收口：

- 教师上传一次“参考 `sb3`”
- 学生客户端本地解析当前 Scratch 项目，上传结构化进度快照
- 服务端只做比对、编排、提示生成和日志沉淀

## 4. 角色与权限

### 4.1 教师

教师可以：

- 注册、登录
- 创建和管理学生账号
- 批量导入学生账号与初始密码
- 上传和管理参考 `sb3`
- 查看 `sb3` 分析结果
- 查看学生实时进度
- 查看服务端生成的提示记录

教师不应该：

- 直接拿学生 token 调用学生接口
- 在浏览器里直接暴露 DeepSeek Key

### 4.2 学生

学生只能通过客户端登录。

学生可以：

- 在客户端输入账号密码登录
- 查看自己被分配的任务
- 上报当前进度
- 主动请求下一步提示
- 接收服务端返回的提示

学生不可以：

- 注册老师账号
- 创建其他学生
- 上传参考 `sb3`
- 访问教师管理接口

### 4.3 服务端

服务端负责：

- 鉴权
- 权限校验
- `sb3` 文件存储
- `sb3` 解析与分析
- 学生进度存储
- DeepSeek 调用
- 提示生成日志
- 教师看板聚合

## 5. 主链路

```text
Teacher Web
    |
    v
Go API  <----> SQLite / Postgres
  |  \
  |   \----> Local File Storage (sb3)
  |
  +----> DeepSeek API
  ^
  |
Student Client
```

主流程：

1. 教师在 `Teacher Web` 注册或登录。
2. 教师批量创建学生账号。
3. 教师上传参考 `sb3`，并填写任务标题、教学目标、说明。
4. `Go API` 保存原始 `sb3`，创建任务记录，并异步分析 `project.json`。
5. 教师等待分析状态变成 `ready`，再把任务分配给一个或多个学生。
6. 学生在客户端登录，只能看到分配给自己的任务。
7. 学生客户端读取本地 Scratch 项目，生成结构化快照并上报。
8. 学生请求提示时，服务端把“参考 `sb3` 分析结果 + 学生最新快照 + 最近提示历史”发给 `DeepSeek API`。
9. 服务端保存提示结果，并把下一步提示返回给学生客户端。
10. 教师 Web 轮询看板接口，查看每个学生的最新进度和最新提示。

## 6. 推荐工程结构

当前仓库建议继续保持 monorepo：

- `apps/server-api`
  - Go 服务端主工程
- `apps/server-web`
  - 教师管理 Web
- `docs`
  - 架构、开发、维护文档

`apps/server-api` 推荐目录：

- `cmd/server-api`
  - 服务启动入口
- `internal/auth`
  - 教师 / 学生登录、会话、权限
- `internal/student`
  - 学生账号、批量创建、查询
- `internal/assignment`
  - 教学任务、分配关系、状态
- `internal/sb3`
  - `sb3` 上传、解包、解析、分析
- `internal/progress`
  - 学生进度快照上报与查询
- `internal/hint`
  - DeepSeek 调用、提示生成、提示日志
- `internal/dashboard`
  - 教师实时看板聚合
- `internal/store`
  - 数据库与存储适配
- `internal/http`
  - 路由、handler、中间件、请求响应结构

当前实现采用：

- HTTP 框架：`Gin`
- JSON：Go 标准库
- ZIP / `project.json` 解析：Go 标准库
- SQLite 驱动：`modernc.org/sqlite`
- Postgres 驱动：`pgx`

当前没有额外依赖 `sb3` 专用三方解析库，仍然按 ZIP 包直接读取 `project.json`。

## 7. 核心数据对象

第一阶段建议收口为 8 组核心对象：

### 7.1 teachers

- `id`
- `username`
- `password_hash`
- `display_name`
- `created_at`

### 7.2 teacher_sessions

- `id`
- `teacher_id`
- `token`
- `expires_at`
- `created_at`

### 7.3 students

- `id`
- `teacher_id`
- `username`
- `password_hash`
- `display_name`
- `status`
- `created_at`

### 7.4 student_sessions

- `id`
- `student_id`
- `token`
- `client_type`
- `expires_at`
- `created_at`

`client_type` 第一阶段固定为 `desktop`，用于明确学生只能从客户端登录。

### 7.5 assignments

- `id`
- `teacher_id`
- `title`
- `goal`
- `description`
- `status`
- `sb3_file_path`
- `analysis_status`
- `analysis_error_message`
- `created_at`
- `updated_at`

这里建议把原先偏发布语义的 `release` 收口为更直接的 `assignment`。

### 7.6 assignment_students

- `assignment_id`
- `student_id`
- `assigned_at`

### 7.7 assignment_analysis

- `assignment_id`
- `role_names_json`
- `script_counts_json`
- `block_counts_json`
- `category_counts_json`
- `broadcast_messages_json`
- `variable_names_json`
- `list_names_json`
- `extensions_json`
- `teaching_points_json`
- `created_at`
- `updated_at`

### 7.8 progress_reports / hint_records

`progress_reports`：

- `id`
- `assignment_id`
- `student_id`
- `current_target`
- `step_summary`
- `snapshot_json`
- `reported_at`

`hint_records`：

- `id`
- `assignment_id`
- `student_id`
- `progress_report_id`
- `prompt_input_json`
- `hint_text`
- `provider_name`
- `created_at`

## 8. API 范围

第一阶段正式接口统一使用 `/api/...`。

### 8.1 教师认证

- `POST /api/teacher/register`
- `POST /api/teacher/login`
- `POST /api/teacher/logout`
- `GET /api/teacher/me`

### 8.2 学生管理

管理员侧同时维护教师/管理员账号，当前后台接口包含：

- `GET /api/admin/teachers`
- `POST /api/admin/teachers`
- `POST /api/admin/teachers/{id}/reset-password`
- `POST /api/admin/teachers/{id}/disable`
- `POST /api/admin/teachers/{id}/enable`
- `POST /api/admin/teachers/{id}/role`
- `GET /api/admin/audit-logs`

其中 `POST /api/admin/teachers/{id}/role` 用于在 `teacher` / `admin` 之间切换角色，自举管理员账号不能把自己降级为教师。

- `GET /api/teacher/students`
- `POST /api/teacher/students`
- `POST /api/teacher/students/batch`
- `POST /api/teacher/students/{id}/reset-password`

`POST /api/teacher/students/batch` 输入建议为 JSON 对象，顶层使用 `students` 数组：

```json
{
  "students": [
    {
      "username": "student01",
      "displayName": "小明",
      "initialPassword": "abc12345"
    },
    {
      "username": "student02",
      "displayName": "小红",
      "initialPassword": "xyz98765"
    }
  ]
}
```

教师 Web 当前已支持“下载 Excel 可打开的模板 + 粘贴表格数据”的批量导入交互；底层仍复用这个 JSON 批量接口，不额外新增文件上传型 API。

返回建议包含：

- 成功创建列表
- 冲突列表
- 失败原因

这样教师一次提交后，不需要手工回查哪几个账号创建成功。

管理员侧还提供全局学生管理入口，用于为指定教师代建账号、统一查询、停用和重置密码：

- `GET /api/admin/students`
- `POST /api/admin/students`
- `POST /api/admin/students/{id}/reset-password`
- `POST /api/admin/students/{id}/disable`
- `POST /api/admin/students/{id}/enable`

`POST /api/admin/students` 请求体应包含：

- `teacherId`
- `username`
- `displayName`
- `initialPassword`

`GET /api/admin/audit-logs` 当前按时间倒序返回管理员对教师/学生账号的敏感操作，MVP 覆盖：

- 教师创建
- 教师密码重置
- 教师启停
- 教师角色切换
- 学生创建
- 学生密码重置
- 学生启停

### 8.3 教学任务与参考 sb3

- `GET /api/teacher/assignments`
- `POST /api/teacher/assignments`
- `GET /api/teacher/assignments/{id}`
- `GET /api/teacher/assignments/{id}/analysis`
- `POST /api/teacher/assignments/{id}/assign-students`
- `POST /api/teacher/assignments/{id}/publish`
- `POST /api/teacher/assignments/{id}/archive`

`POST /api/teacher/assignments` 采用 `multipart/form-data`：

- `title`
- `goal`
- `description`
- `sb3`

上传成功后，接口直接返回已创建的任务基础信息，并把 `analysisStatus` 置为 `pending`。教师端后续通过 `GET /api/teacher/assignments/{id}/analysis` 轮询分析状态。

### 8.4 学生客户端

- `POST /api/student/login`
- `POST /api/student/logout`
- `GET /api/student/me`
- `GET /api/student/assignments`
- `GET /api/student/assignments/{id}`
- `POST /api/student/assignments/{id}/progress`
- `POST /api/student/assignments/{id}/hints`

约束：

- 学生接口只接受学生 token
- 学生只能访问分配给自己的任务
- `student/login` 要求客户端带上固定的 `clientType=desktop`

### 8.5 教师看板

- `GET /api/teacher/dashboard/assignments/{id}/live`
- `GET /api/teacher/dashboard/students/{id}/history`

当前教师 Web 的展示口径：

- 总览页和学生管理页先请求 `GET /api/teacher/students`
- 再按学生补拉 `GET /api/teacher/dashboard/students/{id}/history`
- 前端使用最近一条历史记录渲染 `status / currentTarget / stepSummary / latestAiHint / updatedAt`

第一阶段教师看板统一先用轮询：

- 列表页轮询建议 `10` 秒
- 详情看板轮询建议 `3-5` 秒

## 9. sb3 上传与分析

教师上传 `sb3` 后，服务端需要完成 5 件事：

1. 校验文件扩展名、MIME 和大小限制
2. 保存原始文件
3. 创建任务记录，并把 `analysisStatus` 记为 `pending`
4. 异步解压并读取 `project.json`
5. 生成可用于 AI 提示的结构化分析结果

当前实现已经采用异步分析，不在上传接口里同步等待结果。当前收口方式：

1. 上传接口保存文件和任务记录后立即返回
2. 服务端后台 worker 领取 `pending` 任务
3. worker 把状态改成 `processing`
4. 解析成功后改成 `ready`
5. 解析失败后改成 `failed`，并记录错误信息
6. 服务重启后会恢复 `pending / processing` 任务

`analysisStatus` 建议只有 4 个值：

- `pending`
- `processing`
- `ready`
- `failed`

约束也要写死：

- 任务分析结果未 `ready` 前，教师不能发布任务
- 学生端如果拿到未分析完成的任务详情，可以看到状态，但不能请求提示
- 学生在 `analysisStatus != ready` 时请求提示，服务端返回 `409 Conflict`

第一阶段至少解析这些内容：

- 舞台与角色列表
- 每个角色的脚本数量
- 事件积木分布
- 运动、外观、声音、控制、侦测、变量、画笔等积木使用情况
- 广播消息
- 变量 / 列表
- 扩展使用情况
- 参考作品的关键教学点

建议把分析结果做成结构化 JSON，而不是只存一段文本。原因很直接：后面不管是做 AI 提示、看板摘要，还是做规则兜底，都需要结构化字段。

## 10. DeepSeek 提示链路

学生客户端不直接调用 DeepSeek。

正式链路应当是：

1. 学生客户端请求提示
2. 服务端读取该学生该任务最近一次进度
3. 服务端读取教师参考 `sb3` 的分析结果
4. 服务端拼装 Prompt
5. 服务端调用 `DeepSeek API`
6. 服务端保存提示日志
7. 服务端把提示文本返回给学生客户端

Prompt 至少应包含：

- 教学任务标题与目标
- 教师上传的参考 `sb3` 分析摘要
- 学生当前角色 / 当前目标
- 学生最近一步进度说明
- 学生项目快照摘要
  - 当前有哪些角色
  - 每个角色有哪些积木
- 提示风格约束

提示风格建议固定为：

- 短
- 具体
- 只给下一步
- 不直接替学生写完整作品

服务端需要保存这些 DeepSeek 相关配置：

- `DEEPSEEK_BASE_URL`
- `DEEPSEEK_API_KEY`
- `DEEPSEEK_MODEL`

可选的存储配置：

- `DATABASE_URL`
- `SERVER_API_DB_PATH`
- `SB3_STORAGE_DIR`

为了避免上游偶发失败直接卡住学生端，第一阶段建议保留一层规则兜底：

- 如果 DeepSeek 超时或报错，服务端返回“基于当前目标和最近快照的最小下一步提示”

这样学生端体验不会因为第三方波动完全中断。

## 11. 学生进度上报

“及时上传”不等于“每秒上传一次”。第一阶段建议这样收口：

- 本地 Scratch 项目发生变化后，`3` 秒 debounce 后上报一次
- 学生主动请求提示前，先补传一次最新进度
- 持续编辑过程中，每 `15` 秒补一个心跳快照
- 切换角色、运行项目、保存项目时，立即触发一次上报

`POST /api/student/assignments/{id}/progress` 建议至少上传：

- `currentTarget`
- `stepSummary`
- `snapshot`
- `localProjectHash`
- `reportedAt`

`snapshot` 第一阶段建议包含：

- `currentRoleName`
- `roles`

`roles` 里每一项建议包含：

- `roleName`
- `roleType`
  - `stage` 或 `sprite`
- `blocks`
  - 按当前脚本遍历顺序上传该角色下的积木列表

第一阶段先不要扩到变量状态、进度百分比、脚本坐标、运行时值。先把“当前有哪些角色、每个角色有哪些积木”这层快照稳定下来。

建议的最小 JSON 形状：

```json
{
  "currentRoleName": "Cat",
  "roles": [
    {
      "roleName": "Stage",
      "roleType": "stage",
      "blocks": [
        "当绿旗被点击",
        "广播 开始"
      ]
    },
    {
      "roleName": "Cat",
      "roleType": "sprite",
      "blocks": [
        "当接收到 开始",
        "移动 10 步",
        "如果碰到边缘就反弹"
      ]
    }
  ]
}
```

如果后面需要更强的比对能力，再在不破坏现有结构的前提下，为每个积木补 `opcode`、字段值和脚本分组信息。

教师 Web 不必直接消费原始 `snapshot`，而是消费服务端聚合后的看板数据。

## 12. 教师管理 Web

教师 Web 第一阶段只做管理和查看，不做 AI 运算。

最小页面：

- `/login`
- `/admin`
- `/admin/teachers`
- `/admin/students`
- `/admin/audit-logs`
- `/students`
- `/releases`
- `/releases/:id/live`

每个页面职责：

- `/login`
  - 管理员与教师共用登录入口
  - 管理员登录后跳转 `/admin`
  - 教师登录后跳转原有教学工作区
- `/admin`
  - 查看后台总览
- `/admin/teachers`
  - 创建、启停、重置教师账号
- `/admin/students`
  - 为指定教师创建学生
  - 全局查看、启停、重置学生密码
- `/admin/audit-logs`
  - 查看管理员敏感操作审计日志
  - 第一阶段先支持按 action 基础筛选
- `/students`
  - 单个创建、重置密码、查看状态
  - 支持下载 Excel 可打开的模板、粘贴表格数据并批量创建学生
- `/releases`
  - 上传 `sb3`、创建任务
  - 在同页查看详情与分析结果
  - 分配学生、发布/归档
- `/releases/:id/live`
  - 查看学生最新进度、最新提示、更新时间

补充口径：

- 管理员页面是同一套 Web 应用里的独立后台路由，不是单独前端项目
- 当前发布单详情没有单独拆成 `/releases/:id` 页面，而是收在 `/releases` 的详情面板里
- 当前没有开放 Web 自助教师注册页；首次教师注册需调用 `POST /api/teacher/register`，或在管理员后台上线后由管理员创建
- 教师访问管理员接口时，后端应返回 `403`

## 13. 存储与部署

第一阶段推荐依赖：

- `Go API`
  - 核心服务进程
- `SQLite` / `Postgres`
  - 结构化数据
- 本地文件目录或对象存储
  - 保存原始 `sb3`
- `DeepSeek API`
  - 提示生成
- `Teacher Web`
  - 教师管理界面

本地开发可以先这样起步：

- 数据库：本地 `SQLite`
- 文件存储：本地目录
- 教师 Web：Vite 开发模式
- Go API：单进程

教师 Web 联调真实 API 时，当前实现已经支持浏览器侧 `CORS` 预检请求。

预发布时建议：

- API 单独部署
- Web 单独部署
- `sb3` 文件不要混在临时容器文件系统里
- 线上测试环境和正式环境分开数据库、分开持久目录

如果要直接按当前仓库落地部署，后续以 [`./deployment.zh-CN.md`](./deployment.zh-CN.md) 为准。那份文档已经明确了：

- `Vercel Preview / 固定 staging Web 域名 -> Zeabur staging -> Neon staging`
- `Vercel Production -> Zeabur production -> Neon production`
- `staging` 和 `production` 各自独立 `DATABASE_URL`、`SB3_STORAGE_DIR` 和 `CORS_ALLOWED_ORIGINS`

## 14. 验收路径

Happy path：

1. 教师注册并登录
2. 教师批量创建 30 个学生
3. 教师上传 1 个参考 `sb3`
4. 服务端成功生成分析结果
5. 教师把任务分配给学生
6. 学生在客户端成功登录
7. 学生能看到自己的任务
8. 学生编辑本地 Scratch 项目后成功上报进度
9. 学生请求提示后收到 DeepSeek 返回的下一步建议
10. 教师看板看到该学生的最新进度与最新提示

错误路径：

- 重复创建学生账号
- 非教师 token 调教师接口
- 学生访问未分配给自己的任务
- `sb3` 文件损坏或不是合法 ZIP
- 参考 `sb3` 尚未分析完成时，教师尝试发布任务
- 参考 `sb3` 尚未分析完成时，学生请求提示
- DeepSeek 超时或返回异常

边界路径：

- 批量创建时部分成功、部分失败
- 学生离线后恢复联网再补传进度
- 同一个学生短时间多次请求提示
- 教师归档任务后学生端的可见性规则

## 15. 分阶段开发建议

### 阶段 A：鉴权与学生管理

- 教师注册、登录、退出
- 学生单个创建
- 学生批量创建
- 学生客户端登录

### 阶段 B：任务与 sb3 分析

- 教师上传 `sb3`
- 任务创建、分配、发布
- `sb3` 异步解压与 `project.json` 分析
- 分析状态轮询与失败重试

### 阶段 C：学生进度与提示

- 学生任务列表
- 进度上报
- DeepSeek 提示生成
- 提示日志保存

### 阶段 D：教师看板

- 任务实时看板
- 学生历史进度查看
- 错误态与延迟态展示

当前建议先按 `A -> B -> C -> D` 的顺序推进，不要一开始就同时做“课程模型、推送、对象存储抽象、复杂权限中心”。
