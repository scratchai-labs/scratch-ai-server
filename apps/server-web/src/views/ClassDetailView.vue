<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useTeacherApiClient, type TeacherClassroomDetail } from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const route = useRoute()
const apiClient = useTeacherApiClient()

const classroom = ref<TeacherClassroomDetail | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

const classroomId = computed(() => String(route.params.id ?? ''))
const studentCountText = computed(() => `${classroom.value?.studentCount ?? 0} 名学生`)
const projectCountText = computed(() => `${classroom.value?.projectCount ?? 0} 个项目`)

const tabs = computed(() => [
  {
    label: '概览',
    to: `/classes/${classroomId.value}`,
  },
  {
    label: '学生',
    to: `/classes/${classroomId.value}/students`,
  },
  {
    label: '项目',
    to: `/classes/${classroomId.value}/projects`,
  },
])

async function loadClassroomData() {
  if (!classroomId.value || !apiClient.getClassroomDetail) {
    classroom.value = null
    return
  }

  loading.value = true
  error.value = null
  try {
    classroom.value = await apiClient.getClassroomDetail(classroomId.value)
  } catch (err) {
    error.value = toErrorMessage(err, '班级详情加载失败')
  } finally {
    loading.value = false
  }
}

watch(classroomId, () => {
  void loadClassroomData()
}, { immediate: true })
</script>

<template>
  <AppShell
    :title="classroom?.name || '班级工作区'"
    description="把班级概览、学生管理和项目管理拆开，避免老师在一个长页面里同时处理所有动作。"
  >
    <template #actions>
      <StatusBadge :tone="loading ? 'warning' : 'info'">
        {{ loading ? '加载中' : `${studentCountText} · ${projectCountText}` }}
      </StatusBadge>
    </template>

    <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>

    <section class="summary-hero">
      <div>
        <p class="summary-hero__eyebrow">Class workspace</p>
        <h2 class="summary-hero__title">{{ classroom?.name || '班级工作区' }}</h2>
        <p class="summary-hero__description">
          先看班级概览，再进入学生或项目子页处理具体动作。这样老师不会在一个页面里同时面对导入、创建、上传和列表。
        </p>
      </div>

      <div class="summary-grid">
        <article class="summary-stat">
          <p class="summary-stat__label">学生数</p>
          <p class="summary-stat__value">{{ classroom?.studentCount ?? 0 }}</p>
          <p class="summary-stat__note">创建学生和批量导入都集中在学生子页。</p>
        </article>
        <article class="summary-stat">
          <p class="summary-stat__label">项目数</p>
          <p class="summary-stat__value">{{ classroom?.projectCount ?? 0 }}</p>
          <p class="summary-stat__note">创建项目与 <code>sb3</code> 上传都集中在项目子页。</p>
        </article>
        <article class="summary-stat">
          <p class="summary-stat__label">最近更新</p>
          <p class="summary-stat__value">{{ classroom?.updatedAt || '—' }}</p>
          <p class="summary-stat__note">子页里的新增操作会同步刷新这里的摘要。</p>
        </article>
      </div>
    </section>

    <nav class="workspace-nav" aria-label="班级工作区导航">
      <RouterLink v-for="tab in tabs" :key="tab.to" :to="tab.to" custom v-slot="{ href, navigate, isExactActive }">
        <a
          :href="href"
          class="workspace-nav__link"
          :class="{ 'workspace-nav__link--active': isExactActive }"
          @click="navigate"
        >
          {{ tab.label }}
        </a>
      </RouterLink>
    </nav>

    <RouterView v-slot="{ Component }">
      <component
        :is="Component"
        :classroom="classroom"
        :classroom-id="classroomId"
        :refresh-classroom="loadClassroomData"
      />
    </RouterView>
  </AppShell>
</template>
