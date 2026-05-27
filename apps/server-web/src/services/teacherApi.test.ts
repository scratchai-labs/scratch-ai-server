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
    })
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
        name: 'Ada',
        className: '未分组',
        progress: 0,
        latestAiHint: '等待学生请求提示',
        updatedAt: '2026-05-25T12:00:00Z',
      },
    ])
    await expect(api.listReleases()).resolves.toEqual([
      {
        id: '7',
        title: '第一期发布单',
        className: '未分组',
        status: 'published',
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
      'https://teacher.example/api/teacher/assignments',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
    expect(fetchImpl).toHaveBeenNthCalledWith(
      3,
      'https://teacher.example/api/teacher/dashboard/assignments/7/live',
      expect.objectContaining({
        method: 'GET',
        headers: {
          Authorization: 'Bearer teacher-token',
        },
      }),
    )
  })
})
