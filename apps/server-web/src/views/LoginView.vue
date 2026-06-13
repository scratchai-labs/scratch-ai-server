<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolveTeacherApiRuntime } from '@/services/runtimeEnv'
import { useSessionStore } from '@/stores/session'
import { toErrorMessage } from '@/stores/storeUtils'
import { useTeacherApiClient } from '@/services/teacherApi'

const router = useRouter()
const route = useRoute()
const sessionStore = useSessionStore()
const apiClient = useTeacherApiClient()

const form = reactive({
  username: '',
  password: '',
})

const submitting = ref(false)
const feedback = ref('')
const feedbackTone = ref<'error' | 'success' | ''>('')
const runtime = resolveTeacherApiRuntime(import.meta.env, import.meta.env.PROD)
const documentationHref =
  'https://github.com/scratchai-labs/scratch-ai-server/blob/main/docs/server-development.zh-CN.md'

const redirectTarget = computed(() => {
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/')) {
    return redirect
  }

  return sessionStore.landingPath
})

async function handleSubmit() {
  if (!form.username.trim() || !form.password.trim()) {
    feedback.value = '请输入账号和密码。'
    feedbackTone.value = 'error'
    return
  }

  submitting.value = true
  feedback.value = ''
  feedbackTone.value = ''

  try {
    await sessionStore.login(apiClient, {
      username: form.username.trim(),
      password: form.password,
    })

    feedback.value = '登录成功，正在进入看板。'
    feedbackTone.value = 'success'
    await router.push(redirectTarget.value)
  } catch (error) {
    feedback.value = toErrorMessage(error, '登录失败，请检查账号或密码。')
    feedbackTone.value = 'error'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="auth-layout">
    <header class="auth-header">
      <div class="auth-frame auth-header__bar">
        <div class="auth-brand">
          <div class="auth-brand__mark">S</div>
          <div class="auth-brand__copy">
            <strong class="auth-brand__title">Scratch 教师后台</strong>
            <p class="auth-brand__subtitle">AI 辅助课堂教学工具</p>
          </div>
        </div>

        <div class="auth-header__meta">
          <a
            class="auth-header__link"
            :href="documentationHref"
            target="_blank"
            rel="noreferrer"
          >
            开发说明
          </a>
        </div>
      </div>
    </header>

    <main class="auth-main">
      <div class="site-frame auth-main__frame">
        <section class="auth-card auth-card--solo">
          <div class="stack">
            <h2>登录教学后台</h2>
            <p class="auth-card__description">
              {{ runtime.showMockLoginHint ? '这里先接 mock client，支持教师或管理员演示登录；切到真实环境后统一走 `/api/teacher/login`。' : '当前会直接调用真实 `/api/teacher/login`，并按账号角色进入对应后台。' }}
            </p>
          </div>

          <form class="form-grid" @submit.prevent="handleSubmit">
            <label class="field">
              <span>账号</span>
              <input
                v-model="form.username"
                class="input"
                name="username"
                autocomplete="username"
                placeholder="teacher"
              />
            </label>

            <label class="field">
              <span>密码</span>
              <input
                v-model="form.password"
                class="input"
                name="password"
                type="password"
                autocomplete="current-password"
                placeholder="teach123"
              />
            </label>

            <button class="button button--primary" type="submit" :disabled="submitting">
              {{ submitting ? '登录中…' : '登录' }}
            </button>
          </form>

          <p v-if="runtime.showMockLoginHint" class="helper-text">
            Mock 登录：
            <code>teacher</code> / <code>teach123</code>
            ·
            <code>admin</code> / <code>admin12345</code>
          </p>

          <p
            v-if="feedback"
            :role="feedbackTone === 'error' ? 'alert' : 'status'"
            class="feedback"
            :class="`feedback--${feedbackTone}`"
          >
            {{ feedback }}
          </p>
        </section>
      </div>
    </main>

    <footer class="auth-footer">
      <div class="auth-frame auth-footer__bar">
        <p class="auth-footer__text">
          Scratch 教师后台 · 面向课堂教师的 AI 辅助工具
        </p>
        <div class="auth-footer__meta">
          <a
            class="auth-footer__link"
            :href="documentationHref"
            target="_blank"
            rel="noreferrer"
          >
            开发说明
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>
