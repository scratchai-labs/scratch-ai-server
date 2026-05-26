# 教师后台 Mock 浏览器 Smoke Test

本文记录一次基于本地 mock / fake data 的教师后台浏览器模拟点击验证，目标是确认 `server-web` 在不依赖真实后端的情况下，主链路可登录、可导航、可渲染、可轮询。

## 测试背景

- 测试日期：`2026-05-26`
- 目标应用：`apps/server-web`
- 运行模式：默认 mock client
- 数据来源：`src/services/mockTeacherApi.ts`
- 登录账号：
  - 用户名：`teacher`
  - 密码：`teach123`
- 执行方式：`npm run server:web:smoke:mock`

## 测试环境

- 本地 preview server：脚本自动分配端口
- 路由模式：`Vue Router history`
- API 模式：脚本强制使用 `VITE_SERVER_WEB_API_MODE=mock`

## 重复执行

```bash
npm run server:web:smoke:mock
```

如果本机还没有 Playwright 的 Chromium，可先执行：

```bash
cd apps/server-web
npx playwright install chromium
```

## 覆盖链路

按以下顺序执行：

1. 打开登录页
2. 输入 mock 账号并登录
3. 验证 Dashboard 首屏数据
4. 从 Dashboard 进入 `Students`
5. 从侧边栏进入 `Releases`
6. 点击 `第一期发布单` 的实时看板入口
7. 验证 `Live Release` 首帧数据和轮询后的第二帧数据

## 验证结果

### 1. 登录页

- 表单可输入
- 使用 `teacher / teach123` 登录成功
- 登录成功后自动跳转到 `Dashboard`

### 2. Dashboard

- 当前教师：`王老师`
- 在册学生：`3`
- 发布单：`1 / 2`
- 平均进度：`55%`
- 最新学生：`Ada`
- 最新学生进度：`72%`
- 最新学生提示：`补上广播消息后再测试一次`
- 最新发布单：`第一期发布单`
- 最新发布单状态：`已发布`

### 3. Students

- 页面可正常进入
- 列表共 `3` 人
- `Ada / Alan / Mia` 的班级、进度、AI 提示、更新时间均与 mock 数据一致

### 4. Releases

- 页面可正常进入
- 列表共 `2` 个发布单
- `第一期发布单`：状态 `已发布`，学生数 `24`
- `第二期发布单`：状态 `草稿`，学生数 `18`

### 5. Live Release

- 可从 `Releases` 点击进入 `rel-1`
- 首帧数据：
  - `Ada 42%`
  - `Alan 33%`
- 轮询更新后第二帧数据：
  - `Ada 68%`
  - `Alan 51%`
- 两帧提示文本均与 mock 数据一致
- 页面状态显示 `轮询中`，说明轮询链路正常

## 结论

- `Dashboard -> Students -> Releases -> Live Release` 主链路在 mock / fake data 模式下全部可点击、可跳转、可渲染
- 页面展示数据与 `mockTeacherApi.ts` 内的演示数据一致
- Live Release 的轮询更新行为正常

## 非阻塞问题

- 浏览器控制台曾出现 `favicon.ico 404`
- 该问题不影响主链路功能，但会产生控制台噪音
- 已通过给前端入口补 `favicon.svg` 处理

## 后续建议

- 若需要对接真实后端，再补一轮 `VITE_SERVER_WEB_API_MODE=real` 的浏览器 smoke test
- 若要把这条链路接入 CI，可直接复用 `npm run server:web:smoke:mock`
