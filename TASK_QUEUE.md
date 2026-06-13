# TASK_QUEUE

## 待确认

- 当前无待确认任务。

## 已完成

- 2026-06-13：完成登录首页结构重做：将 `LoginView` 从左右双栏改为 `header / main / footer` 的上中下结构，页头补齐品牌标识、运行模式和开发说明入口，中间收口为紧凑 hero + 居中登录卡，页脚补齐课堂场景、联调模式和文档支持三段说明；同步新增组件测试锁定首页骨架与 `Mock API` 信息。已通过 `npm run server:web:test`、`npm run server:web:smoke:mock`，并在注入占位真实环境变量后通过 `npm run server:web:build`。
- 2026-06-13：根据首页反馈完成登录页 `auth-hero` 二次收口：将登录页双栏整体居中，缩小左侧说明区列宽与内边距，取消其卡片阴影、下调标题与段落尺寸，并保持右侧登录卡为主要焦点；本次仅改 `server-web` 首页视觉布局，不改登录行为与其他页面。已通过 `npm run server:web:test`、`npm run server:web:smoke:mock`，并在注入占位真实环境变量后通过 `npm run server:web:build`。
- 2026-06-13：修正 `.impeccable/live/config.json` 的入口路径为 `apps/server-web/index.html`，解决 monorepo 根目录下执行 `$impeccable live` 时注入失败的问题；当前教师端 live 预览可通过本地 `Vite` 开发服务器正常加载。
- 2026-06-13：完成 `$impeccable polish apps/server-web/src/styles.css`：在不改页面结构的前提下，收口教师后台共享样式系统，降低全局装饰噪音，统一容器圆角与阴影，移除条纹/发光式背景语言，补齐按钮/导航/输入的焦点态与 reduced motion，增强表格、空状态、登录页和实时看板的产品化层级；已通过 `npm run server:web:test`、`npm run server:web:smoke:mock`，并在注入占位真实环境变量后通过 `npm run server:web:build`。
- 2026-06-13：完成 `$impeccable init` 首轮初始化：新增根目录 `PRODUCT.md`，明确本仓库默认 `register=product`、教师课堂工具定位、品牌气质、反例、设计原则与默认 `WCAG AA` 无障碍口径；同时创建 `.impeccable/live/config.json`，按 `Vite SPA` 的 `index.html` 入口完成 live mode 预配置，并确认当前仓库未检测到需处理的 `CSP`。`DESIGN.md` 暂未生成，后续可按现有前端代码执行 `$impeccable document` 补齐视觉真值源。
- 2026-06-12：完成 `Vercel` 上 `server-web`（Teacher Web）部署：已按 `apps/server-web` 子目录完成构建与发布，前端已指向真实 API `https://scratchai.zeabur.app`，并进入部署后联调收尾阶段。
- 2026-06-12：完成 `Zeabur` 上 `server-api`（Go 服务）部署：已核对根目录、环境变量、`/data` 持久卷、`/health` 探活与教师注册写库链路，当前公网 API `https://scratchai.zeabur.app` 已可用。
- 2026-06-12：补齐部署文档，新增 `docs/deployment.zh-CN.md` 作为部署真值源，明确 `Vercel + Zeabur + Neon` 的 staging / production 拓扑、双数据库建议、固定 staging 域名与 `CORS` 约束、环境变量矩阵、上线顺序、回滚和验证清单；并同步更新根 README、`docs/README`、`server-development`、`server-api`、`server-web` 的部署入口。
- 2026-06-12：完成部署前第二轮多 agent code review 收口：把 `server-web` 的真实环境校验前移到 `vite build`，让缺失 `VITE_SERVER_WEB_API_MODE=real` / `VITE_SERVER_WEB_API_BASE_URL` 的生产构建直接失败；同步为 GitHub Actions 注入占位发布变量，避免 CI 假红；修复学生历史补拉把 `401` 吞成空历史的问题；并把 `server:web:smoke:mock` 改为直接跑 `Vite dev` mock 页面，补齐当前 UI 断言，确保 mock 冒烟验证恢复可用。
- 2026-06-12：完成部署前加固与上线准备：为 `server-api` 补上 `release` 配置校验、`CORS` 白名单、readiness 健康检查、`sb3` 上传限流和数据库写错误透传；为 `server-web` 收口真实环境 fail-fast、教师退出登录和 `401` 未授权处理；同步根级 `.env.example`、Zeabur / Vercel README 与 GitHub Actions 的 `Go` 工具链配置，并已通过根级 `npm run test` 与 `npm run build`。
- 2026-06-12：为独立部署前清理服务器端仓库历史与本地生成物：裁掉 `TASK_QUEUE.md` 中和当前仓边界无关的旧桌面端记录，删除本地 `node_modules`、`apps/server-web/node_modules` 与遗留 `server-api.sqlite3`，保持仓库对外只呈现 `server-api + server-web` 主线。
- 2026-06-12：收口仓库边界到 `apps/server-api` + `apps/server-web`，清理支持文档、维护文档、清理脚本、忽略规则、issue 模板和锁文件里残留的 `apps/desktop-companion` / `tools/verification` / `packages/shared` / `installers` 旧引用；当前仓库只保留服务器端 monorepo 入口。
- 2026-05-27：收口老师端多学生状态展示与差异化提示验证；`server-web` 的总览 / 学生管理现已基于 `GET /api/teacher/students` + `GET /api/teacher/dashboard/students/{id}/history` 汇总真实 `status / currentTarget / stepSummary / latestAiHint / updatedAt`，不再把真实课堂状态渲染成假 `0%`；同时新增后端回归测试，验证同一 assignment 下两名学生会因不同进度拿到不同提示，并可在老师实时看板同时看到。已通过 `server:test`，也已在真实老师/学生联调里复核。
- 2026-05-27：收口教师实时看板显示层与进度时间口径；为 `server-api` 的进度记录补 `reportedAt` 缺省兜底，避免学生端未显式上传时间时老师看板长期显示 `—`；同时让 `server-web` 的实时看板消费真实 API 已提供的 `status / currentTarget / stepSummary / lastHintAt`，不再把未提供百分比的真实课堂状态误渲染成假 `0%`。已补 `teacherApi` / `LiveReleaseView` 回归测试，并用 `/Users/tesths/Downloads/Cat and a Mouse.sb3` 实际跑通老师端与学生端多角色联调；`npm run server:test` 也已在本机通过。
- 2026-05-26：完成 `scratch-ai-server` 收尾 code review 并修复阻断项：补齐 `Go API` 多处必填字段校验，新增认证 / 学生创建 / 进度上报 / 密码重置回归测试；同步更新 `server-api/.env.example` 到当前 `DeepSeek + sb3` 配置口径，并忽略本地 `go build` 产物避免污染工作区；已通过 `server:api:docs:check`、根级 `npm run test` 和 `npm run server:build`。
- 2026-05-26：补齐服务器端与教师 Web 的真实联调收口：新增 API 侧 `CORS` 预检支持，完成教师 Web 真实浏览器点击验证，走通登录、学生列表、发布单列表、实时看板和退出登录；同时引入 `@faker-js/faker`，补强教师 Web fake data 测试与后端认证/隔离回归测试；已通过 `server-api` 全量测试、`server-web` 全量测试，以及 `go test ./tests -count=1 -coverpkg=./internal/...` 覆盖验证。
- 2026-05-24：完成从原始 workspace 拆出服务器端独立仓；保留 `server-api + server-web` 联机教学主线，收口根级 workspace、README、架构与部署文档，并完成独立 git 初始化、前后端测试与构建验证。
