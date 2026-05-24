# Scratch 教师后台

`apps/server-web` 是服务器端第一阶段的 Vue 3 + Vite 教师后台 SPA。

更完整的服务器端说明见 [`../../docs/server-development.zh-CN.md`](../../docs/server-development.zh-CN.md)。

## 运行

```bash
npm run server:web:dev
```

## 测试

```bash
npm run server:web:test
npm run server:web:build
```

## API 模式

默认使用本地 mock client，可以直接开发和跑页面。

如果要切到真实后端，设置环境变量：

```bash
VITE_SERVER_WEB_API_MODE=real
VITE_SERVER_WEB_API_BASE_URL=http://localhost:8000
```

仓库里也提供了 [`./.env.example`](./.env.example) 作为最小示例。

如果只想在子目录里单独调试，也可以：

```bash
cd apps/server-web
npm run dev
```

## 约定接口

- `POST /api/teacher/login`
- `GET /api/students`
- `GET /api/releases`
- `GET /api/dashboard/releases/:id/live`

## 默认演示账号

- `teacher`
- `teach123`

## Vercel 预发布部署

- 推荐把项目根目录指向 `apps/server-web`
- 构建命令使用 `npm run build`
- 输出目录使用 `dist`
- 需要注入：
  - `VITE_SERVER_WEB_API_MODE=real`
  - `VITE_SERVER_WEB_API_BASE_URL=https://<your-zeabur-api-domain>`
- 仓库已提供 `vercel.json`，用于把 Vue Router history 路由统一回写到 `index.html`
