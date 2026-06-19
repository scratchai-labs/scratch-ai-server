# 教师后台 Mock 浏览器 Smoke Test

本文记录一次基于本地 mock / fake data 的教师后台浏览器模拟点击验证，目标是确认 `server-web` 在不依赖真实后端的情况下，班级优先主链路可登录、可导航、可渲染。

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
3. 验证班级管理首屏数据
4. 从侧边栏进入 `Dashboard`
5. 回到班级管理并进入班级详情
6. 从班级详情进入项目详情

## 验证结果

### 1. 登录页

- 表单可输入
- 使用 `teacher / teach123` 登录成功
- 登录成功后自动跳转到 `班级管理`

### 2. 班级管理

- 当前教师：`王老师`
- 班级列表共 `2` 个
- `四年级一班`：`2 名学生 · 1 个项目`
- `四年级二班`：`1 名学生 · 1 个项目`

### 3. Dashboard

- 页面可正常进入
- 仍能看到 `Ada` 的最新提示与 `第一期发布单` 的摘要卡片

### 4. 班级详情

- 页面可正常进入
- 默认先展示学生管理和批量导入区域
- 可看到 `Ada / Mia` 两名学生
- 同页可看到项目管理与 `迷宫项目`

### 5. 项目详情

- 可从班级详情点击进入 `迷宫项目`
- 页面能看到项目概览、教学点和学生当前进度与提示
- `Ada` 当前提示为 `先把绿旗事件连起来`

## 结论

- `班级管理 -> Dashboard -> 班级详情 -> 项目详情` 主链路在 mock / fake data 模式下全部可点击、可跳转、可渲染
- 页面展示数据与 `mockTeacherApi.ts` 内的演示数据一致

## 非阻塞问题

- 浏览器控制台曾出现 `favicon.ico 404`
- 该问题不影响主链路功能，但会产生控制台噪音
- 已通过给前端入口补 `favicon.svg` 处理

## 后续建议

- 若需要对接真实后端，再补一轮 `VITE_SERVER_WEB_API_MODE=real` 的浏览器 smoke test
- 若要把这条链路接入 CI，可直接复用 `npm run server:web:smoke:mock`
