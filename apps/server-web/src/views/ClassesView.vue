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

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">新建班级</h2>
          <p class="panel__meta">教师工作台先围绕班级组织，后续学生和项目都进入班级内管理。</p>
        </div>
      </div>

      <form class="form-grid" @submit.prevent="submitCreateClassroom">
        <label class="field">
          <span>班级名称</span>
          <input v-model="createName" class="input" placeholder="四年级一班" />
        </label>
        <button class="button button--primary" type="submit" :disabled="creating">创建班级</button>
      </form>

      <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">班级列表</h2>
          <p class="panel__meta">进入班级后默认先看到学生管理，再切换项目管理。</p>
        </div>
      </div>

      <div v-if="!loading && !classrooms.length" class="empty-state">还没有班级，先创建一个班级。</div>

      <div v-else class="card-grid">
        <article v-for="classroom in classrooms" :key="classroom.id" class="release-card">
          <div class="release-card__head">
            <div>
              <strong>{{ classroom.name }}</strong>
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

          <RouterLink class="button button--primary" :to="`/classes/${classroom.id}`">进入班级</RouterLink>
        </article>
      </div>
    </section>
  </AppShell>
</template>
