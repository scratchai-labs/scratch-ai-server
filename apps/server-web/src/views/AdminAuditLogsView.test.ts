import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminAuditLogsView from './AdminAuditLogsView.vue'
import { teacherApiKey } from '@/services/teacherApi'

function createRouterForTest() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/admin/audit-logs', component: AdminAuditLogsView }],
  })
}

describe('AdminAuditLogsView', () => {
  it('renders admin audit logs and filters them by action', async () => {
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
      listAdminAuditLogs: vi.fn().mockResolvedValue([
        {
          id: '11',
          actorUsername: 'admin',
          action: 'teacher.role_change',
          targetType: 'teacher',
          targetId: '2',
          targetUsername: 'teacher-1',
          before: { role: 'teacher' },
          after: { role: 'admin' },
          createdAt: '2026-06-14T12:00:00Z',
        },
        {
          id: '10',
          actorUsername: 'admin',
          action: 'student.disable',
          targetType: 'student',
          targetId: '5',
          targetUsername: 'student-1',
          before: { status: 'active' },
          after: { status: 'disabled' },
          createdAt: '2026-06-14T11:58:00Z',
        },
      ]),
    }

    const router = createRouterForTest()
    router.push('/admin/audit-logs')
    await router.isReady()

    const wrapper = mount(AdminAuditLogsView, {
      global: {
        plugins: [createPinia(), router],
        provide: {
          [teacherApiKey as symbol]: api,
        },
      },
    })

    await flushPromises()

    expect(api.listAdminAuditLogs).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('操作日志')
    expect(wrapper.text()).toContain('teacher.role_change')
    expect(wrapper.text()).toContain('student.disable')

    await wrapper.get('select[name="audit-action-filter"]').setValue('teacher.role_change')
    await flushPromises()

    expect(wrapper.text()).toContain('teacher.role_change')
    expect(wrapper.text()).not.toContain('student.disable')
  })
})
