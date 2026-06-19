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
  it('loads classrooms and creates a new classroom', async () => {
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
    expect(wrapper.findAll('a[href^="/classes/"]').length).toBeGreaterThan(0)

    await wrapper.get('input').setValue('五年级三班')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('五年级三班')
    expect(wrapper.text()).toContain('创建时间')
  })
})
