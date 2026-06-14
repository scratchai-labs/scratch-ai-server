# 服务器端部署指南

这份文档只回答一件事：当前仓库上线时，`server-api`、`server-web`、数据库和 `sb3` 文件应该怎么拆，怎么配，怎么验。

当前推荐拓扑：

```text
Vercel Preview / 固定 staging 域名
  -> Zeabur staging / apps/server-api
  -> Neon staging
  -> Zeabur staging 持久卷

Vercel Production / 正式域名
  -> Zeabur production / apps/server-api
  -> Neon production
  -> Zeabur production 持久卷
```

## 1. 环境边界

建议直接固定三层环境，不要把“线上测试”和“正式”混在一起：

- 本地开发
  - `server-api` 默认可用 `SQLite`
  - `server-web` 默认可用 mock client
- `staging`
  - 用于真实浏览器联调、提测、部署演练
  - 必须有独立 `DATABASE_URL`
  - 必须有独立 `SB3_STORAGE_DIR`
- `production`
  - 只承载正式教师和学生数据
  - 只连接正式数据库和正式持久卷

“线上测试环境”建议直接等同于 `staging`。不要让测试流量写到正式数据库。

## 2. 为什么要两套数据库

当前项目线上建议至少两套数据库连接：

- `DATABASE_URL` for `staging`
- `DATABASE_URL` for `production`

原因很直接：

- `staging` 可以真实跑登录、上传 `sb3`、发布任务、进度上报，不会污染正式数据
- `production` 可以保持干净，避免测试教师、测试学生、测试进度混入正式课堂
- 当前服务会在启动时初始化表结构，先让 `staging` 首启验证，再上 `production` 更稳
- 发生回滚时，`staging` 可以继续试，`production` 不用跟着一起冒险

如果你用的是 `Neon`，从应用视角看，关键不是“必须两个项目”还是“必须两个 branch”，而是最终要拿到两条完全独立的连接串。`server-api` 只认 `DATABASE_URL`，不认环境名。

## 3. 推荐资源映射

### 3.1 后端

- `Zeabur staging`
  - 根目录：`apps/server-api`
  - 域名：比如 `https://api-staging.example.com`
  - 数据库：`Neon staging DATABASE_URL`
  - 文件目录：独立持久卷，比如 `/data/staging-sb3`
- `Zeabur production`
  - 根目录：`apps/server-api`
  - 域名：比如 `https://api.example.com`
  - 数据库：`Neon production DATABASE_URL`
  - 文件目录：独立持久卷，比如 `/data/prod-sb3`

### 3.2 前端

- `Vercel Preview`
  - 指向 `staging API`
  - 建议绑定固定 staging 域名，比如 `https://teacher-staging.example.com`
- `Vercel Production`
  - 指向 `production API`
  - 域名：比如 `https://teacher.example.com`

### 3.3 为什么强调“固定 staging 域名”

当前后端 `CORS_ALLOWED_ORIGINS` 是精确白名单，不支持 `*`，也不应该在 `release` 模式下放开。

这意味着：

- 如果直接拿每次都变化的 `*.vercel.app` Preview URL 去联调真实后端，`CORS` 会很难维护
- 更稳的做法是给 staging Web 一个固定域名，然后把这个域名写进 staging API 的 `CORS_ALLOWED_ORIGINS`

## 4. 必配环境变量

### 4.1 `server-api` on Zeabur

两套环境都至少要配这些变量：

- `GIN_MODE=release`
- `PORT=8000`
- `DATABASE_URL=<对应环境的 Neon 连接串>`
- `CORS_ALLOWED_ORIGINS=<对应环境允许的前端域名>`
- `SB3_STORAGE_DIR=<对应环境的持久卷目录>`
- `ADMIN_BOOTSTRAP_USERNAME=<首个管理员账号，可选>`
- `ADMIN_BOOTSTRAP_PASSWORD=<首个管理员密码，可选>`
- `DEEPSEEK_BASE_URL=https://api.deepseek.com`
- `DEEPSEEK_API_KEY=<真实密钥，若要走真实提示链路>`
- `DEEPSEEK_MODEL=deepseek-v4-flash`
- `DEEPSEEK_TIMEOUT_SECONDS=8`

推荐值示例：

```env
# staging
GIN_MODE=release
PORT=8000
DATABASE_URL=postgresql://...
CORS_ALLOWED_ORIGINS=https://teacher-staging.example.com
SB3_STORAGE_DIR=/data/staging-sb3

# production
GIN_MODE=release
PORT=8000
DATABASE_URL=postgresql://...
CORS_ALLOWED_ORIGINS=https://teacher.example.com
SB3_STORAGE_DIR=/data/prod-sb3
```

注意：

- `GIN_MODE=release` 下，缺少 `DATABASE_URL`、`SB3_STORAGE_DIR` 或 `CORS_ALLOWED_ORIGINS`，服务会直接启动失败
- `/health` 现在会做数据库 `Ping`，数据库不通时会返回 `503`
- `SB3_STORAGE_DIR` 必须挂到持久卷，不能放在临时容器文件系统里
- 若配置了 `ADMIN_BOOTSTRAP_USERNAME` + `ADMIN_BOOTSTRAP_PASSWORD`，服务启动时会自动创建或提升该管理员账号，适合部署后首次进入教师管理后台

### 4.2 `server-web` on Vercel

`Preview` 和 `Production` 都必须显式配置：

- `VITE_SERVER_WEB_API_MODE=real`
- `VITE_SERVER_WEB_API_BASE_URL=<对应环境的 API 域名>`

推荐值示例：

```env
# Vercel Preview
VITE_SERVER_WEB_API_MODE=real
VITE_SERVER_WEB_API_BASE_URL=https://api-staging.example.com

# Vercel Production
VITE_SERVER_WEB_API_MODE=real
VITE_SERVER_WEB_API_BASE_URL=https://api.example.com
```

当前前端构建已经加了 fail-fast：

- 生产构建缺少 `VITE_SERVER_WEB_API_MODE=real` 会直接失败
- 生产构建缺少 `VITE_SERVER_WEB_API_BASE_URL` 会直接失败

### 4.3 部署后首用口径

部署完成后，推荐按下面顺序首次启用：

1. 先确认 `ADMIN_BOOTSTRAP_USERNAME` 和 `ADMIN_BOOTSTRAP_PASSWORD` 已注入到当前环境的 `server-api`
2. 访问 `https://<web-domain>/login`
3. 使用管理员账号登录
4. 登录成功后进入 `/admin`
5. 在 `/admin/teachers` 创建或维护教师账号
6. 在 `/admin/students` 为指定教师创建学生账号

补充说明：

- 管理员页面不是单独域名，也不是单独前端项目；仍然走同一个 `server-web`
- 教师和管理员共用 `/login`，只是登录成功后的跳转和可见路由不同
- 如果当前环境还没有管理员，但必须先建立一个教师，也可以直接调用 `POST /api/teacher/register` 或使用 Swagger
- 生产环境更推荐先自举管理员，再由管理员统一创建教师账号，避免线上出现失控的自助注册
- 测试数据库建议直接使用 `staging DATABASE_URL`；不要让任何联调、验收或演示数据写进 production 数据库

## 5. 部署顺序

建议按这个顺序做，不要前后乱跳：

1. 先准备数据库
   - 创建 `Neon staging`
   - 创建 `Neon production`
   - 记下两条独立 `DATABASE_URL`
2. 再起 `Zeabur staging`
   - 根目录指向 `apps/server-api`
   - 挂 staging 持久卷
   - 配 staging 环境变量
   - 首次部署后先测 `/health`
3. 再起 `Vercel Preview` 或固定 staging Web
   - 根目录指向 `apps/server-web`
   - 配 `Preview` 环境变量，指向 staging API
   - 用固定 staging 域名做真实联调
4. staging 走通验收路径
   - 教师登录
   - 学生列表
   - 发布单列表
   - 上传 `sb3`
   - `/releases/:id/live` 实时看板
5. 再复制到正式环境
   - 新建或复制 `Zeabur production`
   - 替换成 production `DATABASE_URL`、正式域名、正式持久卷
   - 配好 `Vercel Production`
6. 最后切正式流量
   - 先验 API
   - 再验 Web
   - 再让真实教师使用

## 6. 上线前验证清单

部署前先在仓库里跑一遍：

```bash
npm run test
VITE_SERVER_WEB_API_MODE=real VITE_SERVER_WEB_API_BASE_URL=https://api.example.com npm run build
npm run server:api:docs:check
npm run server:web:smoke:mock
```

部署后至少验这些：

- `GET https://<api-domain>/health` 返回 `200`
- 教师能登录
- 教师能打开学生列表
- 教师能打开发布单列表
- 教师能上传一个合法 `sb3`
- 教师能进入实时看板
- staging 数据不会出现在 production

## 7. 回滚建议

如果 staging 出问题，直接修，不要碰 production。

如果 production 出问题：

- `Vercel` 先回退到上一个稳定部署
- `Zeabur` 回退到上一个稳定版本或稳定 commit
- 不要把 production 数据库临时切回 staging
- 不要把 production 的 `SB3_STORAGE_DIR` 指到 staging 持久卷

## 8. 当前项目的部署底线

当前仓库上线时，至少要满足这些底线：

- `server-api` 和 `server-web` 分开部署
- `staging` 和 `production` 分开数据库
- `staging` 和 `production` 分开 `SB3` 持久目录
- `CORS_ALLOWED_ORIGINS` 只写明确域名，不用 `*`
- `server-web` 的生产构建只连真实 API，不回退 mock

如果下一次对话直接开始部署，可以把这份文档当成部署真值源，按这里的拓扑和变量清单逐项落地。
