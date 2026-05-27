import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import StudentsView from './StudentsView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/students', component: StudentsView }],
  })
}

describe('StudentsView', () => {
  it('renders the student list', async () => {
    const api = {
      listStudents: vi.fn().mockResolvedValue([
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
          progress: 0,
          status: 'active',
          currentTarget: '让 Alan 的角色先说一句话',
          stepSummary: '已经放上开始事件，但还没接外观积木',
          latestAiHint: '先补一个“说 2 秒”测试当前流程',
          updatedAt: '2026-05-07 09:24',
        },
      ]),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/students')
    await router.isReady()

    const wrapper = mount(StudentsView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Ada')
    expect(wrapper.text()).toContain('Alan')
    expect(wrapper.text()).toContain('补上广播消息后再测试一次')
    expect(wrapper.text()).toContain('先补一个“说 2 秒”测试当前流程')
    expect(wrapper.text()).toContain('已上报')
    expect(wrapper.text()).toContain('让 Alan 的角色先说一句话')
    expect(wrapper.text()).toContain('已经放上开始事件，但还没接外观积木')
  })
})
