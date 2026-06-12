
# Scratch AI Server API

`apps/server-api` 是当前阶段基于 `Gin` 的 `Go` 服务端，负责老师/学生认证、任务上传与分配、学生进度、提示生成和教师实时看板接口。机器可读 API 契约以生成的 OpenAPI 为准。

客户端接入指南见 [`../../docs/server-api-contract.zh-CN.md`](../../docs/server-api-contract.zh-CN.md)。
更完整的服务器端开发说明见 [`../../docs/server-development.zh-CN.md`](../../docs/server-development.zh-CN.md)。

## 本地开发

优先直接在当前目录使用 `Go` 命令：

```bash
cd apps/server-api
go test ./...
go run ./cmd/server-api
```

如果需要从仓库根目录统一调度，再用这些 `npm run` 快捷命令：

```bash
npm run server:api:test
npm run server:api:dev
npm run server:api:docs
```

## OpenAPI 文档

当前 `Go` 主线直接从代码注释和类型定义生成 OpenAPI 规格，`docs/swagger.json`、`docs/swagger.yaml`、`docs/docs.go` 是机器可读契约真值源。手写的 `server-api-contract.zh-CN.md` 只保留接入指南和补充说明，不再作为字段定义的单一事实源。

- 运行服务后可访问：
  - `http://127.0.0.1:8000/swagger/index.html`
  - `http://127.0.0.1:8000/swagger/doc.json`
- 仓库内生成产物：
  - `docs/swagger.json`
  - `docs/swagger.yaml`
  - `docs/docs.go`
- 常用命令：

```bash
npm run server:api:docs
npm run server:api:docs:check
```

约定：

- 修改 `internal/http` 的路由、请求体或响应体后，要重新生成 OpenAPI
- 客户端对接优先以生成出的 OpenAPI 为准
- `server-api-contract.zh-CN.md` 只保留接入指南和补充说明，不再作为接口字段的单一事实源

## 关键接口

- `GET /health`
- `POST /api/teacher/register`
- `POST /api/teacher/login`
- `GET /api/teacher/me`
- `POST /api/teacher/logout`
- `GET /api/teacher/students`
- `POST /api/teacher/students`
- `POST /api/teacher/students/batch`
- `POST /api/teacher/students/{id}/reset-password`
- `GET /api/teacher/assignments`
- `POST /api/teacher/assignments`
- `GET /api/teacher/assignments/{id}`
- `GET /api/teacher/assignments/{id}/analysis`
- `POST /api/teacher/assignments/{id}/assign-students`
- `POST /api/teacher/assignments/{id}/publish`
- `POST /api/teacher/assignments/{id}/archive`
- `GET /api/teacher/dashboard/assignments/{id}/live`
- `GET /api/teacher/dashboard/students/{id}/history`
- `POST /api/student/login`
- `GET /api/student/me`
- `POST /api/student/logout`
- `GET /api/student/assignments`
- `GET /api/student/assignments/{id}`
- `POST /api/student/assignments/{id}/progress`
- `POST /api/student/assignments/{id}/hints`

## 环境变量

- `PORT`
- `GIN_MODE`
- `DATABASE_URL`
- `SERVER_API_DB_PATH`
- `CORS_ALLOWED_ORIGINS`
- `SB3_STORAGE_DIR`
- `DEEPSEEK_API_KEY`
- `DEEPSEEK_BASE_URL`
- `DEEPSEEK_MODEL`
- `DEEPSEEK_TIMEOUT_SECONDS`

规则：

- 默认监听 `:8000`
- 未配置 `DATABASE_URL` 时，默认使用本地 `SQLite`
- 配置 `DATABASE_URL` 后，自动切到 `Postgres`
- `CORS_ALLOWED_ORIGINS` 配置后，会按白名单回写 `Access-Control-Allow-Origin`
- 默认把原始 `sb3` 保存到 `SB3_STORAGE_DIR`，未配置时使用系统临时目录下的 `scratch-ai-server-sb3`
- 配置了 `DEEPSEEK_API_KEY` 后，学生提示链路会优先走真实 `DeepSeek`，失败时自动回退到本地 fallback
- 当 `GIN_MODE=release` 时，必须显式提供 `DATABASE_URL`、`SB3_STORAGE_DIR` 和 `CORS_ALLOWED_ORIGINS`

## Zeabur 预发布部署

- 推荐把服务根目录指向 `apps/server-api`
- 需要在 Zeabur 注入：
  - `GIN_MODE=release`
  - `PORT`
  - `DATABASE_URL`
  - `CORS_ALLOWED_ORIGINS=https://<your-vercel-domain>`
  - `SB3_STORAGE_DIR`
  - `DEEPSEEK_API_KEY`
  - `DEEPSEEK_BASE_URL`
  - `DEEPSEEK_MODEL`
  - `DEEPSEEK_TIMEOUT_SECONDS`
- 若使用 `Neon Postgres`，把 `DATABASE_URL` 指向 `Neon` 提供的连接串；不要在 `release` 模式下回退到临时 `SQLite`
- `SB3_STORAGE_DIR` 应该指向 Zeabur 持久卷目录，避免重启后丢失上传的原始 `sb3`
- 部署后先验证：

```bash
curl https://<your-zeabur-api-domain>/health
```
