import type { Router } from 'vue-router'
import { useSessionStore } from '@/stores/session'

export function createUnauthorizedHandler(
  sessionStore: ReturnType<typeof useSessionStore>,
  router: Router,
) {
  let redirecting = false

  return async function handleUnauthorized() {
    sessionStore.logout()
    if (redirecting || router.currentRoute.value.path === '/login') {
      return
    }

    redirecting = true
    try {
      await router.push('/login')
    } finally {
      redirecting = false
    }
  }
}
