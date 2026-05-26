
# Scratch AI Server API

`apps/server-api` 是当前阶段的 Go 服务端，基于 `Gin` 实现老师/学生认证、任务上传与分配、学生进度、提示生成和教师实时看板接口。

更完整的服务器端说明见 [`../../docs/server-development.zh-CN.md`](../../docs/server-development.zh-CN.md)。

## 本地开发

优先从仓库根目录运行：

```bash
npm run server:api:test
npm run server:api:dev
```

如需单独进入目录，也可以使用：

```bash
cd apps/server-api
go test ./...
go run ./cmd/server-api
```

## 关键接口

- `GET /health`
- `POST /api/teacher/register`
- `POST /api/teacher/login`
- `GET /api/teacher/me`
- `POST /api/teacher/logout`
- `GET /api/teacher/students`
- `POST /api/teacher/students/batch`
- `POST /api/teacher/students/{studentId}/reset-password`
- `GET /api/teacher/assignments`
- `POST /api/teacher/assignments`
- `GET /api/teacher/assignments/{assignmentId}`
- `GET /api/teacher/assignments/{assignmentId}/analysis`
- `POST /api/teacher/assignments/{assignmentId}/assign-students`
- `POST /api/teacher/assignments/{assignmentId}/publish`
- `POST /api/teacher/assignments/{assignmentId}/archive`
- `GET /api/teacher/dashboard/assignments/{assignmentId}/live`
- `GET /api/teacher/dashboard/students/{studentId}/history`
- `POST /api/student/login`
- `GET /api/student/me`
- `POST /api/student/logout`
- `GET /api/student/assignments`
- `GET /api/student/assignments/{assignmentId}`
- `POST /api/student/assignments/{assignmentId}/progress`
- `POST /api/student/assignments/{assignmentId}/hints`

## 环境变量

- `PORT`
- `DATABASE_URL`
- `SERVER_API_DB_PATH`
- `SB3_STORAGE_DIR`
- `DEEPSEEK_API_KEY`
- `DEEPSEEK_BASE_URL`
- `DEEPSEEK_MODEL`
- `DEEPSEEK_TIMEOUT_SECONDS`

规则：

- 默认监听 `:8000`
- 未配置 `DATABASE_URL` 时，默认使用本地 `SQLite`
- 配置 `DATABASE_URL` 后，自动切到 `Postgres`
- 默认把原始 `sb3` 保存到 `SB3_STORAGE_DIR`，未配置时使用系统临时目录下的 `scratch-ai-server-sb3`
- 配置了 `DEEPSEEK_API_KEY` 后，学生提示链路会优先走真实 `DeepSeek`，失败时自动回退到本地 fallback
- 当前实现支持教师 Web 本地真实联调所需的 `CORS` 预检请求

## Zeabur 预发布部署

- 推荐把服务根目录指向 `apps/server-api`
- 需要在 Zeabur 注入：
  - `PORT`
  - `DATABASE_URL` 或 `SERVER_API_DB_PATH`
  - `SB3_STORAGE_DIR`
  - `DEEPSEEK_API_KEY`
  - `DEEPSEEK_BASE_URL`
  - `DEEPSEEK_MODEL`
  - `DEEPSEEK_TIMEOUT_SECONDS`
- 部署后先验证：

```bash
curl https://<your-zeabur-api-domain>/health
```
