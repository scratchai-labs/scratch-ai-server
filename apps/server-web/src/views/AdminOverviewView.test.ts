import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminOverviewView from './AdminOverviewView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/admin', component: AdminOverviewView }],
  })
}

describe('AdminOverviewView', () => {
  it('renders admin overview metrics from the admin api', async () => {
    window.localStorage.setItem(
      'scratch-server-web.session',
      JSON.stringify({
        token: 'admin-token',
        teacherName: '系统管理员',
        role: 'admin',
      }),
    )

    const api = {
      login: vi.fn(),
      logout: vi.fn(),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
      getAdminOverview: vi.fn().mockResolvedValue({
        adminCount: 1,
        teacherCount: 3,
        activeTeacherCount: 2,
        disabledTeacherCount: 1,
        studentCount: 48,
        activeStudentCount: 45,
        disabledStudentCount: 3,
      }),
    }

    const router = createRouterForTest()
    router.push('/admin')
    await router.isReady()

    const wrapper = mount(AdminOverviewView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(api.getAdminOverview).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('section.panel')).toHaveLength(2)
    expect(wrapper.text()).toContain('账号规模总览')
    expect(wrapper.text()).toContain('后台总览')
    expect(wrapper.text()).toContain('教师账号')
    expect(wrapper.text()).toContain('48')
    expect(wrapper.text()).toContain('禁用学生')
  })
})
