import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminTeachersView from './AdminTeachersView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/admin/teachers', component: AdminTeachersView }],
  })
}

describe('AdminTeachersView', () => {
  it('renders managed teachers and allows creating and disabling a teacher', async () => {
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
      listTeachers: vi
        .fn()
        .mockResolvedValueOnce([
          {
            id: '1',
            username: 'admin',
            role: 'admin',
            status: 'active',
            createdAt: '2026-06-13T12:00:00Z',
          },
        ])
        .mockResolvedValueOnce([
          {
            id: '1',
            username: 'admin',
            role: 'admin',
            status: 'active',
            createdAt: '2026-06-13T12:00:00Z',
          },
          {
            id: '2',
            username: 'teacher-1',
            role: 'teacher',
            status: 'disabled',
            createdAt: '2026-06-13T12:01:00Z',
          },
        ]),
      createTeacher: vi.fn().mockResolvedValue({
        id: '2',
        username: 'teacher-1',
        role: 'teacher',
        status: 'active',
        createdAt: '2026-06-13T12:01:00Z',
      }),
      resetTeacherPassword: vi.fn().mockResolvedValue({
        id: '2',
        username: 'teacher-1',
        role: 'teacher',
        status: 'active',
        createdAt: '2026-06-13T12:01:00Z',
      }),
      disableTeacher: vi.fn().mockResolvedValue({
        id: '2',
        username: 'teacher-1',
        role: 'teacher',
        status: 'disabled',
        createdAt: '2026-06-13T12:01:00Z',
      }),
      enableTeacher: vi.fn().mockResolvedValue({
        id: '2',
        username: 'teacher-1',
        role: 'teacher',
        status: 'active',
        createdAt: '2026-06-13T12:01:00Z',
      }),
    }
    const router = createRouterForTest()
    router.push('/admin/teachers')
    await router.isReady()

    const wrapper = mount(AdminTeachersView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('admin')

    await wrapper.get('input[name="teacher-username"]').setValue('teacher-1')
    await wrapper.get('input[name="teacher-password"]').setValue('secret123')
    await wrapper.get('form[data-testid="create-teacher-form"]').trigger('submit.prevent')
    await flushPromises()

    expect(api.createTeacher).toHaveBeenCalledWith({
      username: 'teacher-1',
      initialPassword: 'secret123',
    })

    await wrapper.get('button[data-testid="teacher-disable-2"]').trigger('click')
    await flushPromises()

    expect(api.disableTeacher).toHaveBeenCalledWith('2')
    expect(wrapper.text()).toContain('teacher-1')
    expect(wrapper.text()).toContain('disabled')
  })

  it('allows promoting a teacher to admin from the teacher list', async () => {
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
      listTeachers: vi.fn().mockResolvedValue([
        {
          id: '1',
          username: 'admin',
          role: 'admin',
          status: 'active',
          createdAt: '2026-06-13T12:00:00Z',
        },
        {
          id: '2',
          username: 'teacher-1',
          role: 'teacher',
          status: 'active',
          createdAt: '2026-06-13T12:01:00Z',
        },
      ]),
      createTeacher: vi.fn(),
      resetTeacherPassword: vi.fn(),
      disableTeacher: vi.fn(),
      enableTeacher: vi.fn(),
      changeTeacherRole: vi.fn().mockResolvedValue({
        id: '2',
        username: 'teacher-1',
        role: 'admin',
        status: 'active',
        createdAt: '2026-06-13T12:01:00Z',
      }),
    }
    const router = createRouterForTest()
    router.push('/admin/teachers')
    await router.isReady()

    const wrapper = mount(AdminTeachersView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('teacher-1')
    await wrapper.get('button[data-testid="teacher-role-2"]').trigger('click')
    await flushPromises()

    expect(api.changeTeacherRole).toHaveBeenCalledWith('2', 'admin')
    expect(wrapper.text()).toContain('teacher-1')
    expect(wrapper.text()).toContain('admin')
  })
})
