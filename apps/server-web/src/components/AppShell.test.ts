import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AppShell from './AppShell.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', component: { template: '<div>login</div>' } },
      { path: '/dashboard', component: { template: '<div>dashboard</div>' } },
      { path: '/students', component: { template: '<div>students</div>' } },
      { path: '/releases', component: { template: '<div>releases</div>' } },
      { path: '/admin', component: { template: '<div>admin overview</div>' } },
      { path: '/admin/teachers', component: { template: '<div>admin teachers</div>' } },
      { path: '/admin/students', component: { template: '<div>admin students</div>' } },
      { path: '/admin/audit-logs', component: { template: '<div>admin audit logs</div>' } },
    ],
  })
}

describe('AppShell', () => {
  it('renders a sidebar navigation shell', async () => {
    window.localStorage.setItem(
      'scratch-server-web.session',
      JSON.stringify({
        token: 'teacher-token',
        teacherName: '王老师',
        role: 'teacher',
      }),
    )

    const api = {
      login: vi.fn(),
      logout: vi.fn(),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/dashboard')
    await router.isReady()

    const wrapper = mount(AppShell, {
      props: {
        title: 'Dashboard',
      },
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    expect(wrapper.get('aside.shell__sidebar').text()).toContain('Scratch 教师后台')
    expect(wrapper.get('nav.shell__nav').text()).toContain('班级管理')
    expect(wrapper.get('nav.shell__nav').text()).toContain('实时总览')
    expect(wrapper.get('.shell__footer').text()).toContain('当前教师')
  })

  it('calls teacher logout api before clearing session', async () => {
    window.localStorage.setItem(
      'scratch-server-web.session',
      JSON.stringify({
        token: 'teacher-token',
        teacherName: '王老师',
        role: 'teacher',
      }),
    )

    const api = {
      login: vi.fn(),
      logout: vi.fn().mockResolvedValue(undefined),
      listStudents: vi.fn(),
      listReleases: vi.fn(),
      getLiveDashboard: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/dashboard')
    await router.isReady()

    const wrapper = mount(AppShell, {
      props: {
        title: 'Dashboard',
      },
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await wrapper.get('button[type="button"]').trigger('click')
    await flushPromises()

    expect(api.logout).toHaveBeenCalledTimes(1)
    expect(window.localStorage.getItem('scratch-server-web.session')).toBeNull()
    expect(router.currentRoute.value.fullPath).toBe('/login')
  })

  it('renders admin navigation for admin sessions', async () => {
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
      listTeachers: vi.fn(),
      createTeacher: vi.fn(),
      resetTeacherPassword: vi.fn(),
      enableTeacher: vi.fn(),
      disableTeacher: vi.fn(),
      getAdminOverview: vi.fn(),
      listManagedStudents: vi.fn(),
      resetManagedStudentPassword: vi.fn(),
      enableManagedStudent: vi.fn(),
      disableManagedStudent: vi.fn(),
    }
    const router = createRouterForTest()
    router.push('/admin')
    await router.isReady()

    const wrapper = mount(AppShell, {
      props: {
        title: 'Admin',
      },
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    expect(wrapper.get('nav.shell__nav').text()).toContain('后台总览')
    expect(wrapper.get('nav.shell__nav').text()).toContain('教师管理')
    expect(wrapper.get('nav.shell__nav').text()).toContain('学生管理')
    expect(wrapper.get('nav.shell__nav').text()).toContain('操作日志')
    expect(wrapper.get('nav.shell__nav').text()).not.toContain('实时总览')
  })
})
