import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it } from 'vitest'
import { useSessionStore } from '@/stores/session'
import { createUnauthorizedHandler } from './sessionRuntime'

describe('createUnauthorizedHandler', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    window.localStorage.clear()
  })

  it('clears session and redirects to login', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/login', component: { template: '<div>login</div>' } },
        { path: '/dashboard', component: { template: '<div>dashboard</div>' } },
      ],
    })
    await router.push('/dashboard')
    await router.isReady()

    const sessionStore = useSessionStore()
    sessionStore.session = {
      token: 'teacher-token',
      teacherName: '王老师',
    }

    const handleUnauthorized = createUnauthorizedHandler(sessionStore, router)
    await handleUnauthorized()

    expect(sessionStore.session).toBeNull()
    expect(window.localStorage.getItem('scratch-server-web.session')).toBeNull()
    expect(router.currentRoute.value.fullPath).toBe('/login')
  })
})
