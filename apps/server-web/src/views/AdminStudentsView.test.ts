import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminStudentsView from './AdminStudentsView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/admin/students', component: AdminStudentsView }],
  })
}

describe('AdminStudentsView', () => {
  it('renders managed students and allows disabling a student', async () => {
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
      listManagedStudents: vi.fn().mockResolvedValue([
        {
          id: '10',
          teacherId: '2',
          teacherUsername: 'teacher-1',
          username: 'student-1',
          displayName: '小蓝',
          status: 'active',
          createdAt: '2026-06-13T12:05:00Z',
        },
      ]),
      resetManagedStudentPassword: vi.fn().mockResolvedValue({
        id: '10',
        teacherId: '2',
        teacherUsername: 'teacher-1',
        username: 'student-1',
        displayName: '小蓝',
        status: 'active',
        createdAt: '2026-06-13T12:05:00Z',
      }),
      enableManagedStudent: vi.fn(),
      disableManagedStudent: vi.fn().mockResolvedValue({
        id: '10',
        teacherId: '2',
        teacherUsername: 'teacher-1',
        username: 'student-1',
        displayName: '小蓝',
        status: 'disabled',
        createdAt: '2026-06-13T12:05:00Z',
      }),
    }

    const router = createRouterForTest()
    router.push('/admin/students')
    await router.isReady()

    const wrapper = mount(AdminStudentsView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('student-1')
    expect(wrapper.text()).toContain('teacher-1')

    await wrapper.get('button[data-testid="student-disable-10"]').trigger('click')
    await flushPromises()

    expect(api.disableManagedStudent).toHaveBeenCalledWith('10')
    expect(wrapper.text()).toContain('disabled')
  })
})
