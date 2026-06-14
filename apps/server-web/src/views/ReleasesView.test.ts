import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ReleasesView from './ReleasesView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/releases', component: ReleasesView }],
  })
}

describe('ReleasesView', () => {
  it('renders the release list', async () => {
    const api = {
      listReleases: vi.fn().mockResolvedValue([
        {
          id: 'rel-1',
          title: '第一期发布单',
          className: '四年级一班',
          status: 'published',
          studentCount: 24,
          updatedAt: '2026-05-07 09:10',
        },
        {
          id: 'rel-2',
          title: '第二期发布单',
          className: '四年级二班',
          status: 'draft',
          studentCount: 18,
          updatedAt: '2026-05-07 09:30',
        },
      ]),
      listStudents: vi.fn(),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/releases')
    await router.isReady()

    const wrapper = mount(ReleasesView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('第一期发布单')
    expect(wrapper.text()).toContain('第二期发布单')
    expect(wrapper.text()).toContain('24')
    expect(wrapper.text()).toContain('18')
  })

  it('uploads an sb3 release and shows release detail with analysis', async () => {
    const file = new File(['fake-sb3'], 'maze.sb3', {
      type: 'application/zip',
    })
    const api = {
      listReleases: vi
        .fn()
        .mockResolvedValueOnce([
          {
            id: 'rel-1',
            title: '第一期发布单',
            className: '未分组',
            status: 'draft',
            analysisStatus: 'ready',
            studentCount: 1,
            updatedAt: '2026-05-07 09:10',
          },
        ])
        .mockResolvedValueOnce([
          {
            id: 'rel-1',
            title: '第一期发布单',
            className: '未分组',
            status: 'draft',
            analysisStatus: 'ready',
            studentCount: 1,
            updatedAt: '2026-05-07 09:10',
          },
          {
            id: 'rel-2',
            title: '迷宫任务',
            className: '未分组',
            status: 'draft',
            analysisStatus: 'pending',
            studentCount: 0,
            updatedAt: '2026-05-07 09:30',
          },
        ]),
      createRelease: vi.fn().mockResolvedValue({
        id: 'rel-2',
        title: '迷宫任务',
        goal: '让角色移动起来',
        description: '第一课任务',
        status: 'draft',
        analysisStatus: 'pending',
        updatedAt: '2026-05-07 09:30',
      }),
      getReleaseDetail: vi.fn().mockResolvedValue({
        id: 'rel-2',
        title: '迷宫任务',
        goal: '让角色移动起来',
        description: '第一课任务',
        status: 'draft',
        analysisStatus: 'ready',
        roleNames: ['Stage', 'Cat'],
        scriptCounts: { Cat: 2 },
        blockCounts: { event_whenflagclicked: 1 },
        categoryCounts: { event: 1 },
        broadcastMessages: ['开始'],
        variableNames: ['score'],
        listNames: ['targets'],
        extensions: ['pen'],
        teachingPoints: ['先搭好事件入口'],
        assignedStudents: [],
        updatedAt: '2026-05-07 09:31',
      }),
      getReleaseAnalysis: vi.fn().mockResolvedValue({
        assignmentId: 'rel-2',
        analysisStatus: 'ready',
        analysisErrorMessage: '',
        roleNames: ['Stage', 'Cat'],
        scriptCounts: { Cat: 2 },
        blockCounts: { event_whenflagclicked: 1 },
        categoryCounts: { event: 1 },
        broadcastMessages: ['开始'],
        variableNames: ['score'],
        listNames: ['targets'],
        extensions: ['pen'],
        teachingPoints: ['先搭好事件入口'],
      }),
      listStudents: vi.fn().mockResolvedValue([]),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/releases')
    await router.isReady()

    const wrapper = mount(ReleasesView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    await wrapper.get('input[name="release-title"]').setValue('迷宫任务')
    await wrapper.get('input[name="release-goal"]').setValue('让角色移动起来')
    await wrapper.get('textarea[name="release-description"]').setValue('第一课任务')
    const fileInput = wrapper.get('input[name="release-sb3"]')
    Object.defineProperty(fileInput.element, 'files', {
      value: [file],
    })
    await fileInput.trigger('change')
    await wrapper.get('[data-testid="create-release-form"]').trigger('submit')
    await flushPromises()

    expect(api.createRelease).toHaveBeenCalledWith({
      title: '迷宫任务',
      goal: '让角色移动起来',
      description: '第一课任务',
      file,
    })
    expect(api.getReleaseDetail).toHaveBeenCalledWith('rel-2')
    expect(api.getReleaseAnalysis).toHaveBeenCalledWith('rel-2')
    expect(wrapper.text()).toContain('已上传发布单 迷宫任务')
    expect(wrapper.text()).toContain('先搭好事件入口')
    expect(wrapper.text()).toContain('Stage')
  })

  it('assigns students and changes release status from the detail panel', async () => {
    const api = {
      listReleases: vi.fn().mockResolvedValue([
        {
          id: 'rel-1',
          title: '第一期发布单',
          className: '未分组',
          status: 'draft',
          analysisStatus: 'ready',
          studentCount: 1,
          updatedAt: '2026-05-07 09:10',
        },
      ]),
      getReleaseDetail: vi
        .fn()
        .mockResolvedValue({
          id: 'rel-1',
          title: '第一期发布单',
          goal: '让角色移动起来',
          description: '第一课任务',
          status: 'draft',
          analysisStatus: 'ready',
          roleNames: ['Stage'],
          scriptCounts: { Stage: 1 },
          blockCounts: { event_whenflagclicked: 1 },
          categoryCounts: { event: 1 },
          broadcastMessages: [],
          variableNames: [],
          listNames: [],
          extensions: [],
          teachingPoints: ['先搭好事件入口'],
          assignedStudents: [
            {
              id: 'stu-1',
              username: 'student-01',
              displayName: 'Ada',
              status: 'active',
            },
          ],
          updatedAt: '2026-05-07 09:10',
        }),
      getReleaseAnalysis: vi.fn().mockResolvedValue({
        assignmentId: 'rel-1',
        analysisStatus: 'ready',
        analysisErrorMessage: '',
        roleNames: ['Stage'],
        scriptCounts: { Stage: 1 },
        blockCounts: { event_whenflagclicked: 1 },
        categoryCounts: { event: 1 },
        broadcastMessages: [],
        variableNames: [],
        listNames: [],
        extensions: [],
        teachingPoints: ['先搭好事件入口'],
      }),
      listStudents: vi.fn().mockResolvedValue([
        {
          id: 'stu-1',
          username: 'student-01',
          name: 'Ada',
          className: '未分组',
          progress: 0,
          status: 'assigned',
          latestAiHint: '等待学生请求提示',
          updatedAt: '2026-05-07 09:20',
        },
        {
          id: 'stu-2',
          username: 'student-02',
          name: 'Mia',
          className: '未分组',
          progress: 0,
          status: '',
          latestAiHint: '等待学生请求提示',
          updatedAt: '2026-05-07 09:24',
        },
      ]),
      assignStudentsToRelease: vi.fn().mockResolvedValue({
        assignmentId: 'rel-1',
        studentIds: ['stu-1', 'stu-2'],
        assignedCount: 2,
      }),
      publishRelease: vi.fn().mockResolvedValue({
        id: 'rel-1',
        title: '第一期发布单',
        status: 'published',
        analysisStatus: 'ready',
      }),
      archiveRelease: vi.fn().mockResolvedValue({
        id: 'rel-1',
        title: '第一期发布单',
        status: 'archived',
        analysisStatus: 'ready',
      }),
      createRelease: vi.fn(),
      getLiveDashboard: vi.fn(),
      login: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/releases')
    await router.isReady()

    const wrapper = mount(ReleasesView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    await wrapper.get('[data-testid="view-release-rel-1"]').trigger('click')
    await flushPromises()

    await wrapper.get('input[name="assign-student-stu-2"]').setValue(true)
    await wrapper.get('[data-testid="assign-release-rel-1"]').trigger('click')
    await flushPromises()

    expect(api.assignStudentsToRelease).toHaveBeenCalledWith('rel-1', ['stu-1', 'stu-2'])

    await wrapper.get('[data-testid="publish-release-rel-1"]').trigger('click')
    await flushPromises()
    expect(api.publishRelease).toHaveBeenCalledWith('rel-1')

    await wrapper.get('[data-testid="archive-release-rel-1"]').trigger('click')
    await flushPromises()
    expect(api.archiveRelease).toHaveBeenCalledWith('rel-1')
  })
})
