<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useTeacherApiClient, type AdminAuditLog } from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const apiClient = useTeacherApiClient()

const logs = ref<AdminAuditLog[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const actionFilter = ref('all')

const actionOptions = computed(() => {
  return Array.from(new Set(logs.value.map((item) => item.action))).sort()
})

const filteredLogs = computed(() => {
  if (actionFilter.value === 'all') {
    return logs.value
  }

  return logs.value.filter((item) => item.action === actionFilter.value)
})

async function reloadAuditLogs() {
  loading.value = true
  error.value = null

  try {
    const nextLogs = await apiClient.listAdminAuditLogs?.()
    if (!nextLogs) {
      throw new Error('管理员审计日志接口未提供')
    }
    logs.value = [...nextLogs]
  } catch (nextError) {
    error.value = toErrorMessage(nextError, '操作日志加载失败')
  } finally {
    loading.value = false
  }
}

function summaryText(snapshot: Record<string, string>) {
  const entries = Object.entries(snapshot)
  if (!entries.length) {
    return '—'
  }

  return entries.map(([key, value]) => `${key}: ${value}`).join(' · ')
}

function actionLabel(action: string) {
  const labels: Record<string, string> = {
    'teacher.create': '教师创建',
    'teacher.password_reset': '教师密码重置',
    'teacher.disable': '教师禁用',
    'teacher.enable': '教师启用',
    'teacher.role_change': '教师角色变更',
    'student.create': '学生创建',
    'student.password_reset': '学生密码重置',
    'student.disable': '学生禁用',
    'student.enable': '学生启用',
  }

  return labels[action] ?? action
}

onMounted(() => {
  void reloadAuditLogs()
})
</script>

<template>
  <AppShell
    title="操作日志"
    description="记录管理员对教师与学生账号的敏感操作，先覆盖创建、密码重置、角色变更与启停。"
  >
    <template #actions>
      <StatusBadge :tone="loading ? 'warning' : 'success'">
        {{ loading ? '加载中' : `${filteredLogs.length} 条记录` }}
      </StatusBadge>
      <button class="button button--ghost" type="button" :disabled="loading" @click="reloadAuditLogs">
        刷新日志
      </button>
    </template>

    <p v-if="error" role="alert" class="feedback feedback--error">
      {{ error }}
    </p>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">筛选条件</h2>
          <p class="panel__meta">MVP 先支持按 action 本地筛选，便于快速定位账号治理操作。</p>
        </div>
      </div>

      <label class="field field--inline">
        <span>操作类型</span>
        <select v-model="actionFilter" class="input" name="audit-action-filter">
          <option value="all">全部操作</option>
          <option v-for="action in actionOptions" :key="action" :value="action">
            {{ actionLabel(action) }}
          </option>
        </select>
        </label>
      </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">审计记录</h2>
          <p class="panel__meta">按时间倒序展示敏感操作的执行人、目标对象与变更前后快照。</p>
        </div>
      </div>

      <div v-if="loading && !filteredLogs.length" class="empty-state">
        正在拉取审计日志…
      </div>
      <div v-else-if="!filteredLogs.length" class="empty-state">
        暂无匹配日志
      </div>

      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>执行人</th>
              <th>操作</th>
              <th>目标</th>
              <th>变更前</th>
              <th>变更后</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in filteredLogs" :key="log.id">
              <td>{{ log.createdAt }}</td>
              <td>{{ log.actorUsername }}</td>
              <td>
                <div class="stack">
                  <strong>{{ log.action }}</strong>
                  <span class="cell-subtle">{{ log.targetType }}</span>
                </div>
              </td>
              <td>
                <div class="stack">
                  <strong>{{ log.targetUsername }}</strong>
                  <span class="cell-subtle">ID: {{ log.targetId }}</span>
                </div>
              </td>
              <td>{{ summaryText(log.before) }}</td>
              <td>{{ summaryText(log.after) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </AppShell>
</template>
