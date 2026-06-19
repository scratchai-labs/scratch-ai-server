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
})
