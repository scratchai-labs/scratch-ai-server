<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import {
  useTeacherApiClient,
  type LiveDashboardSnapshot,
  type TeacherReleaseAnalysis,
  type TeacherReleaseDetail,
} from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const route = useRoute()
const apiClient = useTeacherApiClient()
const projectId = String(route.params.id ?? '')
const classroomId = computed(() => {
  const value = route.query.classroomId
  return typeof value === 'string' ? value : ''
})
const returnLink = computed(() => classroomId.value ? `/classes/${classroomId.value}/projects` : '/classes')
const returnLabel = computed(() => classroomId.value ? '返回班级项目' : '返回班级管理')

const detail = ref<TeacherReleaseDetail | null>(null)
const analysis = ref<TeacherReleaseAnalysis | null>(null)
const live = ref<LiveDashboardSnapshot | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

async function loadProjectDetail() {
  if (!projectId) {
    return
  }

  loading.value = true
  error.value = null
  try {
    const [nextDetail, nextAnalysis, nextLive] = await Promise.all([
      apiClient.getReleaseDetail?.(projectId),
      apiClient.getReleaseAnalysis?.(projectId),
      apiClient.getLiveDashboard(projectId),
    ])
    detail.value = nextDetail ?? null
    analysis.value = nextAnalysis ?? null
    live.value = nextLive
  } catch (err) {
    error.value = toErrorMessage(err, '项目详情加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadProjectDetail()
})
</script>

<template>
  <AppShell
    :title="detail?.title || '项目详情'"
    description="查看班级项目分析结果、每个学生当前进度和当前提示。"
  >
    <template #actions>
      <div class="inline-actions">
        <RouterLink class="button button--ghost" :to="returnLink">{{ returnLabel }}</RouterLink>
        <StatusBadge :tone="loading ? 'warning' : 'info'">
          {{ loading ? '加载中' : `${live?.students.length ?? 0} 名学生` }}
        </StatusBadge>
      </div>
    </template>

    <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">项目概览</h2>
          <p class="panel__meta">{{ detail?.description || '项目概览加载完成后会显示任务说明。' }}</p>
        </div>
        <StatusBadge
          :tone="loading ? 'warning' : detail ? (detail.status === 'published' ? 'success' : 'warning') : 'muted'"
        >
          {{ detail ? detail.status : loading ? '加载中' : '空' }}
        </StatusBadge>
      </div>

      <div v-if="loading && (!detail || !analysis)" class="empty-state">正在拉取项目概览…</div>
      <div v-else-if="detail && analysis" class="table-wrap">
        <table class="data-table">
          <tbody>
            <tr>
              <th>教学目标</th>
              <td>{{ detail.goal }}</td>
            </tr>
            <tr>
              <th>分析状态</th>
              <td>{{ analysis.analysisStatus }}</td>
            </tr>
            <tr>
              <th>教学点</th>
              <td>{{ analysis.teachingPoints.join('、') || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">暂无项目概览</div>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">学生当前进度与提示</h2>
          <p class="panel__meta">第一版先展示每个学生当前目标、当前步骤和当前提示，不做教师端重生成。</p>
        </div>
      </div>

      <div v-if="loading && !live" class="empty-state">正在拉取项目实时进度…</div>
      <div v-else-if="live" class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>学生</th>
              <th>当前目标</th>
              <th>当前步骤</th>
              <th>当前提示</th>
              <th>更新时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="student in live.students" :key="student.id">
              <td>{{ student.name }}</td>
              <td>{{ student.currentTarget || '—' }}</td>
              <td>{{ student.stepSummary || '—' }}</td>
              <td>{{ student.latestAiHint }}</td>
              <td>{{ student.updatedAt }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">暂无项目实时进度</div>
    </section>
  </AppShell>
</template>
