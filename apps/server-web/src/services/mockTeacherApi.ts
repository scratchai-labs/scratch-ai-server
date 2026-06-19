import {
  type AdminAuditLog,
  type AdminOverview,
  type BatchCreateTeacherStudentsResult,
  type CreateTeacherClassroomInput,
  type CreateTeacherStudentInput,
  TeacherApiError,
  type LiveDashboardSnapshot,
  type ManagedStudent,
  type ManagedTeacher,
  type TeacherClassroom,
  type ManagedTeacherRole,
  type TeacherApiClient,
  type TeacherLoginInput,
  type TeacherReleaseAnalysis,
  type TeacherReleaseDetail,
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
    username: 'student-ada',
    name: 'Ada',
    className: '四年级一班',
    progress: 72,
    latestAiHint: '补上广播消息后再测试一次',
    updatedAt: '2026-05-07 09:20',
    createdAt: '2026-05-07 09:20',
  },
  {
    id: 'stu-2',
    username: 'student-alan',
    name: 'Alan',
    className: '四年级二班',
    progress: 38,
    latestAiHint: '先把重复积木整理成三个步骤',
    updatedAt: '2026-05-07 09:24',
    createdAt: '2026-05-07 09:24',
  },
  {
    id: 'stu-3',
    username: 'student-mia',
    name: 'Mia',
    className: '四年级一班',
    progress: 55,
    latestAiHint: '把下一步提示做成可复用流程',
    updatedAt: '2026-05-07 09:27',
    createdAt: '2026-05-07 09:27',
  },
]

const demoClassrooms: TeacherClassroom[] = [
  {
    id: 'class-1',
    name: '四年级一班',
    studentCount: 2,
    projectCount: 1,
    createdAt: '2026-05-07T09:00:00Z',
    updatedAt: '2026-05-07T09:30:00Z',
  },
  {
    id: 'class-2',
    name: '四年级二班',
    studentCount: 1,
    projectCount: 1,
    createdAt: '2026-05-07T09:05:00Z',
    updatedAt: '2026-05-07T09:35:00Z',
  },
]

const demoClassroomStudents: Record<string, TeacherStudent[]> = {
  'class-1': [
    {
      id: 'class-1-stu-1',
      classroomId: 'class-1',
      username: 'student-ada',
      name: 'Ada',
      className: '四年级一班',
      progress: 72,
      status: 'active',
      currentTarget: '让角色按事件响应',
      stepSummary: '已经接上绿旗事件',
      latestAiHint: '补上广播消息后再测试一次',
      updatedAt: '2026-05-07 09:20',
      createdAt: '2026-05-07 09:20',
    },
    {
      id: 'class-1-stu-2',
      classroomId: 'class-1',
      username: 'student-mia',
      name: 'Mia',
      className: '四年级一班',
      progress: 38,
      status: 'assigned',
      currentTarget: '先把角色移动起来',
      stepSummary: '已经摆好起始位置',
      latestAiHint: '把移动积木接成完整流程',
      updatedAt: '2026-05-07 09:24',
      createdAt: '2026-05-07 09:24',
    },
  ],
  'class-2': [
    {
      id: 'class-2-stu-1',
      classroomId: 'class-2',
      username: 'student-alan',
      name: 'Alan',
      className: '四年级二班',
      progress: 55,
      status: 'active',
      currentTarget: '补上重复执行',
      stepSummary: '已经接上广播消息',
      latestAiHint: '先整理重复执行的脚本块',
      updatedAt: '2026-05-07 09:27',
      createdAt: '2026-05-07 09:27',
    },
  ],
}

const demoClassroomProjects: Record<string, TeacherRelease[]> = {
  'class-1': [
    {
      id: 'rel-1',
      classroomId: 'class-1',
      title: '迷宫项目',
      goal: '让角色按事件响应',
      description: '第一节课项目',
      className: '四年级一班',
      status: 'draft',
      analysisStatus: 'ready',
      studentCount: 2,
      updatedAt: '2026-05-07T09:30:00Z',
    },
  ],
  'class-2': [
    {
      id: 'rel-2',
      classroomId: 'class-2',
      title: '追逐项目',
      goal: '补齐广播与重复执行',
      description: '第二节课项目',
      className: '四年级二班',
      status: 'published',
      analysisStatus: 'ready',
      studentCount: 1,
      updatedAt: '2026-05-07T09:35:00Z',
    },
  ],
}

const demoReleases: TeacherRelease[] = [
  {
    id: 'rel-1',
    title: '第一期发布单',
    goal: '让角色按事件响应',
    description: '第一节课任务',
    className: '四年级一班',
    status: 'published',
    analysisStatus: 'ready',
    studentCount: 24,
    updatedAt: '2026-05-07 09:10',
  },
  {
    id: 'rel-2',
    title: '第二期发布单',
    goal: '补齐广播与重复执行',
    description: '第二节课任务',
    className: '四年级二班',
    status: 'draft',
    analysisStatus: 'ready',
    studentCount: 18,
    updatedAt: '2026-05-07 09:30',
  },
]

const demoReleaseDetails: Record<string, TeacherReleaseDetail> = {
  'rel-1': {
    id: 'rel-1',
    title: '迷宫项目',
    goal: '让角色按事件响应',
    description: '第一节课项目',
    status: 'published',
    analysisStatus: 'ready',
    roleNames: ['Stage', 'Cat'],
    scriptCounts: { Stage: 1, Cat: 2 },
    blockCounts: { event_whenflagclicked: 1, motion_movesteps: 1 },
    categoryCounts: { event: 1, motion: 1 },
    broadcastMessages: ['开始'],
    variableNames: ['score'],
    listNames: ['targets'],
    extensions: ['pen'],
    teachingPoints: ['先搭好事件入口', '再补动作流程'],
    assignedStudents: [
      {
        id: 'stu-1',
        username: 'student-ada',
        displayName: 'Ada',
        status: 'active',
      },
      {
        id: 'stu-2',
        username: 'student-alan',
        displayName: 'Alan',
        status: 'active',
      },
    ],
    updatedAt: '2026-05-07 09:10',
  },
  'rel-2': {
    id: 'rel-2',
    title: '追逐项目',
    goal: '补齐广播与重复执行',
    description: '第二节课项目',
    status: 'draft',
    analysisStatus: 'ready',
    roleNames: ['Stage', 'Mia'],
    scriptCounts: { Stage: 1, Mia: 1 },
    blockCounts: { event_whenflagclicked: 1, control_repeat: 1 },
    categoryCounts: { event: 1, control: 1 },
    broadcastMessages: ['准备'],
    variableNames: [],
    listNames: [],
    extensions: [],
    teachingPoints: ['先确认广播名', '再把重复动作拆成循环'],
    assignedStudents: [
      {
        id: 'stu-3',
        username: 'student-mia',
        displayName: 'Mia',
        status: 'active',
      },
    ],
    updatedAt: '2026-05-07 09:30',
  },
}

const demoSnapshots: Record<string, LiveDashboardSnapshot[]> = {
  'rel-1': [
    {
      releaseId: 'rel-1',
      releaseTitle: '迷宫项目',
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
      releaseTitle: '迷宫项目',
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
      releaseTitle: '追逐项目',
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

const initialAuditLogs: AdminAuditLog[] = [
  {
    id: '4',
    actorUsername: 'admin',
    action: 'teacher.role_change',
    targetType: 'teacher',
    targetId: '2',
    targetUsername: 'teacher',
    before: { role: 'teacher' },
    after: { role: 'admin' },
    createdAt: '2026-06-14T10:06:00Z',
  },
  {
    id: '3',
    actorUsername: 'admin',
    action: 'student.disable',
    targetType: 'student',
    targetId: '11',
    targetUsername: 'student-2',
    before: { status: 'active' },
    after: { status: 'disabled' },
    createdAt: '2026-06-14T10:04:00Z',
  },
  {
    id: '2',
    actorUsername: 'admin',
    action: 'student.create',
    targetType: 'student',
    targetId: '10',
    targetUsername: 'student-1',
    before: {},
    after: {
      username: 'student-1',
      displayName: '小蓝',
      status: 'active',
      teacherUsername: 'teacher',
    },
    createdAt: '2026-06-14T10:02:00Z',
  },
  {
    id: '1',
    actorUsername: 'admin',
    action: 'teacher.create',
    targetType: 'teacher',
    targetId: '2',
    targetUsername: 'teacher',
    before: {},
    after: {
      username: 'teacher',
      role: 'teacher',
      status: 'active',
    },
    createdAt: '2026-06-14T10:00:00Z',
  },
]

function nextTeacherStudentId(students: TeacherStudent[]): string {
  return `stu-${students.length + 1}`
}

function buildTeacherStudentRecord(
  students: TeacherStudent[],
  input: CreateTeacherStudentInput,
  options: {
    classroomId?: string
    className?: string
  } = {},
): TeacherStudent {
  const now = new Date().toISOString()

  return {
    id: nextTeacherStudentId(students),
    username: input.username,
    name: input.displayName,
    className: options.className ?? '未分组',
    classroomId: options.classroomId,
    progress: 0,
    status: '',
    currentTarget: '',
    stepSummary: '',
    latestAiHint: '等待学生请求提示',
    updatedAt: now,
    createdAt: now,
  }
}

function batchCreateTeacherStudents(
  students: TeacherStudent[],
  inputs: CreateTeacherStudentInput[],
  options: {
    classroomId?: string
    className?: string
  } = {},
): BatchCreateTeacherStudentsResult {
  const takenUsernames = new Set(students.map((student) => student.username))
  const created: TeacherStudent[] = []
  const conflicts: string[] = []

  for (const input of inputs) {
    if (takenUsernames.has(input.username)) {
      conflicts.push(input.username)
      continue
    }

    const nextStudent = buildTeacherStudentRecord(students, input, options)
    students.push(nextStudent)
    takenUsernames.add(input.username)
    created.push(clone(nextStudent))
  }

  return {
    created,
    conflicts,
  }
}

export function createMockTeacherApiClient(): TeacherApiClient {
  const cursorByRelease = new Map<string, number>()
  const teachers = clone(managedTeachers)
  const teacherStudents = clone(demoStudents)
  const releases = clone(demoReleases)
  const releaseDetails = clone(demoReleaseDetails)
  const students = clone(managedStudents)
  const auditLogs = clone(initialAuditLogs)
  const classrooms = clone(demoClassrooms)
  const classroomStudents = clone(demoClassroomStudents)
  const classroomProjects = clone(demoClassroomProjects)
  const snapshotsByRelease = clone(demoSnapshots)

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
      return clone(teacherStudents)
    },
    async listClassrooms() {
      return clone(classrooms)
    },
    async createClassroom(input) {
      const now = new Date().toISOString()
      const nextClassroom: TeacherClassroom = {
        id: `class-${classrooms.length + 1}`,
        name: input.name,
        studentCount: 0,
        projectCount: 0,
        createdAt: now,
        updatedAt: now,
      }
      classrooms.push(nextClassroom)
      classroomStudents[nextClassroom.id] = []
      classroomProjects[nextClassroom.id] = []
      return clone(nextClassroom)
    },
    async getClassroomDetail(classroomId) {
      const classroom = classrooms.find((item) => item.id === classroomId)
      if (!classroom) {
        throw new TeacherApiError('classroom not found', 404)
      }
      return clone(classroom)
    },
    async listClassroomStudents(classroomId) {
      return clone(classroomStudents[classroomId] ?? [])
    },
    async createClassroomStudent(classroomId, input) {
      const classroom = classrooms.find((item) => item.id === classroomId)
      if (!classroom) {
        throw new TeacherApiError('classroom not found', 404)
      }

      const nextStudent = buildTeacherStudentRecord(classroomStudents[classroomId] ?? [], input, {
        classroomId,
        className: classroom.name,
      })
      classroomStudents[classroomId] = [...(classroomStudents[classroomId] ?? []), nextStudent]
      classroom.studentCount = classroomStudents[classroomId].length
      classroom.updatedAt = new Date().toISOString()
      return clone(nextStudent)
    },
    async batchCreateClassroomStudents(classroomId, input) {
      const classroom = classrooms.find((item) => item.id === classroomId)
      if (!classroom) {
        throw new TeacherApiError('classroom not found', 404)
      }

      const nextStudents = classroomStudents[classroomId] ?? []
      const result = batchCreateTeacherStudents(nextStudents, input, {
        classroomId,
        className: classroom.name,
      })
      classroomStudents[classroomId] = nextStudents
      classroom.studentCount = nextStudents.length
      classroom.updatedAt = new Date().toISOString()
      return result
    },
    async listClassroomProjects(classroomId) {
      return clone(classroomProjects[classroomId] ?? [])
    },
    async createClassroomProject(classroomId, input) {
      const classroom = classrooms.find((item) => item.id === classroomId)
      if (!classroom) {
        throw new TeacherApiError('classroom not found', 404)
      }

      const nextProjectId = `rel-${releases.length + 1}`
      const updatedAt = new Date().toISOString()
      const nextProject: TeacherRelease = {
        id: nextProjectId,
        classroomId,
        title: input.title,
        goal: input.goal,
        description: input.description,
        className: classroom.name,
        status: 'draft',
        analysisStatus: 'pending',
        studentCount: classroomStudents[classroomId]?.length ?? 0,
        updatedAt,
      }
      releases.push(nextProject)
      releaseDetails[nextProjectId] = {
        id: nextProjectId,
        title: nextProject.title,
        goal: nextProject.goal,
        description: nextProject.description,
        status: nextProject.status,
        analysisStatus: nextProject.analysisStatus,
        roleNames: [],
        scriptCounts: {},
        blockCounts: {},
        categoryCounts: {},
        broadcastMessages: [],
        variableNames: [],
        listNames: [],
        extensions: [],
        teachingPoints: [],
        assignedStudents: [],
        updatedAt,
      }
      snapshotsByRelease[nextProjectId] = [
        {
          releaseId: nextProjectId,
          releaseTitle: nextProject.title,
          updatedAt,
          students: (classroomStudents[classroomId] ?? []).map((student) => ({
            id: student.id,
            name: student.name,
            progress: student.progress,
            status: student.status,
            currentTarget: student.currentTarget,
            stepSummary: student.stepSummary,
            latestAiHint: student.latestAiHint,
            updatedAt: student.updatedAt,
          })),
        },
      ]
      classroomProjects[classroomId] = [...(classroomProjects[classroomId] ?? []), nextProject]
      classroom.projectCount = classroomProjects[classroomId].length
      classroom.updatedAt = updatedAt
      return {
        id: nextProject.id,
        title: nextProject.title,
        status: nextProject.status,
        analysisStatus: nextProject.analysisStatus,
      }
    },
    async createStudent(input) {
      const result = batchCreateTeacherStudents(teacherStudents, [input])
      const createdStudent = result.created[0]
      if (!createdStudent) {
        throw new TeacherApiError(`学生账号冲突：${result.conflicts.join('、')}`, 409)
      }
      return clone(createdStudent)
    },
    async batchCreateStudents(input) {
      return batchCreateTeacherStudents(teacherStudents, input)
    },
    async resetStudentPassword(studentId) {
      const target = teacherStudents.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      return clone(target)
    },
    async listReleases() {
      return clone(releases)
    },
    async createRelease(input) {
      const releaseId = `rel-${releases.length + 1}`
      const updatedAt = new Date().toISOString()
      const nextRelease = {
        id: releaseId,
        title: input.title,
        goal: input.goal,
        description: input.description,
        className: '未分组',
        status: 'draft',
        analysisStatus: 'pending',
        studentCount: 0,
        updatedAt,
      } satisfies TeacherRelease
      releases.push(nextRelease)
      releaseDetails[releaseId] = {
        id: releaseId,
        title: input.title,
        goal: input.goal,
        description: input.description,
        status: 'draft',
        analysisStatus: 'pending',
        roleNames: [],
        scriptCounts: {},
        blockCounts: {},
        categoryCounts: {},
        broadcastMessages: [],
        variableNames: [],
        listNames: [],
        extensions: [],
        teachingPoints: [],
        assignedStudents: [],
        updatedAt,
      }
      return {
        id: releaseId,
        title: input.title,
        status: 'draft',
        analysisStatus: 'pending',
      }
    },
    async getReleaseDetail(releaseId) {
      const detail = releaseDetails[releaseId]
      if (!detail) {
        throw new TeacherApiError('assignment not found', 404)
      }
      return clone(detail)
    },
    async getReleaseAnalysis(releaseId) {
      const detail = releaseDetails[releaseId]
      if (!detail) {
        throw new TeacherApiError('assignment not found', 404)
      }
      return clone(toReleaseAnalysis(detail))
    },
    async assignStudentsToRelease(releaseId, studentIds) {
      const detail = releaseDetails[releaseId]
      const release = releases.find((item) => item.id === releaseId)
      if (!detail || !release) {
        throw new TeacherApiError('assignment not found', 404)
      }
      detail.assignedStudents = teacherStudents
        .filter((student) => studentIds.includes(student.id))
        .map((student) => ({
          id: student.id,
          username: student.username,
          displayName: student.name,
          status: student.status || 'active',
        }))
      detail.updatedAt = new Date().toISOString()
      release.studentCount = detail.assignedStudents.length
      release.updatedAt = detail.updatedAt
      return {
        assignmentId: releaseId,
        studentIds,
        assignedCount: studentIds.length,
      }
    },
    async publishRelease(releaseId) {
      const detail = releaseDetails[releaseId]
      const release = releases.find((item) => item.id === releaseId)
      if (!detail || !release) {
        throw new TeacherApiError('assignment not found', 404)
      }
      if (detail.analysisStatus !== 'ready') {
        throw new TeacherApiError('assignment analysis not ready', 409)
      }
      detail.status = 'published'
      detail.updatedAt = new Date().toISOString()
      release.status = 'published'
      release.updatedAt = detail.updatedAt
      return {
        id: release.id,
        title: release.title,
        status: release.status,
        analysisStatus: detail.analysisStatus,
      }
    },
    async archiveRelease(releaseId) {
      const detail = releaseDetails[releaseId]
      const release = releases.find((item) => item.id === releaseId)
      if (!detail || !release) {
        throw new TeacherApiError('assignment not found', 404)
      }
      detail.status = 'archived'
      detail.updatedAt = new Date().toISOString()
      release.status = 'archived'
      release.updatedAt = detail.updatedAt
      return {
        id: release.id,
        title: release.title,
        status: release.status,
        analysisStatus: detail.analysisStatus,
      }
    },
    async getLiveDashboard(releaseId: string) {
      const fallbackSnapshots = snapshotsByRelease['rel-1']!
      const snapshots =
        snapshotsByRelease[releaseId] ?? fallbackSnapshots
      const cursor = cursorByRelease.get(releaseId) ?? 0
      const index = Math.min(cursor, snapshots.length - 1)
      cursorByRelease.set(releaseId, cursor + 1)
      return clone(snapshots[index] ?? snapshots[snapshots.length - 1]!)
    },
    async getAdminOverview() {
      return buildAdminOverview(teachers, students)
    },
    async listAdminAuditLogs() {
      return clone(auditLogs)
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
      recordAuditLog(auditLogs, {
        action: 'teacher.create',
        targetType: 'teacher',
        targetId: nextTeacher.id,
        targetUsername: nextTeacher.username,
        before: {},
        after: {
          username: nextTeacher.username,
          role: nextTeacher.role,
          status: nextTeacher.status,
        },
      })
      return clone(nextTeacher)
    },
    async resetTeacherPassword(teacherId) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      recordAuditLog(auditLogs, {
        action: 'teacher.password_reset',
        targetType: 'teacher',
        targetId: target.id,
        targetUsername: target.username,
        before: {},
        after: { passwordStatus: 'updated' },
      })
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
      const beforeRole = target.role
      target.role = role
      recordAuditLog(auditLogs, {
        action: 'teacher.role_change',
        targetType: 'teacher',
        targetId: target.id,
        targetUsername: target.username,
        before: { role: beforeRole },
        after: { role: target.role },
      })
      return clone(target)
    },
    async disableTeacher(teacherId) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      const beforeStatus = target.status
      target.status = 'disabled'
      recordAuditLog(auditLogs, {
        action: 'teacher.disable',
        targetType: 'teacher',
        targetId: target.id,
        targetUsername: target.username,
        before: { status: beforeStatus },
        after: { status: target.status },
      })
      return clone(target)
    },
    async enableTeacher(teacherId) {
      const target = teachers.find((teacher) => teacher.id === teacherId)
      if (!target) {
        throw new TeacherApiError('teacher not found', 404)
      }
      const beforeStatus = target.status
      target.status = 'active'
      recordAuditLog(auditLogs, {
        action: 'teacher.enable',
        targetType: 'teacher',
        targetId: target.id,
        targetUsername: target.username,
        before: { status: beforeStatus },
        after: { status: target.status },
      })
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
      recordAuditLog(auditLogs, {
        action: 'student.create',
        targetType: 'student',
        targetId: nextStudent.id,
        targetUsername: nextStudent.username,
        before: {},
        after: {
          username: nextStudent.username,
          displayName: nextStudent.displayName,
          status: nextStudent.status,
          teacherUsername: nextStudent.teacherUsername,
        },
      })
      return clone(nextStudent)
    },
    async resetManagedStudentPassword(studentId) {
      const target = students.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      recordAuditLog(auditLogs, {
        action: 'student.password_reset',
        targetType: 'student',
        targetId: target.id,
        targetUsername: target.username,
        before: {},
        after: { passwordStatus: 'updated' },
      })
      return clone(target)
    },
    async disableManagedStudent(studentId) {
      const target = students.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      const beforeStatus = target.status
      target.status = 'disabled'
      recordAuditLog(auditLogs, {
        action: 'student.disable',
        targetType: 'student',
        targetId: target.id,
        targetUsername: target.username,
        before: { status: beforeStatus },
        after: { status: target.status },
      })
      return clone(target)
    },
    async enableManagedStudent(studentId) {
      const target = students.find((student) => student.id === studentId)
      if (!target) {
        throw new TeacherApiError('student not found', 404)
      }
      const beforeStatus = target.status
      target.status = 'active'
      recordAuditLog(auditLogs, {
        action: 'student.enable',
        targetType: 'student',
        targetId: target.id,
        targetUsername: target.username,
        before: { status: beforeStatus },
        after: { status: target.status },
      })
      return clone(target)
    },
  }
}

function toReleaseAnalysis(detail: TeacherReleaseDetail): TeacherReleaseAnalysis {
  return {
    assignmentId: detail.id,
    analysisStatus: detail.analysisStatus,
    analysisErrorMessage: '',
    roleNames: detail.roleNames,
    scriptCounts: detail.scriptCounts,
    blockCounts: detail.blockCounts,
    categoryCounts: detail.categoryCounts,
    broadcastMessages: detail.broadcastMessages,
    variableNames: detail.variableNames,
    listNames: detail.listNames,
    extensions: detail.extensions,
    teachingPoints: detail.teachingPoints,
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

function recordAuditLog(
  auditLogs: AdminAuditLog[],
  input: Omit<AdminAuditLog, 'id' | 'actorUsername' | 'createdAt'>,
) {
  const nextID = String(Number(auditLogs[0]?.id ?? '0') + 1)
  auditLogs.unshift({
    id: nextID,
    actorUsername: 'admin',
    createdAt: new Date().toISOString(),
    ...input,
  })
}
