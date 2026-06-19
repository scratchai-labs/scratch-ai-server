import { inject, type InjectionKey } from 'vue'
import { buildApiUrl, HttpError, requestJson, type FetchLike } from './http'

export interface TeacherLoginInput {
  username: string
  password: string
}

export interface TeacherSession {
  token: string
  teacherName: string
  role: 'teacher' | 'admin'
}

export type ManagedTeacherRole = 'teacher' | 'admin'

export interface ManagedTeacher {
  id: string
  username: string
  role: string
  status: string
  createdAt: string
}

export interface CreateManagedTeacherInput {
  username: string
  initialPassword: string
}

export interface CreateManagedStudentInput {
  teacherId: string
  username: string
  displayName: string
  initialPassword: string
}

export interface CreateTeacherStudentInput {
  username: string
  displayName: string
  initialPassword: string
}

export interface UpdateTeacherStudentInput {
  username: string
  displayName: string
}

export interface TeacherClassroom {
  id: string
  name: string
  studentCount: number
  projectCount: number
  createdAt: string
  updatedAt: string
}

export interface TeacherClassroomDetail extends TeacherClassroom {}

export interface CreateTeacherClassroomInput {
  name: string
}

export interface BatchCreateTeacherStudentsResult {
  created: TeacherStudent[]
  conflicts: string[]
}

export interface AdminOverview {
  adminCount: number
  teacherCount: number
  activeTeacherCount: number
  disabledTeacherCount: number
  studentCount: number
  activeStudentCount: number
  disabledStudentCount: number
}

export interface AdminAuditLog {
  id: string
  actorUsername: string
  action: string
  targetType: string
  targetId: string
  targetUsername: string
  before: Record<string, string>
  after: Record<string, string>
  createdAt: string
}

export interface ManagedStudent {
  id: string
  teacherId: string
  teacherUsername: string
  username: string
  displayName: string
  status: string
  createdAt: string
}

export interface TeacherStudent {
  id: string
  classroomId?: string
  username: string
  name: string
  className: string
  progress: number
  status?: string
  currentTarget?: string
  stepSummary?: string
  latestAiHint: string
  updatedAt: string
  createdAt: string
}

export type TeacherReleaseStatus = 'draft' | 'published' | 'archived'

export interface TeacherRelease {
  id: string
  classroomId?: string
  title: string
  goal: string
  description: string
  className: string
  status: TeacherReleaseStatus
  analysisStatus: string
  studentCount: number
  updatedAt: string
}

export interface CreateTeacherReleaseInput {
  title: string
  goal: string
  description: string
  file: File
}

export interface TeacherReleaseAssignedStudent {
  id: string
  username: string
  displayName: string
  status: string
}

interface TeacherReleaseAnalysisSummary {
  roleNames: string[]
  scriptCounts: Record<string, number>
  blockCounts: Record<string, number>
  categoryCounts: Record<string, number>
  broadcastMessages: string[]
  variableNames: string[]
  listNames: string[]
  extensions: string[]
  teachingPoints: string[]
}

export interface TeacherReleaseDetail extends TeacherReleaseAnalysisSummary {
  id: string
  title: string
  goal: string
  description: string
  status: TeacherReleaseStatus
  analysisStatus: string
  assignedStudents: TeacherReleaseAssignedStudent[]
  updatedAt: string
}

export interface TeacherReleaseAnalysis extends TeacherReleaseAnalysisSummary {
  assignmentId: string
  analysisStatus: string
  analysisErrorMessage: string
}

export interface TeacherReleaseAssignmentResult {
  assignmentId: string
  studentIds: string[]
  assignedCount: number
}

export interface TeacherReleaseMutationResult {
  id: string
  title: string
  status: TeacherReleaseStatus
  analysisStatus: string
}

export interface LiveStudentSnapshot {
  id: string
  name: string
  progress: number
  status?: string
  currentTarget?: string
  stepSummary?: string
  latestAiHint: string
  updatedAt: string
}

export interface LiveDashboardSnapshot {
  releaseId: string
  releaseTitle: string
  updatedAt: string
  students: LiveStudentSnapshot[]
}

export interface TeacherApiClient {
  login(input: TeacherLoginInput): Promise<TeacherSession>
  logout?(): Promise<void>
  listClassrooms?(): Promise<TeacherClassroom[]>
  createClassroom?(input: CreateTeacherClassroomInput): Promise<TeacherClassroom>
  getClassroomDetail?(classroomId: string): Promise<TeacherClassroomDetail>
  updateClassroom?(classroomId: string, input: CreateTeacherClassroomInput): Promise<TeacherClassroom>
  deleteClassroom?(classroomId: string): Promise<void>
  listStudents(): Promise<TeacherStudent[]>
  listClassroomStudents?(classroomId: string): Promise<TeacherStudent[]>
  createStudent?(input: CreateTeacherStudentInput): Promise<TeacherStudent>
  createClassroomStudent?(classroomId: string, input: CreateTeacherStudentInput): Promise<TeacherStudent>
  batchCreateStudents?(
    input: CreateTeacherStudentInput[],
  ): Promise<BatchCreateTeacherStudentsResult>
  batchCreateClassroomStudents?(
    classroomId: string,
    input: CreateTeacherStudentInput[],
  ): Promise<BatchCreateTeacherStudentsResult>
  resetStudentPassword?(studentId: string, newPassword: string): Promise<TeacherStudent>
  resetClassroomStudentPassword?(classroomId: string, studentId: string, newPassword: string): Promise<TeacherStudent>
  updateClassroomStudent?(classroomId: string, studentId: string, input: UpdateTeacherStudentInput): Promise<TeacherStudent>
  deleteClassroomStudent?(classroomId: string, studentId: string): Promise<void>
  listReleases(): Promise<TeacherRelease[]>
  listClassroomProjects?(classroomId: string): Promise<TeacherRelease[]>
  createRelease?(input: CreateTeacherReleaseInput): Promise<TeacherReleaseMutationResult>
  createClassroomProject?(classroomId: string, input: CreateTeacherReleaseInput): Promise<TeacherReleaseMutationResult>
  getReleaseDetail?(releaseId: string): Promise<TeacherReleaseDetail>
  getReleaseAnalysis?(releaseId: string): Promise<TeacherReleaseAnalysis>
  assignStudentsToRelease?(
    releaseId: string,
    studentIds: string[],
  ): Promise<TeacherReleaseAssignmentResult>
  publishRelease?(releaseId: string): Promise<TeacherReleaseMutationResult>
  archiveRelease?(releaseId: string): Promise<TeacherReleaseMutationResult>
  getLiveDashboard(releaseId: string): Promise<LiveDashboardSnapshot>
  getAdminOverview?(): Promise<AdminOverview>
  listAdminAuditLogs?(): Promise<AdminAuditLog[]>
  listTeachers?(): Promise<ManagedTeacher[]>
  createTeacher?(input: CreateManagedTeacherInput): Promise<ManagedTeacher>
  resetTeacherPassword?(teacherId: string, newPassword: string): Promise<ManagedTeacher>
  changeTeacherRole?(teacherId: string, role: ManagedTeacherRole): Promise<ManagedTeacher>
  enableTeacher?(teacherId: string): Promise<ManagedTeacher>
  disableTeacher?(teacherId: string): Promise<ManagedTeacher>
  listManagedStudents?(): Promise<ManagedStudent[]>
  createManagedStudent?(input: CreateManagedStudentInput): Promise<ManagedStudent>
  resetManagedStudentPassword?(studentId: string, newPassword: string): Promise<ManagedStudent>
  enableManagedStudent?(studentId: string): Promise<ManagedStudent>
  disableManagedStudent?(studentId: string): Promise<ManagedStudent>
}

interface TeacherStudentHistoryItem {
  assignmentStatus?: string
  currentTarget?: string
  stepSummary?: string
  reportedAt?: string
  hintText?: string
  hintCreatedAt?: string
}

export const teacherApiKey: InjectionKey<TeacherApiClient> = Symbol(
  'teacher-api-client',
)

export class TeacherApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'TeacherApiError'
    this.status = status
  }
}

export function useTeacherApiClient(): TeacherApiClient {
  const client = inject(teacherApiKey)

  if (!client) {
    throw new Error('Teacher API client is not provided.')
  }

  return client
}

export function createFetchTeacherApiClient(options: {
  baseUrl?: string
  fetchImpl?: FetchLike
  getToken?: () => string
  onUnauthorized?: () => void | Promise<void>
} = {}): TeacherApiClient {
  const fetchImpl = options.fetchImpl ?? fetch
  const baseUrl = options.baseUrl
  const getToken = options.getToken
  const onUnauthorized = options.onUnauthorized

  async function requestAuthedJson<T>(path: string): Promise<T> {
    return requestJson<T>(
      fetchImpl,
      buildApiUrl(baseUrl, path),
      {
        method: 'GET',
        headers: buildAuthHeaders(getToken),
      },
      {
        onUnauthorized,
      },
    )
  }

  async function requestAuthedMutation<T>(
    path: string,
    body?: Record<string, unknown>,
  ): Promise<T> {
    return requestJson<T>(
      fetchImpl,
      buildApiUrl(baseUrl, path),
      {
        method: 'POST',
        headers: {
          ...(buildAuthHeaders(getToken) ?? {}),
          ...(body ? { 'Content-Type': 'application/json' } : {}),
        },
        body: body ? JSON.stringify(body) : undefined,
      },
      {
        onUnauthorized,
      },
    )
  }

  return {
    async login(input) {
      const payload = await requestJson<TeacherSession>(
        fetchImpl,
        buildApiUrl(baseUrl, '/api/teacher/login'),
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(input),
        },
      )
      return normalizeTeacherSession(payload)
    },
    async logout() {
      await requestJson(
        fetchImpl,
        buildApiUrl(baseUrl, '/api/teacher/logout'),
        {
          method: 'POST',
          headers: buildAuthHeaders(getToken),
        },
        {
          onUnauthorized,
        },
      )
    },
    async listClassrooms() {
      const payload = await requestAuthedJson<unknown>('/api/teacher/classes')
      return normalizeClassrooms(payload)
    },
    async createClassroom(input) {
      const payload = await requestAuthedMutation<unknown>('/api/teacher/classes', input)
      return normalizeTeacherClassroom(payload)
    },
    async getClassroomDetail(classroomId) {
      const payload = await requestAuthedJson<unknown>(`/api/teacher/classes/${classroomId}`)
      return normalizeTeacherClassroomDetail(payload)
    },
    async updateClassroom(classroomId, input) {
      const payload = await requestJson<unknown>(
        fetchImpl,
        buildApiUrl(baseUrl, `/api/teacher/classes/${classroomId}`),
        {
          method: 'PATCH',
          headers: {
            ...(buildAuthHeaders(getToken) ?? {}),
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(input),
        },
        {
          onUnauthorized,
        },
      )
      return normalizeTeacherClassroom(payload)
    },
    async deleteClassroom(classroomId) {
      await requestJson(
        fetchImpl,
        buildApiUrl(baseUrl, `/api/teacher/classes/${classroomId}`),
        {
          method: 'DELETE',
          headers: buildAuthHeaders(getToken),
        },
        {
          onUnauthorized,
        },
      )
    },
    async listStudents() {
      const payload = await requestAuthedJson<unknown>('/api/teacher/students')
      const students = normalizeStudents(payload)
      if (!students.length) {
        return students
      }

      const histories = await Promise.all(
        students.map(async (student) => {
          try {
            const historyPayload = await requestAuthedJson<unknown>(
              `/api/teacher/dashboard/students/${student.id}/history`,
            )
            return normalizeStudentHistoryItems(historyPayload)
          } catch (error) {
            if (error instanceof HttpError && error.status === 401) {
              throw error
            }
            return []
          }
        }),
      )

      return students
        .map((student, index) => applyStudentHistory(student, histories[index] ?? []))
        .sort(compareStudentsByUpdatedAt)
    },
    async listClassroomStudents(classroomId) {
      const payload = await requestAuthedJson<unknown>(`/api/teacher/classes/${classroomId}/students`)
      return normalizeStudents(payload)
    },
    async createStudent(input) {
      const payload = await requestAuthedMutation<unknown>('/api/teacher/students', input)
      return normalizeCreatedTeacherStudent(payload)
    },
    async createClassroomStudent(classroomId, input) {
      const payload = await requestAuthedMutation<unknown>(`/api/teacher/classes/${classroomId}/students`, input)
      return normalizeCreatedTeacherStudent(payload)
    },
    async batchCreateStudents(input) {
      const payload = await requestAuthedMutation<unknown>('/api/teacher/students/batch', {
        students: input,
      })
      return normalizeTeacherStudentBatchResult(payload)
    },
    async batchCreateClassroomStudents(classroomId, input) {
      const payload = await requestAuthedMutation<unknown>(`/api/teacher/classes/${classroomId}/students/batch`, {
        students: input,
      })
      return normalizeTeacherStudentBatchResult(payload)
    },
    async resetStudentPassword(studentId, newPassword) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/teacher/students/${studentId}/reset-password`,
        { newPassword },
      )
      return normalizeTeacherStudentRecord(payload)
    },
    async resetClassroomStudentPassword(classroomId, studentId, newPassword) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/teacher/classes/${classroomId}/students/${studentId}/reset-password`,
        { newPassword },
      )
      return normalizeTeacherStudentRecord(payload)
    },
    async updateClassroomStudent(classroomId, studentId, input) {
      const payload = await requestJson<unknown>(
        fetchImpl,
        buildApiUrl(baseUrl, `/api/teacher/classes/${classroomId}/students/${studentId}`),
        {
          method: 'PATCH',
          headers: {
            ...(buildAuthHeaders(getToken) ?? {}),
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(input),
        },
        {
          onUnauthorized,
        },
      )
      return normalizeTeacherStudentRecord(payload)
    },
    async deleteClassroomStudent(classroomId, studentId) {
      await requestJson(
        fetchImpl,
        buildApiUrl(baseUrl, `/api/teacher/classes/${classroomId}/students/${studentId}`),
        {
          method: 'DELETE',
          headers: buildAuthHeaders(getToken),
        },
        {
          onUnauthorized,
        },
      )
    },
    async listReleases() {
      const payload = await requestAuthedJson<unknown>('/api/teacher/assignments')
      return normalizeReleases(payload)
    },
    async listClassroomProjects(classroomId) {
      const payload = await requestAuthedJson<unknown>(`/api/teacher/classes/${classroomId}/projects`)
      return normalizeReleases(payload)
    },
    async createRelease(input) {
      const body = new FormData()
      body.append('title', input.title)
      body.append('goal', input.goal)
      body.append('description', input.description)
      body.append('sb3', input.file)
      const payload = await requestJson<unknown>(
        fetchImpl,
        buildApiUrl(baseUrl, '/api/teacher/assignments'),
        {
          method: 'POST',
          headers: buildAuthHeaders(getToken),
          body,
        },
        {
          onUnauthorized,
        },
      )
      return normalizeTeacherReleaseMutation(payload)
    },
    async createClassroomProject(classroomId, input) {
      const body = new FormData()
      body.append('title', input.title)
      body.append('goal', input.goal)
      body.append('description', input.description)
      body.append('sb3', input.file)
      const payload = await requestJson<unknown>(
        fetchImpl,
        buildApiUrl(baseUrl, `/api/teacher/classes/${classroomId}/projects`),
        {
          method: 'POST',
          headers: buildAuthHeaders(getToken),
          body,
        },
        {
          onUnauthorized,
        },
      )
      return normalizeTeacherReleaseMutation(payload)
    },
    async getReleaseDetail(releaseId) {
      const payload = await requestAuthedJson<unknown>(
        `/api/teacher/assignments/${releaseId}`,
      )
      return normalizeReleaseDetail(payload)
    },
    async getReleaseAnalysis(releaseId) {
      const payload = await requestAuthedJson<unknown>(
        `/api/teacher/assignments/${releaseId}/analysis`,
      )
      return normalizeReleaseAnalysis(payload)
    },
    async assignStudentsToRelease(releaseId, studentIds) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/teacher/assignments/${releaseId}/assign-students`,
        {
          studentIds: studentIds.map((studentId) => Number(studentId)),
        },
      )
      return normalizeTeacherReleaseAssignmentResult(payload)
    },
    async publishRelease(releaseId) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/teacher/assignments/${releaseId}/publish`,
      )
      return normalizeTeacherReleaseMutation(payload)
    },
    async archiveRelease(releaseId) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/teacher/assignments/${releaseId}/archive`,
      )
      return normalizeTeacherReleaseMutation(payload)
    },
    async getLiveDashboard(releaseId) {
      const payload = await requestAuthedJson<unknown>(
        `/api/teacher/dashboard/assignments/${releaseId}/live`,
      )
      return normalizeLiveDashboard(payload)
    },
    async getAdminOverview() {
      const payload = await requestAuthedJson<unknown>('/api/admin/overview')
      return normalizeAdminOverview(payload)
    },
    async listAdminAuditLogs() {
      const payload = await requestAuthedJson<unknown>('/api/admin/audit-logs')
      return normalizeAdminAuditLogs(payload)
    },
    async listTeachers() {
      const payload = await requestAuthedJson<unknown>('/api/admin/teachers')
      return normalizeManagedTeachers(payload)
    },
    async createTeacher(input) {
      const payload = await requestAuthedMutation<unknown>('/api/admin/teachers', input)
      return normalizeManagedTeacher(payload)
    },
    async resetTeacherPassword(teacherId, newPassword) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/teachers/${teacherId}/reset-password`,
        { newPassword },
      )
      return normalizeManagedTeacher(payload)
    },
    async changeTeacherRole(teacherId, role) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/teachers/${teacherId}/role`,
        { role },
      )
      return normalizeManagedTeacher(payload)
    },
    async enableTeacher(teacherId) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/teachers/${teacherId}/enable`,
      )
      return normalizeManagedTeacher(payload)
    },
    async disableTeacher(teacherId) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/teachers/${teacherId}/disable`,
      )
      return normalizeManagedTeacher(payload)
    },
    async listManagedStudents() {
      const payload = await requestAuthedJson<unknown>('/api/admin/students')
      return normalizeManagedStudents(payload)
    },
    async createManagedStudent(input) {
      const payload = await requestAuthedMutation<unknown>('/api/admin/students', {
        ...input,
        teacherId: Number(input.teacherId),
      })
      return normalizeManagedStudent(payload)
    },
    async resetManagedStudentPassword(studentId, newPassword) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/students/${studentId}/reset-password`,
        { newPassword },
      )
      return normalizeManagedStudent(payload)
    },
    async enableManagedStudent(studentId) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/students/${studentId}/enable`,
      )
      return normalizeManagedStudent(payload)
    },
    async disableManagedStudent(studentId) {
      const payload = await requestAuthedMutation<unknown>(
        `/api/admin/students/${studentId}/disable`,
      )
      return normalizeManagedStudent(payload)
    },
  }
}

function normalizeTeacherSession(payload: TeacherSession): TeacherSession {
  return {
    token: String(payload.token ?? ''),
    teacherName: String(payload.teacherName ?? ''),
    role: payload.role === 'admin' ? 'admin' : 'teacher',
  }
}

function normalizeCollection<T>(payload: unknown): T[] {
  if (Array.isArray(payload)) {
    return payload as T[]
  }

  if (payload && typeof payload === 'object') {
    const record = payload as Record<string, unknown>
    if (Array.isArray(record.items)) {
      return record.items as T[]
    }
  }

  return []
}

function buildAuthHeaders(getToken: (() => string) | undefined): HeadersInit | undefined {
  const token = getToken?.().trim()
  if (!token) {
    return undefined
  }

  return {
    Authorization: `Bearer ${token}`,
  }
}

function normalizeClassrooms(payload: unknown): TeacherClassroom[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) =>
    normalizeTeacherClassroom(item),
  )
}

function normalizeTeacherClassroom(payload: unknown): TeacherClassroom {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    id: String(record.id ?? ''),
    name: String(record.name ?? ''),
    studentCount: Number(record.studentCount ?? 0),
    projectCount: Number(record.projectCount ?? 0),
    createdAt: String(record.createdAt ?? '—'),
    updatedAt: String(record.updatedAt ?? '—'),
  }
}

function normalizeTeacherClassroomDetail(payload: unknown): TeacherClassroomDetail {
  return normalizeTeacherClassroom(payload)
}

function normalizeStudents(payload: unknown): TeacherStudent[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) =>
    normalizeTeacherStudentRecord(item),
  )
}

function normalizeCreatedTeacherStudent(payload: unknown): TeacherStudent {
  const result = normalizeTeacherStudentBatchResult(payload)
  if (result.created[0]) {
    return result.created[0]
  }

  if (result.conflicts.length) {
    throw new TeacherApiError(`学生账号冲突：${result.conflicts.join('、')}`, 409)
  }

  throw new Error('学生创建响应无结果')
}

function normalizeTeacherStudentBatchResult(payload: unknown): BatchCreateTeacherStudentsResult {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}
  const created = Array.isArray(record.created) ? record.created : []
  const conflicts = Array.isArray(record.conflicts) ? record.conflicts : []

  return {
    created: created.map((item) => normalizeTeacherStudentRecord(item)),
    conflicts: conflicts.map((conflict) => String(conflict)),
  }
}

function normalizeTeacherStudentRecord(payload: unknown): TeacherStudent {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}
  const createdAt = String(record.createdAt ?? record.updatedAt ?? '—')

  return {
    id: String(record.id ?? ''),
    classroomId: pickFirstNonEmpty(record.classroomId),
    username: String(record.username ?? ''),
    name: String(record.displayName ?? record.name ?? ''),
    className: pickFirstNonEmpty(record.classroomName, record.className) || '未分组',
    progress: 0,
    status: '',
    currentTarget: '',
    stepSummary: '',
    latestAiHint: '等待学生请求提示',
    updatedAt: createdAt,
    createdAt,
  }
}

function normalizeManagedTeachers(payload: unknown): ManagedTeacher[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) =>
    normalizeManagedTeacher(item),
  )
}

function normalizeAdminAuditLogs(payload: unknown): AdminAuditLog[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) =>
    normalizeAdminAuditLog(item),
  )
}

function normalizeAdminOverview(payload: unknown): AdminOverview {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    adminCount: Number(record.adminCount ?? 0),
    teacherCount: Number(record.teacherCount ?? 0),
    activeTeacherCount: Number(record.activeTeacherCount ?? 0),
    disabledTeacherCount: Number(record.disabledTeacherCount ?? 0),
    studentCount: Number(record.studentCount ?? 0),
    activeStudentCount: Number(record.activeStudentCount ?? 0),
    disabledStudentCount: Number(record.disabledStudentCount ?? 0),
  }
}

function normalizeManagedTeacher(payload: unknown): ManagedTeacher {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    id: String(record.id ?? ''),
    username: String(record.username ?? ''),
    role: String(record.role ?? 'teacher'),
    status: String(record.status ?? 'active'),
    createdAt: String(record.createdAt ?? '—'),
  }
}

function normalizeAdminAuditLog(payload: unknown): AdminAuditLog {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    id: String(record.id ?? ''),
    actorUsername: String(record.actorUsername ?? ''),
    action: String(record.action ?? ''),
    targetType: String(record.targetType ?? ''),
    targetId: String(record.targetId ?? ''),
    targetUsername: String(record.targetUsername ?? ''),
    before: normalizeStringMap(record.before),
    after: normalizeStringMap(record.after),
    createdAt: String(record.createdAt ?? '—'),
  }
}

function normalizeManagedStudents(payload: unknown): ManagedStudent[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) =>
    normalizeManagedStudent(item),
  )
}

function normalizeManagedStudent(payload: unknown): ManagedStudent {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    id: String(record.id ?? ''),
    teacherId: String(record.teacherId ?? ''),
    teacherUsername: String(record.teacherUsername ?? ''),
    username: String(record.username ?? ''),
    displayName: String(record.displayName ?? ''),
    status: String(record.status ?? 'active'),
    createdAt: String(record.createdAt ?? '—'),
  }
}

function normalizeStudentHistoryItems(payload: unknown): TeacherStudentHistoryItem[] {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}
  const items = Array.isArray(record.items) ? record.items : []

  return items
    .map((item) => normalizeStudentHistoryItem(item))
    .filter(Boolean) as TeacherStudentHistoryItem[]
}

function normalizeStudentHistoryItem(payload: unknown): TeacherStudentHistoryItem | null {
  if (!payload || typeof payload !== 'object') {
    return null
  }

  const record = payload as Record<string, unknown>
  return {
    assignmentStatus: pickFirstNonEmpty(record.assignmentStatus),
    currentTarget: pickFirstNonEmpty(record.currentTarget),
    stepSummary: pickFirstNonEmpty(record.stepSummary),
    reportedAt: pickFirstNonEmpty(record.reportedAt),
    hintText: pickFirstNonEmpty(record.hintText),
    hintCreatedAt: pickFirstNonEmpty(record.hintCreatedAt),
  }
}

function applyStudentHistory(student: TeacherStudent, historyItems: TeacherStudentHistoryItem[]): TeacherStudent {
  const latestItem = pickLatestStudentHistory(historyItems)
  if (!latestItem) {
    return student
  }

  const hasProgressUpdate = Boolean(
    latestItem.currentTarget || latestItem.stepSummary || latestItem.reportedAt,
  )

  return {
    ...student,
    status: hasProgressUpdate ? 'active' : latestItem.assignmentStatus === 'published' ? 'assigned' : '',
    currentTarget: latestItem.currentTarget || '',
    stepSummary: latestItem.stepSummary || '',
    latestAiHint: latestItem.hintText || student.latestAiHint,
    updatedAt:
      pickFirstNonEmpty(latestItem.hintCreatedAt, latestItem.reportedAt, student.updatedAt) ||
      student.updatedAt,
  }
}

function pickLatestStudentHistory(historyItems: TeacherStudentHistoryItem[]): TeacherStudentHistoryItem | null {
  if (!historyItems.length) {
    return null
  }

  return [...historyItems].sort((left, right) => {
    return compareTimestampText(
      pickFirstNonEmpty(right.hintCreatedAt, right.reportedAt),
      pickFirstNonEmpty(left.hintCreatedAt, left.reportedAt),
    )
  })[0] ?? null
}

function normalizeReleases(payload: unknown): TeacherRelease[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) => ({
    id: String(item.id ?? ''),
    classroomId: pickFirstNonEmpty(item.classroomId),
    title: String(item.title ?? ''),
    goal: String(item.goal ?? ''),
    description: String(item.description ?? ''),
    className: pickFirstNonEmpty(item.classroomName, item.className) || '未分组',
    status: normalizeReleaseStatus(item.status),
    analysisStatus: String(item.analysisStatus ?? 'pending'),
    studentCount: Number(item.studentCount ?? 0),
    updatedAt: String(item.updatedAt ?? '—'),
  }))
}

function normalizeReleaseDetail(payload: unknown): TeacherReleaseDetail {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    id: String(record.id ?? ''),
    title: String(record.title ?? ''),
    goal: String(record.goal ?? ''),
    description: String(record.description ?? ''),
    status: normalizeReleaseStatus(record.status),
    analysisStatus: String(record.analysisStatus ?? 'pending'),
    roleNames: normalizeStringArray(record.roleNames),
    scriptCounts: normalizeNumberMap(record.scriptCounts),
    blockCounts: normalizeNumberMap(record.blockCounts),
    categoryCounts: normalizeNumberMap(record.categoryCounts),
    broadcastMessages: normalizeStringArray(record.broadcastMessages),
    variableNames: normalizeStringArray(record.variableNames),
    listNames: normalizeStringArray(record.listNames),
    extensions: normalizeStringArray(record.extensions),
    teachingPoints: normalizeStringArray(record.teachingPoints),
    assignedStudents: normalizeAssignedStudents(record.assignedStudents),
    updatedAt: String(record.updatedAt ?? '—'),
  }
}

function normalizeReleaseAnalysis(payload: unknown): TeacherReleaseAnalysis {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    assignmentId: String(record.assignmentId ?? ''),
    analysisStatus: String(record.analysisStatus ?? 'pending'),
    analysisErrorMessage: String(record.analysisErrorMessage ?? ''),
    roleNames: normalizeStringArray(record.roleNames),
    scriptCounts: normalizeNumberMap(record.scriptCounts),
    blockCounts: normalizeNumberMap(record.blockCounts),
    categoryCounts: normalizeNumberMap(record.categoryCounts),
    broadcastMessages: normalizeStringArray(record.broadcastMessages),
    variableNames: normalizeStringArray(record.variableNames),
    listNames: normalizeStringArray(record.listNames),
    extensions: normalizeStringArray(record.extensions),
    teachingPoints: normalizeStringArray(record.teachingPoints),
  }
}

function normalizeAssignedStudents(payload: unknown): TeacherReleaseAssignedStudent[] {
  if (!Array.isArray(payload)) {
    return []
  }

  return payload
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
    .map((item) => ({
      id: String(item.id ?? ''),
      username: String(item.username ?? ''),
      displayName: String(item.displayName ?? ''),
      status: String(item.status ?? ''),
    }))
}

function normalizeTeacherReleaseAssignmentResult(payload: unknown): TeacherReleaseAssignmentResult {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}
  const studentIds = Array.isArray(record.studentIds) ? record.studentIds : []

  return {
    assignmentId: String(record.assignmentId ?? ''),
    studentIds: studentIds.map((studentId) => String(studentId ?? '')),
    assignedCount: Number(record.assignedCount ?? studentIds.length ?? 0),
  }
}

function normalizeTeacherReleaseMutation(payload: unknown): TeacherReleaseMutationResult {
  const record = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {}

  return {
    id: String(record.id ?? ''),
    title: String(record.title ?? ''),
    status: normalizeReleaseStatus(record.status),
    analysisStatus: String(record.analysisStatus ?? 'pending'),
  }
}

function normalizeLiveDashboard(payload: unknown): LiveDashboardSnapshot {
  const record = (payload ?? {}) as Record<string, unknown>
  const students = Array.isArray(record.students) ? record.students : []

  return {
    releaseId: String(record.assignmentId ?? record.releaseId ?? ''),
    releaseTitle: String(record.assignmentTitle ?? record.releaseTitle ?? '实时看板'),
    updatedAt: String(record.updatedAt ?? '—'),
    students: students.map((student) => normalizeLiveStudent(student)).filter(Boolean) as LiveStudentSnapshot[],
  }
}

function normalizeLiveStudent(payload: unknown): LiveStudentSnapshot | null {
  if (!payload || typeof payload !== 'object') {
    return null
  }

  const record = payload as Record<string, unknown>
  return {
    id: String(record.studentId ?? record.id ?? ''),
    name: String(record.studentName ?? record.name ?? ''),
    progress: normalizeProgressValue(record.progress),
    status: pickFirstNonEmpty(record.status),
    currentTarget: pickFirstNonEmpty(record.currentTarget),
    stepSummary: pickFirstNonEmpty(record.stepSummary),
    latestAiHint: String(record.lastHintText ?? record.latestAiHint ?? '等待学生请求提示'),
    updatedAt: pickFirstNonEmpty(record.lastReportedAt, record.lastHintAt, record.updatedAt) || '—',
  }
}

function normalizeReleaseStatus(input: unknown): TeacherReleaseStatus {
  return input === 'published' || input === 'archived' ? input : 'draft'
}

function compareStudentsByUpdatedAt(left: TeacherStudent, right: TeacherStudent) {
  return compareTimestampText(right.updatedAt, left.updatedAt)
}

function compareTimestampText(left: string, right: string) {
  const leftTime = Date.parse(left)
  const rightTime = Date.parse(right)

  if (Number.isFinite(leftTime) && Number.isFinite(rightTime)) {
    return leftTime - rightTime
  }

  if (Number.isFinite(leftTime)) {
    return 1
  }

  if (Number.isFinite(rightTime)) {
    return -1
  }

  return left.localeCompare(right)
}

function pickFirstNonEmpty(...values: unknown[]): string {
  for (const value of values) {
    if (typeof value !== 'string') {
      continue
    }
    const trimmed = value.trim()
    if (trimmed) {
      return trimmed
    }
  }
  return ''
}

function normalizeStringMap(value: unknown): Record<string, string> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {}
  }

  return Object.fromEntries(
    Object.entries(value).map(([key, entryValue]) => [key, String(entryValue ?? '')]),
  )
}

function normalizeStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }

  return value.map((entry) => String(entry ?? ''))
}

function normalizeNumberMap(value: unknown): Record<string, number> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {}
  }

  return Object.fromEntries(
    Object.entries(value).map(([key, entryValue]) => [key, Number(entryValue ?? 0)]),
  )
}

function normalizeProgressValue(value: unknown): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed < 0) {
    return 0
  }
  return Math.round(parsed)
}
