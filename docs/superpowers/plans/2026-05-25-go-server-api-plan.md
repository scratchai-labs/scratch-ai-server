# Go Server API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Go version of `apps/server-api` for the classroom workflow: teacher auth, batch student creation, asynchronous `sb3` analysis, student progress reporting, server-side DeepSeek hints, and live teacher dashboard data.

**Architecture:** Replace the Python prototype in `apps/server-api` with a Go HTTP service built around small internal packages. Start with an in-process app boundary, HTTP handler layer, in-memory/test-friendly repositories, and asynchronous analysis worker hooks, then keep root scripts and the teacher web aligned with the new runtime.

**Tech Stack:** Go, `net/http`, JSON, ZIP parsing, Vitest, npm workspace scripts

---

### Task 1: Go API Skeleton

**Files:**
- Create: `apps/server-api/go.mod`
- Create: `apps/server-api/cmd/server-api/main.go`
- Create: `apps/server-api/internal/config/config.go`
- Create: `apps/server-api/internal/http/router.go`
- Create: `apps/server-api/internal/http/health_handler.go`
- Create: `apps/server-api/internal/app/app.go`
- Create: `apps/server-api/tests/health_test.go`
- Modify: `package.json`
- Modify: `scripts/dev-server-stack.mjs`

- [ ] **Step 1: Write the failing health test**

```go
func TestHealthCheckReturnsOK(t *testing.T) {
	app := testapp.New(t)
	res := app.GET(t, "/health", nil)

	require.Equal(t, http.StatusOK, res.Code)
	require.JSONEq(t, `{"status":"ok"}`, res.Body.String())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/server-api && go test ./tests -run TestHealthCheckReturnsOK`
Expected: FAIL because `go.mod` and the test helper app do not exist yet

- [ ] **Step 3: Write minimal implementation**

```go
router := http.NewServeMux()
router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
})
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/server-api && go test ./tests -run TestHealthCheckReturnsOK`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/server-api package.json scripts/dev-server-stack.mjs
git commit -m "feat: 搭建 go 服务端骨架"
```

### Task 2: Teacher Auth And Student Batch Creation

**Files:**
- Create: `apps/server-api/internal/auth/service.go`
- Create: `apps/server-api/internal/auth/handler.go`
- Create: `apps/server-api/internal/student/service.go`
- Create: `apps/server-api/internal/student/handler.go`
- Create: `apps/server-api/internal/store/memory/store.go`
- Create: `apps/server-api/internal/store/memory/teacher_store.go`
- Create: `apps/server-api/internal/store/memory/student_store.go`
- Create: `apps/server-api/internal/security/password.go`
- Create: `apps/server-api/internal/security/token.go`
- Create: `apps/server-api/tests/teacher_auth_test.go`
- Create: `apps/server-api/tests/student_batch_test.go`

- [ ] **Step 1: Write the failing auth and batch tests**

```go
func TestTeacherCanRegisterAndLogin(t *testing.T) { /* register -> login -> token */ }
func TestTeacherCanBatchCreateStudents(t *testing.T) { /* partial success + conflict */ }
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd apps/server-api && go test ./tests -run 'TestTeacherCanRegisterAndLogin|TestTeacherCanBatchCreateStudents'`
Expected: FAIL because teacher/student handlers and repositories do not exist

- [ ] **Step 3: Write minimal implementation**

```go
type Teacher struct { ID int64; Username string; PasswordHash string }
type StudentBatchResult struct { Created []Student `json:"created"`; Conflicts []string `json:"conflicts"` }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd apps/server-api && go test ./tests -run 'TestTeacherCanRegisterAndLogin|TestTeacherCanBatchCreateStudents'`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/server-api
git commit -m "feat: 实现教师鉴权和批量建学生"
```

### Task 3: Assignment Upload And Async Analysis

**Files:**
- Create: `apps/server-api/internal/assignment/service.go`
- Create: `apps/server-api/internal/assignment/handler.go`
- Create: `apps/server-api/internal/sb3/analyzer.go`
- Create: `apps/server-api/internal/sb3/worker.go`
- Create: `apps/server-api/internal/store/memory/assignment_store.go`
- Create: `apps/server-api/tests/assignment_upload_test.go`
- Create: `apps/server-api/tests/assignment_analysis_test.go`

- [ ] **Step 1: Write the failing upload and async analysis tests**

```go
func TestTeacherCanUploadAssignmentSB3(t *testing.T) { /* multipart upload returns pending */ }
func TestAssignmentAnalysisTransitionsToReady(t *testing.T) { /* pending -> processing -> ready */ }
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd apps/server-api && go test ./tests -run 'TestTeacherCanUploadAssignmentSB3|TestAssignmentAnalysisTransitionsToReady'`
Expected: FAIL because assignment routes, stores, and worker do not exist

- [ ] **Step 3: Write minimal implementation**

```go
type AnalysisStatus string
const (
	AnalysisPending AnalysisStatus = "pending"
	AnalysisProcessing AnalysisStatus = "processing"
	AnalysisReady AnalysisStatus = "ready"
	AnalysisFailed AnalysisStatus = "failed"
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd apps/server-api && go test ./tests -run 'TestTeacherCanUploadAssignmentSB3|TestAssignmentAnalysisTransitionsToReady'`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/server-api
git commit -m "feat: 实现任务上传与异步 sb3 分析"
```

### Task 4: Student Progress, Hints, And Live Dashboard

**Files:**
- Create: `apps/server-api/internal/progress/service.go`
- Create: `apps/server-api/internal/progress/handler.go`
- Create: `apps/server-api/internal/hint/service.go`
- Create: `apps/server-api/internal/hint/handler.go`
- Create: `apps/server-api/internal/dashboard/service.go`
- Create: `apps/server-api/internal/dashboard/handler.go`
- Create: `apps/server-api/tests/progress_test.go`
- Create: `apps/server-api/tests/hint_test.go`
- Create: `apps/server-api/tests/dashboard_test.go`

- [ ] **Step 1: Write the failing progress, hint, and dashboard tests**

```go
func TestStudentCanReportProgress(t *testing.T) { /* currentRoleName + roles + blocks */ }
func TestStudentCannotRequestHintUntilAnalysisReady(t *testing.T) { /* 409 conflict */ }
func TestTeacherCanReadLiveDashboard(t *testing.T) { /* latest progress + latest hint */ }
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd apps/server-api && go test ./tests -run 'TestStudentCanReportProgress|TestStudentCannotRequestHintUntilAnalysisReady|TestTeacherCanReadLiveDashboard'`
Expected: FAIL because progress/hint/dashboard services do not exist

- [ ] **Step 3: Write minimal implementation**

```go
type SnapshotRole struct {
	RoleName string   `json:"roleName"`
	RoleType string   `json:"roleType"`
	Blocks   []string `json:"blocks"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd apps/server-api && go test ./tests -run 'TestStudentCanReportProgress|TestStudentCannotRequestHintUntilAnalysisReady|TestTeacherCanReadLiveDashboard'`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/server-api
git commit -m "feat: 实现学生进度、提示和实时看板"
```

### Task 5: Runtime Alignment And Verification

**Files:**
- Modify: `apps/server-api/README.md`
- Modify: `apps/server-web/src/services/teacherApi.ts`
- Modify: `apps/server-web/src/services/teacherApi.test.ts`
- Modify: `apps/server-web/src/stores/teacherDirectory.ts`
- Modify: `apps/server-web/src/stores/liveDashboard.ts`
- Modify: `docs/server-development.zh-CN.md`

- [ ] **Step 1: Write failing client contract tests if endpoint shapes change**

```ts
it('reads teacher resources from the go api paths', async () => {
  // assert updated fetch paths and response shapes
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npm run server:web:test`
Expected: FAIL if API contract changes are not reflected in teacher web client

- [ ] **Step 3: Write minimal implementation**

```ts
buildApiUrl(baseUrl, '/api/teacher/assignments')
```

- [ ] **Step 4: Run full verification**

Run: `cd apps/server-api && go test ./...`
Expected: PASS

Run: `cd ../../ && npm run server:web:test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/server-api apps/server-web package.json docs
git commit -m "feat: 接通 go 服务端与教师后台"
```
