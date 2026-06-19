import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ProjectDetailView from './ProjectDetailView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/projects/:id', component: ProjectDetailView }],
  })
}

afterEach(() => {
  window.localStorage.clear()
})

describe('ProjectDetailView', () => {
  it('loads project overview and live student progress', async () => {
    const api = {
      getReleaseDetail: vi.fn().mockResolvedValue({
        id: 'project-1',
        title: '迷宫项目',
        goal: '让角色按事件响应',
        description: '第一节课项目',
        status: 'published',
        analysisStatus: 'ready',
        roleNames: ['Stage', 'Cat'],
        scriptCounts: { Cat: 2 },
        blockCounts: { event_whenflagclicked: 1 },
        categoryCounts: { event: 1 },
        broadcastMessages: ['开始'],
        variableNames: ['score'],
        listNames: [],
        extensions: [],
        teachingPoints: ['先搭好事件入口'],
        assignedStudents: [],
        updatedAt: '2026-06-19T12:30:00Z',
      }),
      getReleaseAnalysis: vi.fn().mockResolvedValue({
        assignmentId: 'project-1',
        analysisStatus: 'ready',
        analysisErrorMessage: '',
        roleNames: ['Stage', 'Cat'],
        scriptCounts: { Cat: 2 },
        blockCounts: { event_whenflagclicked: 1 },
        categoryCounts: { event: 1 },
        broadcastMessages: ['开始'],
        variableNames: ['score'],
        listNames: [],
        extensions: [],
        teachingPoints: ['先搭好事件入口'],
      }),
      getLiveDashboard: vi.fn().mockResolvedValue({
        releaseId: 'project-1',
        releaseTitle: '迷宫项目',
        updatedAt: '2026-06-19T12:35:00Z',
        students: [
          {
            id: 'stu-1',
            name: '小明',
            progress: 0,
            status: 'active',
            currentTarget: '让角色从起点出发',
            stepSummary: '已经接上绿旗事件',
            latestAiHint: '先补一个移动 10 步',
            updatedAt: '2026-06-19T12:35:00Z',
          },
        ],
      }),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/projects/project-1')
    await router.isReady()

    const wrapper = mount(ProjectDetailView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(api.getReleaseDetail).toHaveBeenCalledWith('project-1')
    expect(api.getReleaseAnalysis).toHaveBeenCalledWith('project-1')
    expect(api.getLiveDashboard).toHaveBeenCalledWith('project-1')
    expect(wrapper.text()).toContain('迷宫项目')
    expect(wrapper.text()).toContain('让角色按事件响应')
    expect(wrapper.text()).toContain('先搭好事件入口')
    expect(wrapper.text()).toContain('小明')
    expect(wrapper.text()).toContain('让角色从起点出发')
    expect(wrapper.text()).toContain('已经接上绿旗事件')
    expect(wrapper.text()).toContain('先补一个移动 10 步')
  })
})
