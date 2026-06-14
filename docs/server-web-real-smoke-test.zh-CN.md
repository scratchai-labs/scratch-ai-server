# 教师后台 Real-Mode 浏览器 Smoke Test

本文记录当前教师后台真实 API 自举冒烟脚本的用途、执行方式和覆盖范围，作为 mock 冒烟之外的第二条验证链路。

## 测试背景

- 测试日期：`2026-06-14`
- 目标应用：`apps/server-web`
- 运行模式：`VITE_SERVER_WEB_API_MODE=real`
- 执行方式：`npm run server:web:smoke:real`
- 对应脚本：`apps/server-web/scripts/real-smoke-test.mjs`

## 脚本会自动做什么

- 分配临时端口，避免撞上本机已运行的 `server-api` 或 `Vite`
- 启动临时 `server-api`
- 使用临时 `SQLite` 和临时 `SB3_STORAGE_DIR`
- 自动生成样例 `sb3`
- 自动注册教师账号
- 启动 real-mode 的 `server-web`
- 用 Playwright 在浏览器里走完整教师链路

脚本结束后会自动清理临时目录和后台进程。

## 前置条件

- 本机已安装 `Go`
- 本机已安装 Playwright 的 Chromium

如果本机还没有 Chromium，可先执行：

```bash
cd apps/server-web
npx playwright install chromium
```

## 重复执行

```bash
npm run server:web:smoke:real
```

## 覆盖链路

按当前脚本，链路顺序如下：

1. 教师注册并登录 Web
2. 进入学生页，新建学生
3. 在学生列表里重置密码
4. 进入发布单页，上传样例 `sb3`
5. 等待分析完成并查看详情
6. 分配学生
7. 发布发布单
8. 进入实时看板
9. 学生通过真实 API 登录
10. 学生上报进度并请求提示
11. 教师实时看板看到最新进度和提示
12. 回到发布单页归档任务

## 验证重点

- 页面不应出现未捕获报错
- 不应出现失败请求
- 发布单分析必须进入 `ready`
- 学生进度与提示必须能回流到教师实时看板
- 归档后发布单状态必须更新

## 当前结果

最近一轮执行已通过，未发现明显阻断问题。

补充说明：

- 若环境里配置了 `DEEPSEEK_API_KEY`，提示链路会优先走真实 `DeepSeek`
- 若未配置，脚本仍会验证 `fallback` 提示链路
