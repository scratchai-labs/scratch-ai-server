import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClassesView from './ClassesView.vue'
import { createMockTeacherApiClient } from '@/services/mockTeacherApi'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/classes', component: ClassesView }],
  })
}

describe('ClassesView', () => {
  it('renders the canonical mock classrooms and navigates into a classroom', async () => {
    const api = createMockTeacherApiClient()
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

    expect(wrapper.text()).toContain('四年级一班')
    expect(wrapper.text()).toContain('2 名学生')
    expect(wrapper.text()).toContain('1 个项目')
    expect(wrapper.text()).toContain('四年级二班')

    await wrapper.get('a[href="/classes/class-1"]').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/classes/class-1')
  })
})
