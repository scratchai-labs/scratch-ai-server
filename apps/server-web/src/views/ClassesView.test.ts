import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ClassesView from './ClassesView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/classes', component: ClassesView }],
  })
}

afterEach(() => {
  window.localStorage.clear()
})

describe('ClassesView', () => {
  it('loads classrooms and creates a new classroom', async () => {
    const api = {
      listClassrooms: vi.fn().mockResolvedValue([
        {
          id: 'class-1',
          name: '四年级一班',
          studentCount: 24,
          projectCount: 2,
          createdAt: '2026-06-19T12:00:00Z',
          updatedAt: '2026-06-19T12:30:00Z',
        },
      ]),
      createClassroom: vi.fn().mockResolvedValue({
        id: 'class-2',
        name: '五年级二班',
        studentCount: 0,
        projectCount: 0,
        createdAt: '2026-06-19T13:00:00Z',
        updatedAt: '2026-06-19T13:00:00Z',
      }),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/classes')
    await router.isReady()

    const wrapper = mount(ClassesView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(api.listClassrooms).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('四年级一班')
    expect(wrapper.text()).toContain('24 名学生')
    expect(wrapper.text()).toContain('2 个项目')

    await wrapper.get('input').setValue('五年级二班')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(api.createClassroom).toHaveBeenCalledWith({
      name: '五年级二班',
    })
    expect(wrapper.text()).toContain('五年级二班')
  })
})
