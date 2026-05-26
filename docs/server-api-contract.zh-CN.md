# Server API 契约

本文面向客户端和教师端联调，直接以当前 `apps/server-api` 的真实实现为准，不再讨论需求草案。

如果你是学生客户端开发，优先看：

1. `POST /api/student/login`
2. `GET /api/student/assignments`
3. `GET /api/student/assignments/{id}`
4. `POST /api/student/assignments/{id}/progress`
5. `POST /api/student/assignments/{id}/hints`

## 1. 基础约定

### 1.1 Base URL

- 本地默认：`http://127.0.0.1:8000`
- 所有正式接口都带 `/api`

### 1.2 请求格式

- 普通接口使用 `application/json`
- 上传 `sb3` 使用 `multipart/form-data`
- 当前接口没有分页参数，列表接口一次返回全量结果

### 1.3 认证

- 教师接口使用教师 token
- 学生接口使用学生 token
- 请求头格式：

```http
Authorization: Bearer <token>
```

当前实现没有 refresh token，也没有 token 过期时间字段。`logout` 后原 token 立即失效。

### 1.4 通用错误响应

失败时统一返回：

```json
{
  "error": "具体错误信息"
}
```

常见状态码：

- `400 Bad Request`：请求体格式错误、参数错误、文件类型错误
- `401 Unauthorized`：缺少或使用了无效 token
- `404 Not Found`：资源不存在，或当前用户无权访问该资源
- `409 Conflict`：资源状态不满足，比如分析未完成就发布任务、未上报进度就请求提示
- `500 Internal Server Error`：服务端内部错误

### 1.5 时间和 ID

- 所有 `id` 都是数字，服务端内部类型是 `int64`
- 服务端生成的时间字段使用 RFC3339 字符串，比如 `2026-05-26T12:00:00Z`
- `reportedAt` 由客户端上传，服务端当前不会强校验格式，但建议始终使用 RFC3339

### 1.6 状态枚举

任务 `status`：

- `draft`
- `published`
- `archived`

任务分析 `analysisStatus`：

- `pending`
- `processing`
- `ready`
- `failed`

教师实时看板中的学生 `status`：

- `assigned`：已分配任务，但还没有进度上报
- `active`：已上报过进度

## 2. 学生客户端最小接入链路

学生客户端最小链路建议按这个顺序接：

1. 用 `POST /api/student/login` 登录，`clientType` 必须是 `desktop`
2. 调 `GET /api/student/assignments` 拉取已发布且已分配给当前学生的任务列表
3. 用户进入某个任务后，调 `GET /api/student/assignments/{id}` 拉详情
4. 本地 Scratch 状态变化后，调 `POST /api/student/assignments/{id}/progress` 上报快照
5. 上报成功后，再调 `POST /api/student/assignments/{id}/hints` 请求下一步提示
6. 客户端退出账号时，调 `POST /api/student/logout`

如果 `hints` 返回 `409`：

- `"assignment analysis is not ready"`：教师上传的参考 `sb3` 还没分析完成
- `"student progress is required before requesting a hint"`：还没先调 `progress`

## 3. 教师接口

### 3.1 注册

`POST /api/teacher/register`

请求：

```json
{
  "username": "teacher01",
  "password": "secret123"
}
```

成功响应 `201 Created`：

```json
{
  "token": "teacher-token",
  "teacherName": "teacher01"
}
```

失败：

- `409`：`teacher username already exists`

### 3.2 登录

`POST /api/teacher/login`

请求：

```json
{
  "username": "teacher01",
  "password": "secret123"
}
```

成功响应 `200 OK`：

```json
{
  "token": "teacher-token",
  "teacherName": "teacher01"
}
```

失败：

- `401`：`invalid teacher credentials`

### 3.3 当前教师信息

`GET /api/teacher/me`

成功响应：

```json
{
  "teacherId": 1,
  "teacherName": "teacher01"
}
```

### 3.4 退出登录

`POST /api/teacher/logout`

成功响应：

```json
{
  "status": "ok"
}
```

### 3.5 学生列表

`GET /api/teacher/students`

成功响应：

```json
{
  "items": [
    {
      "id": 1,
      "username": "student01",
      "displayName": "小明",
      "status": "active",
      "createdAt": "2026-05-26T12:00:00Z"
    }
  ]
}
```

### 3.6 单个创建学生

`POST /api/teacher/students`

请求：

```json
{
  "username": "student01",
  "displayName": "小明",
  "initialPassword": "abc12345"
}
```

成功响应 `201 Created`：

```json
{
  "created": [
    {
      "id": 1,
      "username": "student01",
      "displayName": "小明",
      "status": "active",
      "createdAt": "2026-05-26T12:00:00Z"
    }
  ],
  "conflicts": []
}
```

注意：这个接口虽然是单个创建，但返回结构和批量创建完全一致。

### 3.7 批量创建学生

`POST /api/teacher/students/batch`

请求：

```json
{
  "students": [
    {
      "username": "student01",
      "displayName": "小明",
      "initialPassword": "abc12345"
    },
    {
      "username": "student02",
      "displayName": "小红",
      "initialPassword": "xyz98765"
    }
  ]
}
```

成功响应 `201 Created`：

```json
{
  "created": [
    {
      "id": 1,
      "username": "student01",
      "displayName": "小明",
      "status": "active",
      "createdAt": "2026-05-26T12:00:00Z"
    }
  ],
  "conflicts": ["student02"]
}
```

`conflicts` 只返回冲突的用户名。

### 3.8 重置学生密码

`POST /api/teacher/students/{id}/reset-password`

请求：

```json
{
  "newPassword": "new-pass-2"
}
```

成功响应：

```json
{
  "id": 1,
  "username": "student01",
  "displayName": "小明",
  "status": "active",
  "createdAt": "2026-05-26T12:00:00Z"
}
```

失败：

- `404`：`student not found`

### 3.9 教师任务列表

`GET /api/teacher/assignments`

成功响应：

```json
{
  "items": [
    {
      "id": 1,
      "title": "Maze Game",
      "goal": "让角色按事件响应",
      "description": "第二阶段任务",
      "status": "draft",
      "analysisStatus": "ready",
      "studentCount": 2,
      "updatedAt": "2026-05-26T12:00:00Z"
    }
  ]
}
```

### 3.10 上传任务和参考 `sb3`

`POST /api/teacher/assignments`

请求格式：`multipart/form-data`

表单字段：

- `title`：字符串
- `goal`：字符串
- `description`：字符串
- `sb3`：文件字段，必填

约束：

- 文件扩展名必须是 `.sb3`
- 文件大小不能超过 `16 MiB`
- 允许的 MIME：`application/zip`、`application/x-zip-compressed`、`application/octet-stream`

成功响应 `201 Created`：

```json
{
  "id": 1,
  "title": "Maze Game",
  "goal": "让角色按事件响应",
  "description": "第二阶段任务",
  "status": "draft",
  "analysisStatus": "pending"
}
```

说明：

- 上传完成后只代表文件已保存成功
- `analysisStatus` 会在后台异步从 `pending -> processing -> ready / failed`
- 教师端需要轮询 `GET /api/teacher/assignments/{id}/analysis`

### 3.11 教师任务详情

`GET /api/teacher/assignments/{id}`

成功响应：

```json
{
  "id": 1,
  "title": "Maze Game",
  "goal": "让角色按事件响应",
  "description": "第二阶段任务",
  "status": "published",
  "analysisStatus": "ready",
  "roleNames": ["Stage", "Cat"],
  "scriptCounts": {
    "Stage": 1,
    "Cat": 2
  },
  "blockCounts": {
    "event_whenflagclicked": 1,
    "motion_movesteps": 1
  },
  "categoryCounts": {
    "event": 1,
    "motion": 1
  },
  "broadcastMessages": ["开始"],
  "variableNames": ["score"],
  "listNames": ["targets"],
  "extensions": ["pen"],
  "teachingPoints": ["先搭好事件入口", "再补动作流程"],
  "assignedStudents": [
    {
      "id": 1,
      "username": "student01",
      "displayName": "小明",
      "status": "active"
    }
  ],
  "updatedAt": "2026-05-26T12:00:00Z"
}
```

### 3.12 教师任务分析详情

`GET /api/teacher/assignments/{id}/analysis`

成功响应：

```json
{
  "assignmentId": 1,
  "analysisStatus": "ready",
  "analysisErrorMessage": "",
  "roleNames": ["Stage", "Cat"],
  "scriptCounts": {
    "Stage": 1,
    "Cat": 2
  },
  "blockCounts": {
    "event_whenflagclicked": 1,
    "motion_movesteps": 1
  },
  "categoryCounts": {
    "event": 1,
    "motion": 1
  },
  "broadcastMessages": ["开始"],
  "variableNames": ["score"],
  "listNames": ["targets"],
  "extensions": ["pen"],
  "teachingPoints": ["先搭好事件入口", "再补动作流程"]
}
```

如果分析失败：

```json
{
  "assignmentId": 1,
  "analysisStatus": "failed",
  "analysisErrorMessage": "zip: not a valid zip file",
  "roleNames": [],
  "scriptCounts": {},
  "blockCounts": {},
  "categoryCounts": {},
  "broadcastMessages": [],
  "variableNames": [],
  "listNames": [],
  "extensions": [],
  "teachingPoints": []
}
```

### 3.13 给任务分配学生

`POST /api/teacher/assignments/{id}/assign-students`

请求：

```json
{
  "studentIds": [1, 2]
}
```

成功响应：

```json
{
  "assignmentId": 1,
  "studentIds": [1, 2],
  "assignedCount": 2
}
```

失败：

- `404`：`assignment not found`
- `404`：`student not found`

### 3.14 发布任务

`POST /api/teacher/assignments/{id}/publish`

请求体：无

成功响应：

```json
{
  "id": 1,
  "title": "Maze Game",
  "status": "published",
  "analysisStatus": "ready"
}
```

失败：

- `404`：`assignment not found`
- `409`：`assignment analysis not ready`

### 3.15 归档任务

`POST /api/teacher/assignments/{id}/archive`

请求体：无

成功响应：

```json
{
  "id": 1,
  "title": "Maze Game",
  "status": "archived"
}
```

### 3.16 实时看板

`GET /api/teacher/dashboard/assignments/{id}/live`

成功响应：

```json
{
  "assignmentId": 1,
  "assignmentTitle": "Maze Game",
  "updatedAt": "2026-05-26T12:00:00Z",
  "students": [
    {
      "studentId": 1,
      "studentName": "小明",
      "status": "active",
      "currentTarget": "让 Cat 角色移动起来",
      "stepSummary": "已经把事件积木接上了",
      "currentRoleName": "Cat",
      "lastReportedAt": "2026-05-26T12:00:00Z",
      "lastHintText": "继续完善 Cat 角色。",
      "lastHintAt": "2026-05-26T12:00:05Z"
    }
  ]
}
```

说明：当前实时能力是轮询，不是 WebSocket / SSE 推送。

### 3.17 学生历史

`GET /api/teacher/dashboard/students/{id}/history`

成功响应：

```json
{
  "studentId": 1,
  "studentName": "小明",
  "items": [
    {
      "assignmentId": 1,
      "assignmentTitle": "Maze Game",
      "assignmentStatus": "published",
      "currentTarget": "让 Cat 角色移动起来",
      "stepSummary": "已经把事件积木接上了",
      "currentRoleName": "Cat",
      "reportedAt": "2026-05-26T12:00:00Z",
      "hintText": "继续完善 Cat 角色。",
      "hintProvider": "fallback",
      "hintCreatedAt": "2026-05-26T12:00:05Z"
    }
  ]
}
```

## 4. 学生接口

### 4.1 登录

`POST /api/student/login`

请求：

```json
{
  "username": "student01",
  "password": "abc12345",
  "clientType": "desktop"
}
```

成功响应：

```json
{
  "token": "student-token",
  "studentName": "小明"
}
```

失败：

- `400`：`student login only supports desktop client`
- `401`：`invalid student credentials`

`clientType` 当前必须严格等于 `desktop`。

### 4.2 当前学生信息

`GET /api/student/me`

成功响应：

```json
{
  "studentId": 1,
  "studentName": "小明",
  "username": "student01"
}
```

### 4.3 退出登录

`POST /api/student/logout`

成功响应：

```json
{
  "status": "ok"
}
```

### 4.4 任务列表

`GET /api/student/assignments`

返回当前学生“已分配且已发布”的任务。

成功响应：

```json
{
  "items": [
    {
      "id": 1,
      "title": "Maze Game",
      "goal": "让角色按事件响应",
      "description": "第二阶段任务",
      "status": "published",
      "analysisStatus": "ready",
      "roleNames": ["Stage", "Cat"],
      "scriptCounts": {
        "Stage": 1,
        "Cat": 2
      },
      "blockCounts": {
        "event_whenflagclicked": 1,
        "motion_movesteps": 1
      },
      "categoryCounts": {
        "event": 1,
        "motion": 1
      },
      "broadcastMessages": ["开始"],
      "variableNames": ["score"],
      "listNames": ["targets"],
      "extensions": ["pen"],
      "teachingPoints": ["先搭好事件入口", "再补动作流程"]
    }
  ]
}
```

### 4.5 任务详情

`GET /api/student/assignments/{id}`

成功响应：

```json
{
  "id": 1,
  "title": "Maze Game",
  "goal": "让角色按事件响应",
  "description": "第二阶段任务",
  "status": "published",
  "analysisStatus": "ready",
  "roleNames": ["Stage", "Cat"],
  "scriptCounts": {
    "Stage": 1,
    "Cat": 2
  },
  "blockCounts": {
    "event_whenflagclicked": 1,
    "motion_movesteps": 1
  },
  "categoryCounts": {
    "event": 1,
    "motion": 1
  },
  "broadcastMessages": ["开始"],
  "variableNames": ["score"],
  "listNames": ["targets"],
  "extensions": ["pen"],
  "teachingPoints": ["先搭好事件入口", "再补动作流程"],
  "latestProgress": {
    "currentTarget": "让 Cat 角色移动起来",
    "stepSummary": "已经把事件积木接上了",
    "currentRoleName": "Cat",
    "reportedAt": "2026-05-26T12:00:00Z"
  },
  "latestHint": {
    "hintText": "继续完善 Cat 角色。",
    "providerName": "fallback",
    "createdAt": "2026-05-26T12:00:05Z"
  }
}
```

说明：

- `latestProgress` 和 `latestHint` 可能不存在
- 当前学生拿不到未分配任务，也拿不到未发布任务

失败：

- `404`：`assignment is not available to the student`

### 4.6 上报进度

`POST /api/student/assignments/{id}/progress`

请求：

```json
{
  "currentTarget": "让 Cat 角色移动起来",
  "stepSummary": "已经把事件积木接上了",
  "localProjectHash": "hash-1",
  "reportedAt": "2026-05-26T12:00:00Z",
  "snapshot": {
    "currentRoleName": "Cat",
    "roles": [
      {
        "roleName": "Stage",
        "roleType": "stage",
        "blocks": ["当绿旗被点击"]
      },
      {
        "roleName": "Cat",
        "roleType": "sprite",
        "blocks": ["当接收到 开始", "移动 10 步"]
      }
    ]
  }
}
```

成功响应 `201 Created`：

```json
{
  "id": 1,
  "assignmentId": 1,
  "currentTarget": "让 Cat 角色移动起来",
  "stepSummary": "已经把事件积木接上了",
  "reportedAt": "2026-05-26T12:00:00Z"
}
```

当前 `snapshot` 没有强 schema 校验，但为了让服务端提示更稳定，第一版建议固定传这些字段：

- `currentRoleName`：当前正在编辑的角色名
- `roles[]`：当前项目中所有角色
- `roles[].roleName`
- `roles[].roleType`：建议使用 `stage` / `sprite`
- `roles[].blocks[]`：当前角色已有的积木文本列表

失败：

- `404`：`assignment is not available to the student`

### 4.7 请求提示

`POST /api/student/assignments/{id}/hints`

请求体：无

成功响应 `201 Created`：

```json
{
  "id": 1,
  "assignmentId": 1,
  "hintText": "继续完善 Cat 角色。你已经有 当接收到 开始、移动 10 步，下一步把这些积木串成完整流程，并对照任务目标：让角色按事件响应。",
  "providerName": "fallback"
}
```

`providerName` 可能是：

- `fallback`
- 真实 `DeepSeek` provider 名称

失败：

- `404`：`assignment is not available to the student`
- `409`：`assignment analysis is not ready`
- `409`：`student progress is required before requesting a hint`

建议客户端总是在最近一次 `progress` 成功后，再调用 `hints`。

## 5. 客户端开发时最需要注意的点

### 5.1 学生端只认 `desktop`

学生登录的 `clientType` 不是建议值，是硬条件。传 `web`、`ios`、`android` 都会直接 `400`。

### 5.2 学生任务列表只返回已发布任务

教师已经分配但还没发布的任务，不会出现在 `GET /api/student/assignments`。

### 5.3 提示链路依赖最近一次进度

当前 `POST /api/student/assignments/{id}/hints` 不接收请求体，它完全依赖服务端保存的“最近一次 progress + 参考任务分析结果”。

### 5.4 `snapshot` 会直接影响提示质量

虽然服务端当前允许更松的 JSON，但 fallback 和后续提示编排已经会读取：

- `snapshot.currentRoleName`
- `snapshot.roles[].roleName`
- `snapshot.roles[].blocks[]`

这几个字段缺失时，提示会明显变泛。

### 5.5 教师实时能力当前是轮询

如果你后面也要做教师端客户端，不要按 WebSocket/SSE 设计。当前真实接口是轮询 `GET /api/teacher/dashboard/assignments/{id}/live`。

## 6. 推荐联调顺序

如果只开发学生客户端，建议这样联调：

1. 教师先注册、登录
2. 教师创建学生
3. 教师上传 `sb3`
4. 教师轮询分析结果直到 `analysisStatus=ready`
5. 教师分配学生并发布任务
6. 学生登录
7. 学生拉任务列表
8. 学生拉任务详情
9. 学生上报进度
10. 学生请求提示

如果你要先做脚手架，最先打通这 4 个接口就够了：

- `POST /api/student/login`
- `GET /api/student/assignments`
- `POST /api/student/assignments/{id}/progress`
- `POST /api/student/assignments/{id}/hints`
