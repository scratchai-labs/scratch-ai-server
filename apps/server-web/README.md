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
npm run server:web:smoke:mock
```

最近一轮基于 mock / fake data 的浏览器点击验证记录见 [`../../docs/server-web-mock-smoke-test.zh-CN.md`](../../docs/server-web-mock-smoke-test.zh-CN.md)。

首次执行浏览器 smoke test 前，如果本机还没有 Playwright 的 Chromium，可在子目录执行一次：

```bash
cd apps/server-web
npx playwright install chromium
```

## API 模式

默认使用本地 mock client，可以直接开发和跑页面。

mock 模式下可直接用这组演示账号登录：

- `teacher`
- `teach123`

如果要切到真实后端，设置环境变量：

```bash
VITE_SERVER_WEB_API_MODE=real
VITE_SERVER_WEB_API_BASE_URL=http://localhost:8000
```

仓库里也提供了 [`./.env.example`](./.env.example) 作为最小示例。

生产构建约束：

- `VITE_SERVER_WEB_API_MODE` 在生产环境必须是 `real`
- `VITE_SERVER_WEB_API_BASE_URL` 在生产环境必须显式配置
- 若缺失上述变量，`vite build` 期间会直接失败，不再静默回退到 mock 或同源 `/api`

当前真实 API 模式下，教师总览和学生管理会先请求 `GET /api/teacher/students`，再按学生补拉 `GET /api/teacher/dashboard/students/:id/history`，用最近一条学习历史渲染真实的 `status / currentTarget / stepSummary / latestAiHint / updatedAt`。

如果只想在子目录里单独调试，也可以：

```bash
cd apps/server-web
npm run dev
```

## 约定接口

- `POST /api/teacher/login`
- `GET /api/teacher/students`
- `GET /api/teacher/assignments`
- `GET /api/teacher/dashboard/students/:id/history`
- `GET /api/teacher/dashboard/assignments/:id/live`

## Vercel 预发布部署

- 推荐把项目根目录指向 `apps/server-web`
- 构建命令使用 `npm run build`
- 输出目录使用 `dist`
- 需要注入：
  - `VITE_SERVER_WEB_API_MODE=real`
  - `VITE_SERVER_WEB_API_BASE_URL=https://<your-zeabur-api-domain>`
- 生产环境不再自动展示 mock 登录提示，也不会回退到 mock client
- 仓库已提供 `vercel.json`，用于把 Vue Router history 路由统一回写到 `index.html`
