import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import LiveReleaseView from './LiveReleaseView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/releases/:id/live', component: LiveReleaseView }],
  })
}

afterEach(() => {
  vi.useRealTimers()
})

describe('LiveReleaseView', () => {
  it('polls live dashboard updates and refreshes the latest progress', async () => {
    vi.useFakeTimers()
    const api = {
      getLiveDashboard: vi
        .fn()
        .mockResolvedValueOnce({
          releaseId: 'rel-1',
          releaseTitle: '第一期发布单',
          updatedAt: '2026-05-07 09:40',
          students: [
            {
              id: 'stu-1',
              name: 'Ada',
              progress: 42,
              status: 'active',
              currentTarget: '让 Cat 左右移动',
              stepSummary: '已经接上方向键事件',
              latestAiHint: '先把绿旗事件连起来',
              updatedAt: '2026-05-07 09:40',
            },
          ],
        })
        .mockResolvedValueOnce({
          releaseId: 'rel-1',
          releaseTitle: '第一期发布单',
          updatedAt: '2026-05-07 09:44',
          students: [
            {
              id: 'stu-1',
              name: 'Ada',
              progress: 68,
              status: 'active',
              currentTarget: '让 Cat 碰到边缘后转向',
              stepSummary: '已经加上碰到边缘反弹',
              latestAiHint: '现在补上角色切换逻辑',
              updatedAt: '2026-05-07 09:44',
            },
          ],
        }),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/releases/rel-1/live')
    await router.isReady()

    const wrapper = mount(LiveReleaseView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Ada')
    expect(wrapper.text()).toContain('42')
    expect(wrapper.text()).toContain('让 Cat 左右移动')
    expect(wrapper.text()).toContain('已经接上方向键事件')
    expect(wrapper.text()).toContain('先把绿旗事件连起来')
    expect(wrapper.text()).toContain('2026-05-07 09:40')

    await vi.advanceTimersByTimeAsync(4000)
    await flushPromises()

    expect(wrapper.text()).toContain('68')
    expect(wrapper.text()).toContain('让 Cat 碰到边缘后转向')
    expect(wrapper.text()).toContain('已经加上碰到边缘反弹')
    expect(wrapper.text()).toContain('现在补上角色切换逻辑')
    expect(wrapper.text()).toContain('2026-05-07 09:44')
  })

  it('shows status and summary details when the real API has no numeric progress', async () => {
    vi.useFakeTimers()
    const api = {
      getLiveDashboard: vi.fn().mockResolvedValue({
        releaseId: 'rel-1',
        releaseTitle: '第一期发布单',
        updatedAt: '2026-05-07 09:44',
        students: [
          {
            id: 'stu-1',
            name: 'Ada',
            progress: 0,
            status: 'active',
            currentTarget: '让 Cat 角色移动起来',
            stepSummary: '已经接上方向键事件',
            latestAiHint: '现在补上角色切换逻辑',
            updatedAt: '2026-05-07 09:44',
          },
        ],
      }),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/releases/rel-1/live')
    await router.isReady()

    const wrapper = mount(LiveReleaseView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('已上报')
    expect(wrapper.text()).toContain('让 Cat 角色移动起来')
    expect(wrapper.text()).toContain('已经接上方向键事件')
  })
})
