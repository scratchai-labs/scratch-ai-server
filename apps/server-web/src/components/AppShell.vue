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

const navigation = computed(() => [
  ...(session.isAdmin
    ? [
        {
          label: '后台总览',
          to: '/admin',
        },
        {
          label: '教师管理',
          to: '/admin/teachers',
        },
        {
          label: '学生管理',
          to: '/admin/students',
        },
        {
          label: '操作日志',
          to: '/admin/audit-logs',
        },
      ]
    : [
        {
          label: '班级管理',
          to: '/classes',
        },
        {
          label: '实时总览',
          to: '/dashboard',
        },
      ]),
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
    <aside class="shell__sidebar">
      <div class="shell__brand">
        <div class="shell__brand-mark">S</div>
        <div>
          <strong>{{ session.isAdmin ? 'Scratch 管理后台' : 'Scratch 教师后台' }}</strong>
          <p>{{ session.isAdmin ? '管理员入口 · Accounts Overview' : 'mockable API · Vue 3 + Vite' }}</p>
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

      <div class="shell__footer">
        <p class="shell__footer-label">当前教师</p>
        <strong>{{ session.teacherName || '未登录' }}</strong>
        <span>{{ session.isAuthenticated ? `${session.role} 会话已加载` : '请先登录' }}</span>
        <button class="button button--ghost" type="button" @click="handleLogout">
          退出登录
        </button>
      </div>
    </aside>

    <main class="shell__content">
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
    </main>
  </div>
</template>
