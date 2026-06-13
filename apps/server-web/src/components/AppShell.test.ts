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
    ],
  })
}

describe('AppShell', () => {
  it('renders a top header nav and footer layout', async () => {
    window.localStorage.setItem(
      'scratch-server-web.session',
      JSON.stringify({
        token: 'teacher-token',
        teacherName: '王老师',
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

    expect(wrapper.get('header.shell__header').text()).toContain('Scratch 教师后台')
    expect(wrapper.get('nav.shell__nav').text()).toContain('实时总览')
    expect(wrapper.get('footer.shell__site-footer').text()).toContain('课堂工具')
  })

  it('calls teacher logout api before clearing session', async () => {
    window.localStorage.setItem(
      'scratch-server-web.session',
      JSON.stringify({
        token: 'teacher-token',
        teacherName: '王老师',
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
})
