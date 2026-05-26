import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import { createMockTeacherApiClient } from './services/mockTeacherApi'
import { createFetchTeacherApiClient, teacherApiKey } from './services/teacherApi'
import { createTeacherRouter } from './router'
import { useSessionStore } from './stores/session'
import './styles.css'

const app = createApp(App)
const pinia = createPinia()
const router = createTeacherRouter(pinia)

const apiClient =
  import.meta.env.VITE_SERVER_WEB_API_MODE === 'real'
    ? createFetchTeacherApiClient({
        baseUrl: import.meta.env.VITE_SERVER_WEB_API_BASE_URL ?? '',
        getToken: () => useSessionStore(pinia).token,
      })
    : createMockTeacherApiClient()

app.use(pinia)
app.use(router)
app.provide(teacherApiKey, apiClient)
app.mount('#app')
