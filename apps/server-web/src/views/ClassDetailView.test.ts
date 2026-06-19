import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ClassDetailView from './ClassDetailView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/classes/:id', component: ClassDetailView }],
  })
}

afterEach(() => {
  window.localStorage.clear()
})

describe('ClassDetailView', () => {
  it('loads classroom students and projects', async () => {
    const api = {
      getClassroomDetail: vi.fn().mockResolvedValue({
        id: 'class-1',
        name: '四年级一班',
        studentCount: 2,
        projectCount: 1,
        createdAt: '2026-06-19T12:00:00Z',
        updatedAt: '2026-06-19T12:30:00Z',
      }),
      listClassroomStudents: vi.fn().mockResolvedValue([
        {
          id: 'stu-1',
          classroomId: 'class-1',
          username: 'student-01',
          name: '小明',
          className: '四年级一班',
          progress: 0,
          latestAiHint: '等待学生请求提示',
          updatedAt: '2026-06-19T12:20:00Z',
          createdAt: '2026-06-19T12:00:00Z',
        },
      ]),
      listClassroomProjects: vi.fn().mockResolvedValue([
        {
          id: 'project-1',
          classroomId: 'class-1',
          title: '迷宫项目',
          goal: '让角色按事件响应',
          description: '第一节课项目',
          className: '四年级一班',
          status: 'draft',
          analysisStatus: 'ready',
          studentCount: 1,
          updatedAt: '2026-06-19T12:30:00Z',
        },
      ]),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/classes/class-1')
    await router.isReady()

    const wrapper = mount(ClassDetailView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(api.getClassroomDetail).toHaveBeenCalledWith('class-1')
    expect(api.listClassroomStudents).toHaveBeenCalledWith('class-1')
    expect(api.listClassroomProjects).toHaveBeenCalledWith('class-1')
    expect(wrapper.text()).toContain('小明')
    expect(wrapper.text()).toContain('迷宫项目')
    expect(wrapper.text()).toContain('让角色按事件响应')
  })

  it('creates students by single form and batch import, then creates a project', async () => {
    const file = new File(['fake-sb3'], 'maze.sb3', { type: 'application/zip' })
    const api = {
      getClassroomDetail: vi.fn().mockResolvedValue({
        id: 'class-1',
        name: '四年级一班',
        studentCount: 0,
        projectCount: 0,
        createdAt: '2026-06-19T12:00:00Z',
        updatedAt: '2026-06-19T12:30:00Z',
      }),
      listClassroomStudents: vi.fn().mockResolvedValue([]),
      listClassroomProjects: vi.fn().mockResolvedValue([]),
      createClassroomStudent: vi.fn().mockResolvedValue({
        id: 'stu-1',
        classroomId: 'class-1',
        username: 'student-01',
        name: '小明',
        className: '四年级一班',
        progress: 0,
        latestAiHint: '等待学生请求提示',
        updatedAt: '2026-06-19T12:40:00Z',
        createdAt: '2026-06-19T12:40:00Z',
      }),
      batchCreateClassroomStudents: vi.fn().mockResolvedValue({
        created: [
          {
            id: 'stu-2',
            classroomId: 'class-1',
            username: 'student-02',
            name: '小红',
            className: '四年级一班',
            progress: 0,
            latestAiHint: '等待学生请求提示',
            updatedAt: '2026-06-19T12:41:00Z',
            createdAt: '2026-06-19T12:41:00Z',
          },
        ],
        conflicts: [],
      }),
      createClassroomProject: vi.fn().mockResolvedValue({
        id: 'project-2',
        title: '追逐项目',
        status: 'draft',
        analysisStatus: 'pending',
      }),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/classes/class-1')
    await router.isReady()

    const wrapper = mount(ClassDetailView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    const inputs = wrapper.findAll('input')
    await inputs[0]?.setValue('student-01')
    await inputs[1]?.setValue('小明')
    await inputs[2]?.setValue('abc12345')
    await wrapper.findAll('form')[0]?.trigger('submit')
    await flushPromises()

    expect(api.createClassroomStudent).toHaveBeenCalledWith('class-1', {
      username: 'student-01',
      displayName: '小明',
      initialPassword: 'abc12345',
    })
    expect(wrapper.text()).toContain('student-01')

    await inputs[3]?.setValue('abc12345')
    await wrapper.get('textarea').setValue('姓名\n小红')
    await wrapper.findAll('form')[1]?.trigger('submit')
    await flushPromises()

    expect(api.batchCreateClassroomStudents).toHaveBeenCalledWith('class-1', [
      {
        username: 'student-02',
        displayName: '小红',
        initialPassword: 'abc12345',
      },
    ])
    expect(wrapper.text()).toContain('student-02')

    await inputs[4]?.setValue('追逐项目')
    await inputs[5]?.setValue('让角色追逐')
    await inputs[6]?.setValue('第二节课项目')
    const fileInput = inputs[7]
    Object.defineProperty(fileInput.element, 'files', {
      value: [file],
    })
    await fileInput.trigger('change')
    await wrapper.findAll('form')[2]?.trigger('submit')
    await flushPromises()

    expect(api.createClassroomProject).toHaveBeenCalledWith('class-1', {
      title: '追逐项目',
      goal: '让角色追逐',
      description: '第二节课项目',
      file,
    })
    expect(wrapper.text()).toContain('追逐项目')
    expect(wrapper.text()).toContain('pending')
  })
})
