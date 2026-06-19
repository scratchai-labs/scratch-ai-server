import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClassDetailView from './ClassDetailView.vue'
import { createMockTeacherApiClient } from '@/services/mockTeacherApi'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/classes/:id', component: ClassDetailView }],
  })
}

describe('ClassDetailView', () => {
  it('loads classroom students and projects', async () => {
    const api = createMockTeacherApiClient()
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

    expect(wrapper.text()).toContain('四年级一班')
    expect(wrapper.text()).toContain('Ada')
    expect(wrapper.text()).toContain('Mia')
    expect(wrapper.text()).toContain('迷宫项目')
    expect(wrapper.text()).toContain('让角色按事件响应')
    expect(wrapper.text()).toContain('查看项目详情')
  })

  it('creates classroom students and projects from the detail page', async () => {
    const file = new File(['fake-sb3'], 'maze.sb3', { type: 'application/zip' })
    const api = createMockTeacherApiClient()
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

    await wrapper.get('input[placeholder="student-01"]').setValue('student-03')
    await wrapper.get('input[placeholder="小明"]').setValue('小明')
    await wrapper.get('input[type="password"][placeholder="abc12345"]').setValue('abc12345')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('student-03')
    expect(wrapper.text()).toContain('小明')

    await wrapper.findAll('input[type="password"]')[1]?.setValue('abc12345')
    await wrapper.get('textarea').setValue('姓名\t账号\t密码\n小红\tstudent-04\t')
    await wrapper.findAll('form')[1]?.trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('student-04')
    expect(wrapper.text()).toContain('小红')

    const fileInput = wrapper.get('input[type="file"]')
    Object.defineProperty(fileInput.element, 'files', {
      value: [file],
    })
    await fileInput.trigger('change')
    await wrapper.get('input[placeholder="迷宫项目"]').setValue('追逐项目')
    await wrapper.get('input[placeholder="让角色按事件响应"]').setValue('补齐广播与重复执行')
    await wrapper.get('input[placeholder="第一节课项目"]').setValue('第二节课项目')
    await wrapper.findAll('form')[2]?.trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('追逐项目')
    expect(wrapper.text()).toContain('分析状态：pending')
  })
})
