import { describe, expect, it, vi } from 'vitest'
import { faker } from '@faker-js/faker'
import { createFetchTeacherApiClient } from './teacherApi'

function createFetchResponse(body: unknown) {
  return {
    ok: true,
    status: 200,
    json: async () => body,
  }
}

describe('teacherApi normalization with fake data', () => {
  it('normalizes large fake collections from backend payloads', async () => {
    faker.seed(20260526)

    const studentPayload = Array.from({ length: 18 }, () => ({
      id: faker.number.int({ min: 1, max: 9999 }),
      displayName: faker.person.fullName(),
      createdAt: faker.date.recent().toISOString(),
    }))
    const releasePayload = Array.from({ length: 9 }, () => ({
      id: faker.number.int({ min: 1, max: 9999 }),
      title: faker.company.buzzPhrase(),
      status: faker.helpers.arrayElement(['draft', 'published', 'archived'] as const),
      studentCount: faker.number.int({ min: 0, max: 40 }),
      updatedAt: faker.date.recent().toISOString(),
    }))
    const dashboardPayload = Array.from({ length: 7 }, () => ({
      studentId: faker.number.int({ min: 1, max: 9999 }),
      studentName: faker.person.fullName(),
      lastHintText: faker.lorem.sentence(),
      lastReportedAt: faker.date.recent().toISOString(),
    }))

    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(createFetchResponse({ items: studentPayload }))
      .mockResolvedValueOnce(createFetchResponse({ items: releasePayload }))
      .mockResolvedValueOnce(
        createFetchResponse({
          assignmentId: 88,
          assignmentTitle: 'Batch Fake Dashboard',
          updatedAt: faker.date.recent().toISOString(),
          students: dashboardPayload,
        }),
      )

    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    })

    const students = await api.listStudents()
    const releases = await api.listReleases()
    const dashboard = await api.getLiveDashboard('88')

    expect(students).toHaveLength(studentPayload.length)
    expect(students[0]).toMatchObject({
      id: String(studentPayload[0]?.id),
      name: String(studentPayload[0]?.displayName),
      className: '未分组',
    })

    expect(releases).toHaveLength(releasePayload.length)
    expect(releases[0]).toMatchObject({
      id: String(releasePayload[0]?.id),
      title: String(releasePayload[0]?.title),
      studentCount: Number(releasePayload[0]?.studentCount),
    })

    expect(dashboard.releaseId).toBe('88')
    expect(dashboard.students).toHaveLength(dashboardPayload.length)
    expect(dashboard.students[0]).toMatchObject({
      id: String(dashboardPayload[0]?.studentId),
      name: String(dashboardPayload[0]?.studentName),
      latestAiHint: String(dashboardPayload[0]?.lastHintText),
    })
  })
})
