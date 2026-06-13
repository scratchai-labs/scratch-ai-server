<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTeacherApiClient } from '@/services/teacherApi'
import { useSessionStore } from '@/stores/session'

defineProps<{
  title: string
  description?: string
}>()

const router = useRouter()
const route = useRoute()
const apiClient = useTeacherApiClient()
const session = useSessionStore()
const documentationHref =
  'https://github.com/scratchai-labs/scratch-ai-server/blob/main/docs/server-development.zh-CN.md'

const navigation = computed(() => [
  {
    label: '实时总览',
    to: '/dashboard',
  },
  {
    label: '学生管理',
    to: '/students',
  },
  {
    label: '发布单管理',
    to: '/releases',
  },
])

function isActive(path: string) {
  if (path === '/dashboard') {
    return route.path === '/dashboard'
  }

  return route.path === path || route.path.startsWith(`${path}/`)
}

async function handleLogout() {
  try {
    await apiClient.logout?.()
  } finally {
    session.logout()
    await router.push('/login')
  }
}
</script>

<template>
  <div class="shell">
    <header class="shell__header">
      <div class="site-frame shell__header-bar">
        <div class="shell__brand">
          <div class="shell__brand-mark">S</div>
          <div>
            <strong>Scratch 教师后台</strong>
            <p>AI 辅助课堂教学工具</p>
          </div>
        </div>

        <nav class="shell__nav">
          <RouterLink
            v-for="item in navigation"
            :key="item.to"
            :to="item.to"
            class="shell__nav-link"
            :class="{ 'shell__nav-link--active': isActive(item.to) }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="shell__header-meta">
          <div class="shell__session">
            <p class="shell__session-label">当前教师</p>
            <strong>{{ session.teacherName || '未登录' }}</strong>
          </div>
          <button class="button button--ghost" type="button" @click="handleLogout">
            退出登录
          </button>
        </div>
      </div>
    </header>

    <main class="shell__main">
      <div class="site-frame shell__main-frame">
        <header class="page-header">
          <div class="stack">
            <p class="page-header__eyebrow">Teacher Console</p>
            <h1 class="page-header__title">{{ title }}</h1>
            <p v-if="description" class="page-header__description">
              {{ description }}
            </p>
          </div>
          <div class="page-header__actions">
            <slot name="actions" />
          </div>
        </header>

        <slot />
      </div>
    </main>

    <footer class="shell__site-footer">
      <div class="site-frame shell__site-footer-bar">
        <p class="shell__site-footer-text">Scratch 教师后台 · 课堂工具</p>
        <a
          class="shell__site-footer-link"
          :href="documentationHref"
          target="_blank"
          rel="noreferrer"
        >
          开发说明
        </a>
      </div>
    </footer>
  </div>
</template>
