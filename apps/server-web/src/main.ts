import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { createMockTeacherApiClient } from './services/mockTeacherApi'
import { resolveTeacherApiRuntime } from './services/runtimeEnv'
import { createUnauthorizedHandler } from './services/sessionRuntime'
import { createFetchTeacherApiClient, teacherApiKey } from './services/teacherApi'
import { createTeacherRouter } from './router'
import { useSessionStore } from './stores/session'
import './styles.css'

const app = createApp(App)
const pinia = createPinia()
const router = createTeacherRouter(pinia)
const sessionStore = useSessionStore(pinia)
const runtime = resolveTeacherApiRuntime(import.meta.env, import.meta.env.PROD)

const apiClient =
  runtime.mode === 'real'
    ? createFetchTeacherApiClient({
        baseUrl: runtime.baseUrl,
        getToken: () => sessionStore.token,
        onUnauthorized: createUnauthorizedHandler(sessionStore, router),
      })
    : createMockTeacherApiClient()

app.use(pinia)
app.use(router)
app.provide(teacherApiKey, apiClient)
app.mount('#app')
