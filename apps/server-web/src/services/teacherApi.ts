import { inject, type InjectionKey } from 'vue'
import { buildApiUrl, HttpError, requestJson, type FetchLike } from './http'

export interface TeacherLoginInput {
  username: string
  password: string
}

export interface TeacherSession {
  token: string
  teacherName: string
}

export interface TeacherStudent {
  id: string
  name: string
  className: string
  progress: number
  status?: string
  currentTarget?: string
  stepSummary?: string
  latestAiHint: string
  updatedAt: string
}

export type TeacherReleaseStatus = 'draft' | 'published' | 'archived'

export interface TeacherRelease {
  id: string
  title: string
  className: string
  status: TeacherReleaseStatus
  studentCount: number
  updatedAt: string
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
  listStudents(): Promise<TeacherStudent[]>
  listReleases(): Promise<TeacherRelease[]>
  getLiveDashboard(releaseId: string): Promise<LiveDashboardSnapshot>
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

  return {
    async login(input) {
      return requestJson<TeacherSession>(
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
    async listReleases() {
      const payload = await requestAuthedJson<unknown>('/api/teacher/assignments')
      return normalizeReleases(payload)
    },
    async getLiveDashboard(releaseId) {
      const payload = await requestAuthedJson<unknown>(
        `/api/teacher/dashboard/assignments/${releaseId}/live`,
      )
      return normalizeLiveDashboard(payload)
    },
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

function normalizeStudents(payload: unknown): TeacherStudent[] {
  return normalizeCollection<Record<string, unknown>>(payload).map((item) => ({
    id: String(item.id ?? ''),
    name: String(item.displayName ?? item.name ?? ''),
    className: '未分组',
    progress: 0,
    status: '',
    currentTarget: '',
    stepSummary: '',
    latestAiHint: '等待学生请求提示',
    updatedAt: String(item.createdAt ?? item.updatedAt ?? '—'),
  }))
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
    title: String(item.title ?? ''),
    className: '未分组',
    status: normalizeReleaseStatus(item.status),
    studentCount: Number(item.studentCount ?? 0),
    updatedAt: String(item.updatedAt ?? '—'),
  }))
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

function normalizeProgressValue(value: unknown): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed < 0) {
    return 0
  }
  return Math.round(parsed)
}
