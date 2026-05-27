import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import DashboardView from './DashboardView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/dashboard', component: DashboardView }],
  })
}

describe('DashboardView', () => {
  it('shows real latest student status instead of a fake percent placeholder', async () => {
    const api = {
      listStudents: vi.fn().mockResolvedValue([
        {
          id: 'stu-1',
          name: 'Ada',
          className: '四年级一班',
          progress: 0,
          status: 'active',
          currentTarget: '让 Ada 的角色先移动起来',
          stepSummary: '已经接上方向键事件',
          latestAiHint: '先把移动积木接成一条完整脚本',
          updatedAt: '2026-05-07 09:24',
        },
        {
          id: 'stu-2',
          name: 'Alan',
          className: '四年级二班',
          progress: 0,
          status: 'assigned',
          currentTarget: '',
          stepSummary: '',
          latestAiHint: '等待学生请求提示',
          updatedAt: '2026-05-07 09:10',
        },
      ]),
      listReleases: vi.fn().mockResolvedValue([
        {
          id: 'rel-1',
          title: '第一期发布单',
          className: '四年级一班',
          status: 'published',
          studentCount: 2,
          updatedAt: '2026-05-07 09:20',
        },
      ]),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    window.localStorage.setItem(
      'scratch-server-web.session',
      JSON.stringify({
        token: 'teacher-token',
        teacherName: '王老师',
      }),
    )
    const pinia = createPinia()

    const router = createRouterForTest()
    router.push('/dashboard')
    await router.isReady()

    const wrapper = mount(DashboardView, {
      global: {
        plugins: [pinia, router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('已上报学生')
    expect(wrapper.text()).toContain('1 / 2')
    expect(wrapper.text()).toContain('已上报')
    expect(wrapper.text()).toContain('让 Ada 的角色先移动起来')
    expect(wrapper.text()).toContain('已经接上方向键事件')
    expect(wrapper.text()).not.toContain('0%')
  })
})
