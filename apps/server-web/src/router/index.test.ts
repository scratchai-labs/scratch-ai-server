import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { createTeacherRouter } from './index'
import { useSessionStore } from '@/stores/session'

describe('createTeacherRouter', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    window.localStorage.clear()
  })

  it('redirects unauthenticated users to login and teachers to classes', async () => {
    const pinia = createPinia()
    const router = createTeacherRouter(pinia)

    await router.push('/classes')
    await router.isReady()

    expect(router.currentRoute.value.fullPath).toBe('/login?redirect=/classes')

    const sessionStore = useSessionStore(pinia)
    sessionStore.session = {
      token: 'teacher-token',
      teacherName: '王老师',
      role: 'teacher',
    }

    await router.push('/login')
    await router.isReady()

    expect(router.currentRoute.value.fullPath).toBe('/classes')
  })

  it('allows teachers into class workspace children and keeps admins out of teacher routes', async () => {
    const pinia = createPinia()
    const router = createTeacherRouter(pinia)
    const sessionStore = useSessionStore(pinia)

    sessionStore.session = {
      token: 'teacher-token',
      teacherName: '王老师',
      role: 'teacher',
    }

    await router.push('/classes/class-1/students')
    await router.isReady()

    expect(router.currentRoute.value.fullPath).toBe('/classes/class-1/students')

    sessionStore.session = {
      token: 'admin-token',
      teacherName: '系统管理员',
      role: 'admin',
    }

    await router.push('/classes/class-1/projects')

    expect(router.currentRoute.value.fullPath).toBe('/admin')
  })
})
