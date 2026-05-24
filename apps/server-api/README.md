
# Scratch AI Server API

`apps/server-api` 是服务器端第一阶段的 FastAPI 后端，负责老师/学生认证、发布单、学生进度、AI 提示和教师实时看板接口。

更完整的服务器端说明见 [`../../docs/server-development.zh-CN.md`](../../docs/server-development.zh-CN.md)。

## 本地开发

优先从仓库根目录运行：

```bash
npm run server:api:test
uv run --project apps/server-api python -m app.main
```

如需单独进入目录，也可以使用：

```bash
cd apps/server-api
uv run --python python3 pytest
```

## 关键接口

- `GET /health`
- `POST /api/teacher/register`
- `POST /api/teacher/login`
- `POST /api/student/login`
- `GET /api/students`
- `POST /api/students`
- `GET /api/releases`
- `POST /api/releases`
- `POST /api/student/releases/{releaseId}/progress`
- `POST /api/student/releases/{releaseId}/hints`
- `GET /api/dashboard/releases/{releaseId}/live`

## 环境变量

- `DATABASE_URL`
- `SERVER_API_DB_PATH`
- `CORS_ALLOWED_ORIGINS`
- `AI_PROVIDER`
- `AI_BASE_URL`
- `AI_API_KEY`
- `AI_MODEL`

规则：

- 生产部署优先使用 `DATABASE_URL`
- 未配置 `DATABASE_URL` 时，默认回退到本地 SQLite `SERVER_API_DB_PATH`
- `CORS_ALLOWED_ORIGINS` 使用逗号分隔，至少应包含教师后台的 Vercel 域名

## Zeabur 预发布部署

- 推荐把服务根目录指向 `apps/server-api`
- 仓库已提供 `zbpack.json`
  - 固定 Python `3.11`
  - 固定使用 `uv`
  - 入口固定为 `app/main.py`
- 需要在 Zeabur 注入：
  - `DATABASE_URL`
  - `CORS_ALLOWED_ORIGINS`
  - `AI_PROVIDER`
  - `AI_BASE_URL`
  - `AI_API_KEY`
  - `AI_MODEL`
- 部署后先验证：

```bash
curl https://<your-zeabur-api-domain>/health
```
