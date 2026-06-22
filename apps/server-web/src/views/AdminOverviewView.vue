<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useTeacherApiClient, type AdminOverview } from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const apiClient = useTeacherApiClient()

const loading = ref(false)
const error = ref<string | null>(null)
const overview = ref<AdminOverview>({
  adminCount: 0,
  teacherCount: 0,
  activeTeacherCount: 0,
  disabledTeacherCount: 0,
  studentCount: 0,
  activeStudentCount: 0,
  disabledStudentCount: 0,
})

async function reloadOverview() {
  loading.value = true
  error.value = null

  try {
    const nextOverview = await apiClient.getAdminOverview?.()
    if (!nextOverview) {
      throw new Error('管理员总览接口未提供')
    }
    overview.value = nextOverview
  } catch (nextError) {
    error.value = toErrorMessage(nextError, '管理员总览加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void reloadOverview()
})
</script>

<template>
  <AppShell
    title="后台总览"
    description="集中查看管理员、教师、学生账号规模与当前启停状态，作为后台运维入口。"
  >
    <template #actions>
      <StatusBadge :tone="loading ? 'warning' : 'success'">
        {{ loading ? '加载中' : '数据已同步' }}
      </StatusBadge>
      <button class="button button--ghost" type="button" :disabled="loading" @click="reloadOverview">
        刷新总览
      </button>
    </template>

    <p v-if="error" role="alert" class="feedback feedback--error">
      {{ error }}
    </p>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">账号规模总览</h2>
          <p class="panel__meta">先看管理员、教师和学生的当前规模，再进入具体治理页。</p>
        </div>
      </div>

      <div class="metric-grid">
        <article class="metric-card">
          <p class="metric-card__label">管理员账号</p>
          <strong class="metric-card__value">{{ overview.adminCount }}</strong>
          <p class="metric-card__note">当前拥有后台权限的系统账号数量。</p>
        </article>

        <article class="metric-card">
          <p class="metric-card__label">教师账号</p>
          <strong class="metric-card__value">{{ overview.teacherCount }}</strong>
          <p class="metric-card__note">
            启用 {{ overview.activeTeacherCount }} · 禁用 {{ overview.disabledTeacherCount }}
          </p>
        </article>

        <article class="metric-card">
          <p class="metric-card__label">学生账号</p>
          <strong class="metric-card__value">{{ overview.studentCount }}</strong>
          <p class="metric-card__note">
            启用 {{ overview.activeStudentCount }} · 禁用 {{ overview.disabledStudentCount }}
          </p>
        </article>

        <article class="metric-card">
          <p class="metric-card__label">禁用学生</p>
          <strong class="metric-card__value">{{ overview.disabledStudentCount }}</strong>
          <p class="metric-card__note">优先检查是否为已毕业、误操作或临时停用账号。</p>
        </article>
      </div>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">当前后台边界</h2>
          <p class="panel__meta">第一批先收口账号治理：管理员看总览、管教师、管学生，不改教师课堂主链路。</p>
        </div>
      </div>

      <div class="card-grid">
        <article class="release-card">
          <div class="release-card__head">
            <h2>教师账号治理</h2>
            <StatusBadge tone="info">已接通</StatusBadge>
          </div>
          <p class="metric-card__note">支持新建教师、重置密码、账号启停，教师仍登录原有教学后台。</p>
        </article>

        <article class="release-card">
          <div class="release-card__head">
            <h2>学生账号治理</h2>
            <StatusBadge tone="info">已接通</StatusBadge>
          </div>
          <p class="metric-card__note">支持管理员全局查看学生、重置密码、启停账号，便于统一运维排障。</p>
        </article>
      </div>
    </section>
  </AppShell>
</template>
