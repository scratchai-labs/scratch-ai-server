import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { faker } from '@faker-js/faker'
import { useTeacherDirectoryStore } from './teacherDirectory'
import type { TeacherApiClient, TeacherRelease, TeacherStudent } from '@/services/teacherApi'

describe('teacherDirectory store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    faker.seed(20260526)
  })

  it('tracks counts correctly for fake students and releases', async () => {
    const students = createFakeStudents(24)
    const releases = createFakeReleases(15)
    const store = useTeacherDirectoryStore()

    const api: TeacherApiClient = {
      login: async () => ({ token: '', teacherName: '' }),
      listStudents: async () => students,
      listReleases: async () => releases,
      getLiveDashboard: async () => ({
        releaseId: '',
        releaseTitle: '',
        updatedAt: '',
        students: [],
      }),
    }

    await store.loadStudents(api)
    await store.loadReleases(api)

    expect(store.studentCount).toBe(students.length)
    expect(store.releaseCount).toBe(releases.length)
    expect(store.publishedReleaseCount).toBe(
      releases.filter((release) => release.status === 'published').length,
    )
    expect(store.students[0]?.name).toBe(students[0]?.name)
    expect(store.releases.at(-1)?.title).toBe(releases.at(-1)?.title)
  })

  it('replaces a previous error after a successful fake-data reload', async () => {
    const store = useTeacherDirectoryStore()
    const failingApi: TeacherApiClient = {
      login: async () => ({ token: '', teacherName: '' }),
      listStudents: async () => {
        throw new Error('students unavailable')
      },
      listReleases: async () => [],
      getLiveDashboard: async () => ({
        releaseId: '',
        releaseTitle: '',
        updatedAt: '',
        students: [],
      }),
    }
    const successStudents = createFakeStudents(8)
    const successApi: TeacherApiClient = {
      ...failingApi,
      listStudents: async () => successStudents,
    }

    await store.loadStudents(failingApi)
    expect(store.studentsError).toContain('students unavailable')

    await store.loadStudents(successApi)
    expect(store.studentsError).toBeNull()
    expect(store.studentCount).toBe(successStudents.length)
  })
})

function createFakeStudents(count: number): TeacherStudent[] {
  return Array.from({ length: count }, (_, index) => ({
    id: faker.string.uuid(),
    name: faker.person.fullName(),
    className: `${faker.number.int({ min: 1, max: 6 })} 年级 ${faker.number.int({ min: 1, max: 4 })} 班`,
    progress: faker.number.int({ min: 0, max: 100 }),
    latestAiHint: faker.helpers.arrayElement([
      '先把广播消息接上',
      '把重复动作整理成三步',
      '先让角色移动，再调试边界',
    ]),
    updatedAt: faker.date.recent().toISOString(),
  }))
}

function createFakeReleases(count: number): TeacherRelease[] {
  const statuses: TeacherRelease['status'][] = ['draft', 'published', 'archived']

  return Array.from({ length: count }, (_, index) => ({
    id: `rel-${index + 1}`,
    title: faker.helpers.arrayElement([
      '迷宫挑战',
      '太空漫游',
      '角色对话',
      '画笔实验室',
    ]) + ` ${index + 1}`,
    className: `${faker.number.int({ min: 1, max: 6 })} 年级 ${faker.number.int({ min: 1, max: 4 })} 班`,
    status: faker.helpers.arrayElement(statuses),
    studentCount: faker.number.int({ min: 1, max: 40 }),
    updatedAt: faker.date.recent().toISOString(),
  }))
}
