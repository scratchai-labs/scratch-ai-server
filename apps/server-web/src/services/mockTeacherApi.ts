import {
  type AdminOverview,
  TeacherApiError,
  type LiveDashboardSnapshot,
  type ManagedStudent,
  type ManagedTeacher,
  type ManagedTeacherRole,
  type TeacherApiClient,
  type TeacherLoginInput,
  type TeacherRelease,
  type TeacherSession,
  type TeacherStudent,
} from './teacherApi'

const demoSession: TeacherSession = {
  token: 'mock-session-token',
  teacherName: '王老师',
  role: 'teacher',
}

const demoAdminSession: TeacherSession = {
  token: 'mock-admin-token',
  teacherName: '系统管理员',
  role: 'admin',
}

const demoStudents: TeacherStudent[] = [
  {
    id: 'stu-1',
    name: 'Ada',
    className: '四年级一班',
    progress: 72,
    latestAiHint: '补上广播消息后再测试一次',
    updatedAt: '2026-05-07 09:20',
  },
  {
    id: 'stu-2',
    name: 'Alan',
    className: '四年级二班',
    progress: 38,
    latestAiHint: '先把重复积木整理成三个步骤',
    updatedAt: '2026-05-07 09:24',
  },
  {
    id: 'stu-3',
    name: 'Mia',
    className: '四年级一班',
    progress: 55,
    latestAiHint: '把下一步提示做成可复用流程',
    updatedAt: '2026-05-07 09:27',
  },
]

const demoReleases: TeacherRelease[] = [
  {
    id: 'rel-1',
    title: '第一期发布单',
    className: '四年级一班',
    status: 'published',
    studentCount: 24,
    updatedAt: '2026-05-07 09:10',
  },
  {
    id: 'rel-2',
    title: '第二期发布单',
    className: '四年级二班',
    status: 'draft',
    studentCount: 18,
    updatedAt: '2026-05-07 09:30',
  },
]

const demoSnapshots: Record<string, LiveDashboardSnapshot[]> = {
  'rel-1': [
    {
      releaseId: 'rel-1',
      releaseTitle: '第一期发布单',
      updatedAt: '2026-05-07 09:40',
      students: [
        {
          id: 'stu-1',
          name: 'Ada',
          progress: 42,
          latestAiHint: '先把绿旗事件连起来',
          updatedAt: '2026-05-07 09:40',
        },
        {
          id: 'stu-2',
          name: 'Alan',
          progress: 33,
          latestAiHint: '先整理重复执行的脚本块',
          updatedAt: '2026-05-07 09:40',
        },
      ],
    },
    {
      releaseId: 'rel-1',
      releaseTitle: '第一期发布单',
      updatedAt: '2026-05-07 09:44',
      students: [
        {
          id: 'stu-1',
          name: 'Ada',
          progress: 68,
          latestAiHint: '现在补上角色切换逻辑',
          updatedAt: '2026-05-07 09:44',
        },
        {
          id: 'stu-2',
          name: 'Alan',
          progress: 51,
          latestAiHint: '把等待和广播组合起来',
          updatedAt: '2026-05-07 09:44',
        },
      ],
    },
  ],
  'rel-2': [
    {
      releaseId: 'rel-2',
      releaseTitle: '第二期发布单',
      updatedAt: '2026-05-07 09:36',
      students: [
        {
          id: 'stu-3',
          name: 'Mia',
          progress: 24,
          latestAiHint: '先确认消息广播的接收端',
          updatedAt: '2026-05-07 09:36',
        },
      ],
    },
  ],
}

const managedTeachers: ManagedTeacher[] = [
  {
    id: '1',
    username: 'admin',
    role: 'admin',
    status: 'active',
    createdAt: '2026-06-13T12:00:00Z',
  },
  {
    id: '2',
    username: 'teacher',
    role: 'teacher',
    status: 'active',
    createdAt: '2026-06-13T12:10:00Z',
  },
]

const managedStudents: ManagedStudent[] = [
  {
    id: '10',
    teacherId: '2',
    teacherUsername: 'teacher',
    username: 'student-1',
    displayName: '小蓝',
    status: 'active',
    createdAt: '2026-06-13T12:20:00Z',
  },
  {
    id: '11',
    teacherId: '2',
    teacherUsername: 'teacher',
    username: 'student-2',
    displayName: '小橙',
    status: 'disabled',
    createdAt: '2026-06-13T12:22:00Z',
  },
]

export function createMockTeacherApiClient(): TeacherApiClient {
  const cursorByRelease = new Map<string, number>()
  const teachers = clone(managedTeachers)
  const students = clone(managedStudents)

  return {
    async login(input: TeacherLoginInput) {
      if (input.username === 'admin' && input.password === 'admin12345') {
        return clone(demoAdminSession)
      }
      if (input.username === 'teacher' && input.password === 'teach123') {
        return clone(demoSession)
      }

      throw new TeacherApiError('用户名或密码错误', 401)
    },
    async logout() {
      return undefined
    },
    async listStudents() {
      return clone(demoStudents)
    },
    async listReleases() {
      return clone(demoReleases)
    },
    async getLiveDashboard(releaseId: string) {
      const fallbackSnapshots = demoSnapshots['rel-1']!
      const snapshots =
        demoSnapshots[releaseId] ?? fallbackSnapshots
      const cursor = cursorByRelease.get(releaseId) ?? 0
      const index = Math.min(cursor, snapshots.length - 1)
      cursorByRelease.set(releaseId, cursor + 1)
      return clone(snapshots[index] ?? snapshots[snapshots.length - 1]!)
    },
    async getAdminOverview() {
      return buildAdminOverview(teachers, students)
    },
    async listTeachers() {
      return clone(teachers)
    },
    async createTeacher(input) {
      const nextTeacher = {
        id: String(teachers.length + 1),
        username: input.username,
        role: 'teacher',
        status: 'active',
        createdAt: '2026-06-13T12:30:00Z',
      } satisfies ManagedTeacher
      teachers.push(nextTeacher)
      return clone(nextTeacher)
    },
    async resetTeacherPassword(teacherId) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      return clone(target)
    },
    async changeTeacherRole(teacherId, role: ManagedTeacherRole) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      if (target.username === 'admin' && role !== 'admin') {
        throw new TeacherApiError('admin cannot change own role', 409)
      }
      target.role = role
      return clone(target)
    },
    async disableTeacher(teacherId) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      target.status = 'disabled'
      return clone(target)
    },
    async enableTeacher(teacherId) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      target.status = 'active'
      return clone(target)
    },
    async listManagedStudents() {
      return clone(students)
    },
    async createManagedStudent(input) {
      const teacher = teachers.find(
        (item) => item.id === input.teacherId && item.role === 'teacher',
      )
      if (!teacher) {
        throw new TeacherApiError('teacher not found', 404)
      }
      const nextStudent = {
        id: String(students.length + 10),
        teacherId: input.teacherId,
        teacherUsername: teacher.username,
        username: input.username,
        displayName: input.displayName,
        status: 'active',
        createdAt: '2026-06-14T10:00:00Z',
      } satisfies ManagedStudent
      students.push(nextStudent)
      return clone(nextStudent)
    },
    async resetManagedStudentPassword(studentId) {
      const target = students.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      return clone(target)
    },
    async disableManagedStudent(studentId) {
      const target = students.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      target.status = 'disabled'
      return clone(target)
    },
    async enableManagedStudent(studentId) {
      const target = students.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      target.status = 'active'
      return clone(target)
    },
  }
}

function buildAdminOverview(teachers: ManagedTeacher[], students: ManagedStudent[]): AdminOverview {
  return {
    adminCount: teachers.filter((teacher) => teacher.role === 'admin').length,
    teacherCount: teachers.filter((teacher) => teacher.role !== 'admin').length,
    activeTeacherCount: teachers.filter(
      (teacher) => teacher.role !== 'admin' && teacher.status !== 'disabled',
    ).length,
    disabledTeacherCount: teachers.filter(
      (teacher) => teacher.role !== 'admin' && teacher.status === 'disabled',
    ).length,
    studentCount: students.length,
    activeStudentCount: students.filter((student) => student.status !== 'disabled').length,
    disabledStudentCount: students.filter((student) => student.status === 'disabled').length,
  }
}

function clone<T>(value: T): T {
  return structuredClone(value)
}
