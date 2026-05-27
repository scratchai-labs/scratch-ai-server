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
    const studentHistoryById = new Map(
      studentPayload.map((student, index) => [
        String(student.id),
        {
          studentId: student.id,
          studentName: student.displayName,
          items: [
            {
              assignmentId: 100 + index,
              assignmentTitle: `Fake Assignment ${index + 1}`,
              assignmentStatus: index % 2 === 0 ? 'published' : 'draft',
              currentTarget: `Fake Target ${index + 1}`,
              stepSummary: `Fake Step ${index + 1}`,
              hintText: `Fake Hint ${index + 1}`,
              reportedAt: new Date(Date.UTC(2026, 4, 27, 12, 0, index)).toISOString(),
              hintCreatedAt: new Date(Date.UTC(2026, 4, 27, 12, 1, index)).toISOString(),
            },
          ],
        },
      ]),
    )

    const fetchImpl = vi
      .fn()
      .mockImplementation(async (url: string) => {
        if (url.endsWith('/api/teacher/students')) {
          return createFetchResponse({ items: studentPayload })
        }

        if (url.endsWith('/api/teacher/assignments')) {
          return createFetchResponse({ items: releasePayload })
        }

        if (url.endsWith('/api/teacher/dashboard/assignments/88/live')) {
          return createFetchResponse({
            assignmentId: 88,
            assignmentTitle: 'Batch Fake Dashboard',
            updatedAt: faker.date.recent().toISOString(),
            students: dashboardPayload,
          })
        }

        const historyMatch = url.match(/\/api\/teacher\/dashboard\/students\/([^/]+)\/history$/)
        if (historyMatch) {
          const historyPayload = studentHistoryById.get(historyMatch[1] ?? '')
          return createFetchResponse(historyPayload ?? { items: [] })
        }

        throw new Error(`unexpected fetch url: ${url}`)
      })

    const api = createFetchTeacherApiClient({
      baseUrl: 'https://teacher.example',
      fetchImpl,
      getToken: () => 'teacher-token',
    })

    const students = await api.listStudents()
    const releases = await api.listReleases()
    const dashboard = await api.getLiveDashboard('88')

    expect(students).toHaveLength(studentPayload.length)

    const matchedStudent = students.find((student) => student.id === String(studentPayload[0]?.id))
    expect(matchedStudent).toMatchObject({
      id: String(studentPayload[0]?.id),
      name: String(studentPayload[0]?.displayName),
      className: '未分组',
      currentTarget: 'Fake Target 1',
      stepSummary: 'Fake Step 1',
      latestAiHint: 'Fake Hint 1',
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
