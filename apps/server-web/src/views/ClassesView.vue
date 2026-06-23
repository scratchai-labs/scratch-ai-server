<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useTeacherApiClient, type TeacherClassroom } from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const apiClient = useTeacherApiClient()

const classrooms = ref<TeacherClassroom[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const createName = ref('')
const creating = ref(false)

async function loadClassrooms() {
  if (!apiClient.listClassrooms) {
    classrooms.value = []
    return
  }

  loading.value = true
  error.value = null
  try {
    classrooms.value = await apiClient.listClassrooms()
  } catch (err) {
    error.value = toErrorMessage(err, '班级列表加载失败')
  } finally {
    loading.value = false
  }
}

async function submitCreateClassroom() {
  if (!apiClient.createClassroom || !createName.value.trim()) {
    return
  }

  creating.value = true
  error.value = null
  try {
    const created = await apiClient.createClassroom({
      name: createName.value.trim(),
    })
    classrooms.value = [...classrooms.value, created]
    createName.value = ''
  } catch (err) {
    error.value = toErrorMessage(err, '班级创建失败')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  void loadClassrooms()
})
</script>

<template>
  <AppShell title="班级管理" description="先创建班级，再进入班级管理学生和项目。">
    <template #actions>
      <StatusBadge :tone="loading ? 'warning' : 'success'">
        {{ loading ? '加载中' : `${classrooms.length} 个班级` }}
      </StatusBadge>
    </template>

    <section class="summary-hero">
      <div>
        <p class="summary-hero__eyebrow">Class workspace</p>
        <h2 class="summary-hero__title">班级是老师日常操作的主入口</h2>
        <p class="summary-hero__description">
          新建班级后，再进入对应工作区管理学生、批量导入和项目上传，不再把所有动作堆在一个长页面里。
        </p>
      </div>

      <div class="summary-grid">
        <article class="summary-stat">
          <p class="summary-stat__label">当前班级</p>
          <p class="summary-stat__value">{{ classrooms.length }}</p>
          <p class="summary-stat__note">已创建并可进入工作区的班级数量。</p>
        </article>
        <article class="summary-stat">
          <p class="summary-stat__label">当前状态</p>
          <p class="summary-stat__value">{{ loading ? '同步中' : '已就绪' }}</p>
          <p class="summary-stat__note">班级、学生和项目入口会按当前接口数据刷新。</p>
        </article>
      </div>
    </section>

    <section class="section-grid section-grid--aside">
      <div class="panel">
        <div>
          <h2 class="panel__title">新建班级</h2>
          <p class="panel__meta">先建班级，再把学生与项目分别收进班级工作区。</p>
        </div>

        <form class="form-grid" @submit.prevent="submitCreateClassroom">
          <label class="field">
            <span>班级名称</span>
            <input v-model="createName" class="input" placeholder="四年级一班" />
            <p class="field__hint">建议直接使用年级与班别命名，方便后续区分项目与学生。</p>
          </label>
          <button class="button button--primary" type="submit" :disabled="creating">创建班级</button>
        </form>

        <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>
      </div>

      <div class="panel">
        <div>
          <h2 class="panel__title">使用节奏</h2>
          <p class="panel__meta">首期把班级工作区拆成三段，老师进入后先看概览，再决定去学生还是项目页。</p>
        </div>

        <div class="quick-links">
          <article class="quick-link-card">
            <h2>学生页</h2>
            <p>单个创建、批量导入、学生列表各自成块，不再和项目创建挤在一起。</p>
          </article>
          <article class="quick-link-card">
            <h2>项目页</h2>
            <p>项目标题、教学目标、说明与 <code>sb3</code> 上传统一放到项目页，再往下看项目列表。</p>
          </article>
        </div>
      </div>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">班级列表</h2>
          <p class="panel__meta">每个班级进入后都拥有独立的概览、学生和项目子页。</p>
        </div>
      </div>

      <div v-if="!loading && !classrooms.length" class="empty-state">还没有班级，先创建一个班级。</div>

      <div v-else class="card-grid">
        <article v-for="classroom in classrooms" :key="classroom.id" class="release-card">
          <div class="release-card__head">
            <div>
              <h2>{{ classroom.name }}</h2>
              <p>{{ classroom.studentCount }} 名学生 · {{ classroom.projectCount }} 个项目</p>
            </div>
          </div>

          <dl class="release-card__meta">
            <div>
              <dt>创建时间</dt>
              <dd>{{ classroom.createdAt }}</dd>
            </div>
            <div>
              <dt>最近更新</dt>
              <dd>{{ classroom.updatedAt }}</dd>
            </div>
          </dl>

          <RouterLink class="button button--primary" :to="`/classes/${classroom.id}`">进入工作区</RouterLink>
        </article>
      </div>
    </section>
  </AppShell>
</template>
