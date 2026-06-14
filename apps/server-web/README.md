# Scratch 教师后台

`apps/server-web` 是服务器端第一阶段的 Vue 3 + Vite 教师后台 SPA。

更完整的服务器端说明见 [`../../docs/server-development.zh-CN.md`](../../docs/server-development.zh-CN.md)。
实际部署时的环境拆分、变量矩阵和上线顺序见 [`../../docs/deployment.zh-CN.md`](../../docs/deployment.zh-CN.md)。

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
- `admin`
- `admin12345`

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

管理员账号登录后，会进入 `/admin`，通过 `GET /api/admin/overview`、`GET /api/admin/teachers`、`POST /api/admin/teachers/{id}/role`、`GET /api/admin/students`、`POST /api/admin/students` 以及对应的启停/重置接口统一维护账号；教师管理页支持直接切换教师/管理员角色，学生管理页支持直接为指定教师创建学生账号。

## 登录与路由

- Web 对外统一入口是 `/login`
- 管理员和教师共用同一登录页，不需要单独的管理员域名
- 管理员登录后进入 `/admin`，再通过 `/admin/teachers` 和 `/admin/students` 维护教师与学生
- 教师登录后进入原有教学管理页面，不会进入管理员页面
- 当前 Web 没有教师自助注册页；首次教师注册需走 `POST /api/teacher/register`，生产环境更推荐由管理员在 `/admin/teachers` 创建
- 教师访问管理员接口时，后端会返回 `403`

如果只想在子目录里单独调试，也可以：

```bash
cd apps/server-web
npm run dev
```

## 约定接口

- `POST /api/teacher/login`
- `POST /api/teacher/register`
- `GET /api/teacher/students`
- `GET /api/teacher/assignments`
- `GET /api/teacher/dashboard/students/:id/history`
- `GET /api/teacher/dashboard/assignments/:id/live`

## Vercel 预发布部署

- 推荐把项目根目录指向 `apps/server-web`
- 构建命令使用 `npm run build`
- 输出目录使用 `dist`
- `Preview` 建议指向 `staging API`
- `Production` 指向正式 API
- 需要注入：
  - `VITE_SERVER_WEB_API_MODE=real`
  - `VITE_SERVER_WEB_API_BASE_URL=https://<your-zeabur-api-domain>`
- 如果要用真实 API 做 staging 联调，建议给 Web 绑定固定 staging 域名，不要依赖每次变化的 Preview URL
- 生产环境不再自动展示 mock 登录提示，也不会回退到 mock client
- 仓库已提供 `vercel.json`，用于把 Vue Router history 路由统一回写到 `index.html`
