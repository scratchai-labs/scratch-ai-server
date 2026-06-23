import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClassDetailView from './ClassDetailView.vue'
import ClassOverviewView from './ClassOverviewView.vue'
import ClassProjectsView from './ClassProjectsView.vue'
import ClassStudentsView from './ClassStudentsView.vue'
import { createMockTeacherApiClient } from '@/services/mockTeacherApi'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/classes/:id',
        component: ClassDetailView,
        children: [
          {
            path: '',
            component: ClassOverviewView,
          },
          {
            path: 'students',
            component: ClassStudentsView,
          },
          {
            path: 'projects',
            component: ClassProjectsView,
          },
        ],
      },
    ],
  })
}

describe('ClassDetailView', () => {
  it('renders the overview workspace by default', async () => {
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

    expect(router.currentRoute.value.fullPath).toBe('/classes/class-1')
    expect(wrapper.text()).toContain('四年级一班')
    expect(wrapper.text()).toContain('班级摘要')
    expect(wrapper.text()).toContain('建议操作顺序')
    expect(wrapper.text()).toContain('概览')
    expect(wrapper.text()).toContain('学生')
    expect(wrapper.text()).toContain('项目')
    const linkTexts = wrapper.findAll('a').map((node) => node.text())
    expect(linkTexts).not.toContain('进入学生页')
    expect(linkTexts).not.toContain('进入项目页')
  })

  it('creates students and batch imports from the students subpage', async () => {
    const api = createMockTeacherApiClient()
    const router = createRouterForTest()
    router.push('/classes/class-1/students')
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

    expect(wrapper.text()).toContain('批量导入学生')
    expect(wrapper.get('textarea.input').exists()).toBe(true)

    await wrapper.get('input[placeholder="student-01"]').setValue('student-03')
    await wrapper.get('input[placeholder="小明"]').setValue('小明')
    await wrapper.get('input[type="password"][placeholder="abc12345"]').setValue('abc12345')
    await wrapper.findAll('form')[0]?.trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('student-03')
    expect(wrapper.text()).toContain('已创建学生账号 student-03')

    await wrapper.findAll('input[type="password"]')[1]?.setValue('abc12345')
    await wrapper.get('textarea').setValue('姓名\t账号\t密码\n小红\tstudent-04\t')
    await wrapper.findAll('form')[1]?.trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('student-04')
    expect(wrapper.text()).toContain('小红')
    expect(wrapper.text()).toContain('已批量创建 1 名学生')
  })

  it('creates classroom projects from the projects subpage', async () => {
    const file = new File(['fake-sb3'], 'maze.sb3', { type: 'application/zip' })
    const api = createMockTeacherApiClient()
    const router = createRouterForTest()
    router.push('/classes/class-1/projects')
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

    const fileInput = wrapper.get('input[type="file"]')
    Object.defineProperty(fileInput.element, 'files', {
      value: [file],
    })
    await fileInput.trigger('change')
    await wrapper.get('input[placeholder="迷宫项目"]').setValue('追逐项目')
    await wrapper.get('input[placeholder="让角色按事件响应"]').setValue('补齐广播与重复执行')
    await wrapper.get('input[placeholder="第一节课项目"]').setValue('第二节课项目')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('追逐项目')
    expect(wrapper.text()).toContain('已创建项目 追逐项目')
    expect(wrapper.text()).toContain('pending')
  })
})
