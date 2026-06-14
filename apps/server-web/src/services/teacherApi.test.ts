import { describe, expect, it, vi } from 'vitest'
import { createFetchTeacherApiClient } from './teacherApi'

function createFetchResponse(body: unknown) {
  return {
    ok: true,
    status: 200,
    json: async () => body,
  }
}

describe('createFetchTeacherApiClient', () => {
  it('posts teacher login to /api/teacher/login', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(
      createFetchResponse({
        token: 'token-1',
        teacherName: '王老师',
        role: 'teacher',
      }),
    )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
    })

    const session = await api.login({
      username: 'teacher',
      password: 'teach123',
    })

    expect(fetchImpl).toHaveBeenCalledWith(
      'https://teacher.example/api/teacher/login',
      expect.objectContaining({
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'teacher',
          password: 'teach123',
        }),
      }),
    )
    expect(session).toEqual({
      token: 'token-1',
      teacherName: '王老师',
      role: 'teacher',
    })
  })

  it('reads and mutates managed teachers from admin endpoints', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          items: [
            {
              id: 1,
              username: 'admin',
              role: 'admin',
              status: 'active',
              createdAt: '2026-06-13T12:00:00Z',
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 2,
          username: 'teacher-1',
          role: 'teacher',
          status: 'active',
          createdAt: '2026-06-13T12:01:00Z',
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 2,
          username: 'teacher-1',
          role: 'teacher',
          status: 'disabled',
          createdAt: '2026-06-13T12:01:00Z',
        }),
      )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'admin-token',
    })

    await expect(api.listTeachers()).resolves.toEqual([
      {
        id: '1',
        username: 'admin',
        role: 'admin',
        status: 'active',
        createdAt: '2026-06-13T12:00:00Z',
      },
    ])

    await expect(
      api.createTeacher({
        username: 'teacher-1',
        initialPassword: 'secret123',
      }),
    ).resolves.toEqual({
      id: '2',
      username: 'teacher-1',
      role: 'teacher',
      status: 'active',
      createdAt: '2026-06-13T12:01:00Z',
    })

    await expect(api.disableTeacher('2')).resolves.toEqual({
      id: '2',
      username: 'teacher-1',
      role: 'teacher',
      status: 'disabled',
      createdAt: '2026-06-13T12:01:00Z',
    })

    expect(fetchImpl).toHaveBeenNthCalledWith(
      1,
      'https://teacher.example/api/admin/teachers',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer admin-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      2,
      'https://teacher.example/api/admin/teachers',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer admin-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'teacher-1',
          initialPassword: 'secret123',
        }),
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      3,
      'https://teacher.example/api/admin/teachers/2/disable',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer admin-token',
        },
      }),
    )
  })

  it('changes a managed teacher role from the admin endpoint', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(
      createFetchResponse({
        id: 2,
        username: 'teacher-1',
        role: 'admin',
        status: 'active',
        createdAt: '2026-06-13T12:01:00Z',
      }),
    )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'admin-token',
    }) as ReturnType<typeof createFetchTeacherApiClient> & {
      changeTeacherRole?: (teacherId: string, role: string) => Promise<unknown>
    }

    expect(api.changeTeacherRole).toBeTypeOf('function')
    if (!api.changeTeacherRole) {
      return
    }

    await expect(api.changeTeacherRole('2', 'admin')).resolves.toEqual({
      id: '2',
      username: 'teacher-1',
      role: 'admin',
      status: 'active',
      createdAt: '2026-06-13T12:01:00Z',
    })

    expect(fetchImpl).toHaveBeenCalledWith(
      'https://teacher.example/api/admin/teachers/2/role',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer admin-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          role: 'admin',
        }),
      }),
    )
  })

  it('reads admin overview and managed students from admin endpoints', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          adminCount: 1,
          teacherCount: 2,
          activeTeacherCount: 2,
          disabledTeacherCount: 0,
          studentCount: 24,
          activeStudentCount: 22,
          disabledStudentCount: 2,
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          items: [
            {
              id: 5,
              teacherId: 2,
              teacherUsername: 'teacher-1',
              username: 'student-1',
              displayName: '小蓝',
              status: 'active',
              createdAt: '2026-06-13T12:05:00Z',
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 5,
          teacherId: 2,
          teacherUsername: 'teacher-1',
          username: 'student-1',
          displayName: '小蓝',
          status: 'disabled',
          createdAt: '2026-06-13T12:05:00Z',
        }),
      )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'admin-token',
    })

    await expect(api.getAdminOverview?.()).resolves.toEqual({
      adminCount: 1,
      teacherCount: 2,
      activeTeacherCount: 2,
      disabledTeacherCount: 0,
      studentCount: 24,
      activeStudentCount: 22,
      disabledStudentCount: 2,
    })

    await expect(api.listManagedStudents?.()).resolves.toEqual([
      {
        id: '5',
        teacherId: '2',
        teacherUsername: 'teacher-1',
        username: 'student-1',
        displayName: '小蓝',
        status: 'active',
        createdAt: '2026-06-13T12:05:00Z',
      },
    ])

    await expect(api.disableManagedStudent?.('5')).resolves.toEqual({
      id: '5',
      teacherId: '2',
      teacherUsername: 'teacher-1',
      username: 'student-1',
      displayName: '小蓝',
      status: 'disabled',
      createdAt: '2026-06-13T12:05:00Z',
    })

    expect(fetchImpl).toHaveBeenNthCalledWith(
      1,
      'https://teacher.example/api/admin/overview',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer admin-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      2,
      'https://teacher.example/api/admin/students',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer admin-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      3,
      'https://teacher.example/api/admin/students/5/disable',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer admin-token',
        },
      }),
    )
  })

  it('reads admin audit logs from the admin endpoint', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(
      createFetchResponse({
        items: [
          {
            id: 9,
            actorUsername: 'admin',
            action: 'teacher.role_change',
            targetType: 'teacher',
            targetId: 2,
            targetUsername: 'teacher-1',
            before: {
              role: 'teacher',
            },
            after: {
              role: 'admin',
            },
            createdAt: '2026-06-14T12:00:00Z',
          },
        ],
      }),
    )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'admin-token',
    }) as ReturnType<typeof createFetchTeacherApiClient> & {
      listAdminAuditLogs?: () => Promise<unknown>
    }

    expect(api.listAdminAuditLogs).toBeTypeOf('function')
    if (!api.listAdminAuditLogs) {
      return
    }

    await expect(api.listAdminAuditLogs()).resolves.toEqual([
      {
        id: '9',
        actorUsername: 'admin',
        action: 'teacher.role_change',
        targetType: 'teacher',
        targetId: '2',
        targetUsername: 'teacher-1',
        before: {
          role: 'teacher',
        },
        after: {
          role: 'admin',
        },
        createdAt: '2026-06-14T12:00:00Z',
      },
    ])

    expect(fetchImpl).toHaveBeenCalledWith(
      'https://teacher.example/api/admin/audit-logs',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer admin-token',
        },
      }),
    )
  })

  it('creates a managed student from the admin endpoint', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(
      createFetchResponse({
        id: 6,
        teacherId: 2,
        teacherUsername: 'teacher-1',
        username: 'student-2',
        displayName: '小绿',
        status: 'active',
        createdAt: '2026-06-14T10:00:00Z',
      }),
    )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'admin-token',
    })

    await expect(
      api.createManagedStudent?.({
        teacherId: '2',
        username: 'student-2',
        displayName: '小绿',
        initialPassword: 'stud1234',
      }),
    ).resolves.toEqual({
      id: '6',
      teacherId: '2',
      teacherUsername: 'teacher-1',
      username: 'student-2',
      displayName: '小绿',
      status: 'active',
      createdAt: '2026-06-14T10:00:00Z',
    })

    expect(fetchImpl).toHaveBeenCalledWith(
      'https://teacher.example/api/admin/students',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer admin-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          teacherId: 2,
          username: 'student-2',
          displayName: '小绿',
          initialPassword: 'stud1234',
        }),
      }),
    )
  })

  it('posts teacher logout to /api/teacher/logout', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(
      createFetchResponse({
        status: 'ok',
      }),
    )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    })

    await expect(api.logout()).resolves.toBeUndefined()

    expect(fetchImpl).toHaveBeenCalledWith(
      'https://teacher.example/api/teacher/logout',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
  })

  it('reads students, releases and live dashboard from the expected paths', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          items: [
            {
              id: 1,
              displayName: 'Ada',
              createdAt: '2026-05-25T12:00:00Z',
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          studentId: 1,
          studentName: 'Ada',
          items: [
            {
              assignmentId: 7,
              assignmentTitle: '第一期发布单',
              assignmentStatus: 'published',
              currentTarget: '让 Cat 角色移动起来',
              stepSummary: '已经接上方向键事件',
              hintText: '先把事件积木连起来',
              reportedAt: '2026-05-25T12:09:00Z',
              hintCreatedAt: '2026-05-25T12:11:00Z',
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          items: [
            {
              id: 7,
              title: '第一期发布单',
              status: 'published',
              studentCount: 3,
              updatedAt: '2026-05-25T12:10:00Z',
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          assignmentId: 7,
          assignmentTitle: '第一期发布单',
          updatedAt: '2026-05-25T12:12:00Z',
          students: [
            {
              studentId: 1,
              studentName: 'Ada',
              status: 'active',
              currentTarget: '让 Cat 角色移动起来',
              stepSummary: '已经接上方向键事件',
              lastHintText: '先把事件积木连起来',
              lastReportedAt: '',
              lastHintAt: '2026-05-25T12:11:00Z',
            },
          ],
        }),
      )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    })

    await expect(api.listStudents()).resolves.toEqual([
      {
        id: '1',
        username: '',
        name: 'Ada',
        className: '未分组',
        progress: 0,
        status: 'active',
        currentTarget: '让 Cat 角色移动起来',
        stepSummary: '已经接上方向键事件',
        latestAiHint: '先把事件积木连起来',
        updatedAt: '2026-05-25T12:11:00Z',
        createdAt: '2026-05-25T12:00:00Z',
      },
    ])
    await expect(api.listReleases()).resolves.toEqual([
      {
        id: '7',
        title: '第一期发布单',
        goal: '',
        description: '',
        className: '未分组',
        status: 'published',
        analysisStatus: 'pending',
        studentCount: 3,
        updatedAt: '2026-05-25T12:10:00Z',
      },
    ])
    await expect(api.getLiveDashboard('7')).resolves.toEqual({
      releaseId: '7',
      releaseTitle: '第一期发布单',
      updatedAt: '2026-05-25T12:12:00Z',
      students: [
        {
          id: '1',
          name: 'Ada',
          progress: 0,
          status: 'active',
          currentTarget: '让 Cat 角色移动起来',
          stepSummary: '已经接上方向键事件',
          latestAiHint: '先把事件积木连起来',
          updatedAt: '2026-05-25T12:11:00Z',
        },
      ],
    })

    expect(fetchImpl).toHaveBeenNthCalledWith(
      1,
      'https://teacher.example/api/teacher/students',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      2,
      'https://teacher.example/api/teacher/dashboard/students/1/history',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      3,
      'https://teacher.example/api/teacher/assignments',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      4,
      'https://teacher.example/api/teacher/dashboard/assignments/7/live',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
  })

  it('keeps the student list usable when one history request fails', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          items: [
            {
              id: 1,
              displayName: 'Ada',
              createdAt: '2026-05-25T12:00:00Z',
            },
          ],
        }),
      )
      .mockRejectedValueOnce(new Error('history unavailable'))
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    })

    await expect(api.listStudents()).resolves.toEqual([
      {
        id: '1',
        username: '',
        name: 'Ada',
        className: '未分组',
        progress: 0,
        status: '',
        currentTarget: '',
        stepSummary: '',
        latestAiHint: '等待学生请求提示',
        updatedAt: '2026-05-25T12:00:00Z',
        createdAt: '2026-05-25T12:00:00Z',
      },
    ])
  })

  it('calls unauthorized handler when a protected request returns 401', async () => {
    const onUnauthorized = vi.fn()
    const fetchImpl = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({
        message: 'missing or invalid bearer token',
      }),
    })
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'expired-token',
      onUnauthorized,
    })

    await expect(api.listStudents()).rejects.toThrow('missing or invalid bearer token')
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })

  it('does not swallow a 401 from student history requests', async () => {
    const onUnauthorized = vi.fn()
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          items: [
            {
              id: 1,
              displayName: 'Ada',
              createdAt: '2026-05-25T12:00:00Z',
            },
          ],
        }),
      )
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          message: 'missing or invalid bearer token',
        }),
      })
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'expired-token',
      onUnauthorized,
    })

    await expect(api.listStudents()).rejects.toThrow('missing or invalid bearer token')
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })

  it('creates and mutates teacher students from teacher endpoints', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          created: [
            {
              id: 5,
              username: 'student-1',
              displayName: '小蓝',
              status: 'active',
              createdAt: '2026-06-14T12:05:00Z',
            },
          ],
          conflicts: [],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 5,
          username: 'student-1',
          displayName: '小蓝',
          status: 'active',
          createdAt: '2026-06-14T12:05:00Z',
        }),
      )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    }) as ReturnType<typeof createFetchTeacherApiClient> & {
      createStudent?: (input: {
        username: string
        displayName: string
        initialPassword: string
      }) => Promise<unknown>
      resetStudentPassword?: (studentId: string, newPassword: string) => Promise<unknown>
    }

    expect(api.createStudent).toBeTypeOf('function')
    expect(api.resetStudentPassword).toBeTypeOf('function')
    if (!api.createStudent || !api.resetStudentPassword) {
      return
    }

    await expect(
      api.createStudent({
        username: 'student-1',
        displayName: '小蓝',
        initialPassword: 'abc12345',
      }),
    ).resolves.toEqual(
      expect.objectContaining({
        id: '5',
        username: 'student-1',
        name: '小蓝',
      }),
    )

    await expect(api.resetStudentPassword('5', 'updated123')).resolves.toEqual(
      expect.objectContaining({
        id: '5',
        username: 'student-1',
        name: '小蓝',
      }),
    )

    expect(fetchImpl).toHaveBeenNthCalledWith(
      1,
      'https://teacher.example/api/teacher/students',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'student-1',
          displayName: '小蓝',
          initialPassword: 'abc12345',
        }),
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      2,
      'https://teacher.example/api/teacher/students/5/reset-password',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          newPassword: 'updated123',
        }),
      }),
    )
  })

  it('uploads and manages teacher releases from teacher endpoints', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 9,
          title: '迷宫任务',
          goal: '让角色移动起来',
          description: '第一课任务',
          status: 'draft',
          analysisStatus: 'pending',
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 9,
          title: '迷宫任务',
          goal: '让角色移动起来',
          description: '第一课任务',
          status: 'draft',
          analysisStatus: 'ready',
          roleNames: ['Stage'],
          scriptCounts: { Stage: 1 },
          blockCounts: { event_whenflagclicked: 1 },
          categoryCounts: { event: 1 },
          broadcastMessages: ['开始'],
          variableNames: ['score'],
          listNames: ['targets'],
          extensions: ['pen'],
          teachingPoints: ['先搭好事件入口'],
          assignedStudents: [],
          updatedAt: '2026-06-14T12:06:00Z',
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          assignmentId: 9,
          analysisStatus: 'ready',
          analysisErrorMessage: '',
          roleNames: ['Stage'],
          scriptCounts: { Stage: 1 },
          blockCounts: { event_whenflagclicked: 1 },
          categoryCounts: { event: 1 },
          broadcastMessages: ['开始'],
          variableNames: ['score'],
          listNames: ['targets'],
          extensions: ['pen'],
          teachingPoints: ['先搭好事件入口'],
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          assignmentId: 9,
          studentIds: [1, 2],
          assignedCount: 2,
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 9,
          title: '迷宫任务',
          status: 'published',
          analysisStatus: 'ready',
        }),
      )
      .mockResolvedValueOnce(
        createFetchResponse({
          id: 9,
          title: '迷宫任务',
          status: 'archived',
          analysisStatus: 'ready',
        }),
      )
    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    }) as ReturnType<typeof createFetchTeacherApiClient> & {
      createRelease?: (input: {
        title: string
        goal: string
        description: string
        file: File
      }) => Promise<unknown>
      getReleaseDetail?: (releaseId: string) => Promise<unknown>
      getReleaseAnalysis?: (releaseId: string) => Promise<unknown>
      assignStudentsToRelease?: (releaseId: string, studentIds: string[]) => Promise<unknown>
      publishRelease?: (releaseId: string) => Promise<unknown>
      archiveRelease?: (releaseId: string) => Promise<unknown>
    }
    const file = new File(['fake-sb3'], 'maze.sb3', {
      type: 'application/zip',
    })

    expect(api.createRelease).toBeTypeOf('function')
    expect(api.getReleaseDetail).toBeTypeOf('function')
    expect(api.getReleaseAnalysis).toBeTypeOf('function')
    expect(api.assignStudentsToRelease).toBeTypeOf('function')
    expect(api.publishRelease).toBeTypeOf('function')
    expect(api.archiveRelease).toBeTypeOf('function')
    if (
      !api.createRelease
      || !api.getReleaseDetail
      || !api.getReleaseAnalysis
      || !api.assignStudentsToRelease
      || !api.publishRelease
      || !api.archiveRelease
    ) {
      return
    }

    await expect(
      api.createRelease({
        title: '迷宫任务',
        goal: '让角色移动起来',
        description: '第一课任务',
        file,
      }),
    ).resolves.toEqual(
      expect.objectContaining({
        id: '9',
        title: '迷宫任务',
        analysisStatus: 'pending',
      }),
    )
    await expect(api.getReleaseDetail('9')).resolves.toEqual(
      expect.objectContaining({
        id: '9',
        roleNames: ['Stage'],
      }),
    )
    await expect(api.getReleaseAnalysis('9')).resolves.toEqual(
      expect.objectContaining({
        assignmentId: '9',
        roleNames: ['Stage'],
      }),
    )
    await expect(api.assignStudentsToRelease('9', ['1', '2'])).resolves.toEqual({
      assignmentId: '9',
      studentIds: ['1', '2'],
      assignedCount: 2,
    })
    await expect(api.publishRelease('9')).resolves.toEqual(
      expect.objectContaining({
        id: '9',
        status: 'published',
      }),
    )
    await expect(api.archiveRelease('9')).resolves.toEqual(
      expect.objectContaining({
        id: '9',
        status: 'archived',
      }),
    )

    const createReleaseCall = fetchImpl.mock.calls[0]
    expect(createReleaseCall[0]).toBe('https://teacher.example/api/teacher/assignments')
    expect(createReleaseCall[1]).toEqual(
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(createReleaseCall[1]?.body).toBeInstanceOf(FormData)
    const releaseBody = createReleaseCall[1]?.body as FormData
    expect(releaseBody.get('title')).toBe('迷宫任务')
    expect(releaseBody.get('goal')).toBe('让角色移动起来')
    expect(releaseBody.get('description')).toBe('第一课任务')
    expect(releaseBody.get('sb3')).toBe(file)

    expect(fetchImpl).toHaveBeenNthCalledWith(
      2,
      'https://teacher.example/api/teacher/assignments/9',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      3,
      'https://teacher.example/api/teacher/assignments/9/analysis',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      4,
      'https://teacher.example/api/teacher/assignments/9/assign-students',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          studentIds: [1, 2],
        }),
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      5,
      'https://teacher.example/api/teacher/assignments/9/publish',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      6,
      'https://teacher.example/api/teacher/assignments/9/archive',
      expect.objectContaining({
        method: 'POST',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
  })
})
