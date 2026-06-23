import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ProjectDetailView from './ProjectDetailView.vue'
import { createMockTeacherApiClient } from '@/services/mockTeacherApi'
import { teacherApiKey } from '@/services/teacherApi'

const StubView = { template: '<div />' }

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/classes', component: StubView },
      { path: '/classes/:id/projects', component: StubView },
      { path: '/dashboard', component: StubView },
      { path: '/projects/:id', component: ProjectDetailView },
    ],
  })
}

describe('ProjectDetailView', () => {
  it('loads project overview and live student progress', async () => {
    const api = createMockTeacherApiClient()
    const router = createRouterForTest()
    router.push('/projects/rel-1')
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

    expect(wrapper.findAll('section.panel')).toHaveLength(2)
    expect(wrapper.text()).toContain('迷宫项目')
    expect(wrapper.text()).toContain('让角色按事件响应')
    expect(wrapper.text()).toContain('分析状态')
    expect(wrapper.text()).toContain('Ada')
    expect(wrapper.text()).toContain('学生当前进度与提示')
    expect(wrapper.text()).toContain('先把绿旗事件连起来')
  })

  it('shows a return link to the classroom projects workspace when context exists', async () => {
    const api = createMockTeacherApiClient()
    const router = createRouterForTest()
    router.push('/projects/rel-1?classroomId=class-1')
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

    const backLink = wrapper.get('a[href="/classes/class-1/projects"]')
    expect(backLink.text()).toContain('返回班级项目')
  })
})
