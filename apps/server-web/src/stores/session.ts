import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type {
  TeacherApiClient,
  TeacherLoginInput,
  TeacherSession,
} from '@/services/teacherApi'

const storageKey = 'scratch-server-web.session'

export const useSessionStore = defineStore('session', () => {
  const session = ref<TeacherSession | null>(loadSession())

  const isAuthenticated = computed(() => session.value !== null)
  const role = computed(() => session.value?.role ?? 'teacher')
  const isAdmin = computed(() => role.value === 'admin')
  const token = computed(() => session.value?.token ?? '')
  const teacherName = computed(() => session.value?.teacherName ?? '')
  const landingPath = computed(() => (isAdmin.value ? '/admin' : '/classes'))

  async function login(api: TeacherApiClient, input: TeacherLoginInput) {
    const nextSession = await api.login(input)
    session.value = nextSession
    saveSession(nextSession)
    return nextSession
  }

  function logout() {
    session.value = null
    clearSession()
  }

  return {
    session,
    isAuthenticated,
    role,
    isAdmin,
    token,
    teacherName,
    landingPath,
    login,
    logout,
  }
})

function loadSession(): TeacherSession | null {
  if (typeof window === 'undefined') {
    return null
  }

  const raw = window.localStorage.getItem(storageKey)

  if (!raw) {
    return null
  }

  try {
    const parsed = JSON.parse(raw) as Partial<TeacherSession>
    if (!parsed?.token || !parsed?.teacherName) {
      return null
    }

    return {
      token: parsed.token,
      teacherName: parsed.teacherName,
      role: parsed.role === 'admin' ? 'admin' : 'teacher',
    }
  } catch {
    return null
  }
}

function saveSession(session: TeacherSession) {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(storageKey, JSON.stringify(session))
}

function clearSession() {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.removeItem(storageKey)
}
