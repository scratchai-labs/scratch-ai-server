<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import StatusBadge from '@/components/StatusBadge.vue'
import {
  useTeacherApiClient,
  type TeacherClassroomDetail,
  type TeacherRelease,
} from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const props = defineProps<{
  classroom: TeacherClassroomDetail | null
  classroomId: string
  refreshClassroom: () => Promise<void>
}>()

const apiClient = useTeacherApiClient()

const projects = ref<TeacherRelease[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const feedback = ref<string | null>(null)
const fileInputKey = ref(0)

const projectForm = ref({
  title: '',
  goal: '',
  description: '',
  file: null as File | null,
})

function projectTone(project: TeacherRelease) {
  return project.status === 'published' ? 'success' : 'warning'
}

function projectDetailRoute(project: TeacherRelease) {
  return {
    name: 'project-detail',
    params: {
      id: project.id,
    },
    query: {
      classroomId: project.classroomId || props.classroomId,
    },
  }
}

async function loadProjects() {
  if (!props.classroomId || !apiClient.listClassroomProjects) {
    projects.value = []
    return
  }

  loading.value = true
  error.value = null
  try {
    projects.value = await apiClient.listClassroomProjects(props.classroomId)
  } catch (err) {
    error.value = toErrorMessage(err, '项目列表加载失败')
  } finally {
    loading.value = false
  }
}

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  projectForm.value.file = target.files?.[0] ?? null
}

async function submitCreateProject() {
  if (!apiClient.createClassroomProject || !props.classroomId || !projectForm.value.file) {
    return
  }

  error.value = null
  feedback.value = null
  try {
    const created = await apiClient.createClassroomProject(props.classroomId, {
      title: projectForm.value.title,
      goal: projectForm.value.goal,
      description: projectForm.value.description,
      file: projectForm.value.file,
    })

    projects.value = [
      ...projects.value,
      {
        id: created.id,
        classroomId: props.classroomId,
        title: created.title,
        goal: projectForm.value.goal,
        description: projectForm.value.description,
        className: props.classroom?.name ?? '',
        status: created.status,
        analysisStatus: created.analysisStatus,
        studentCount: 0,
        updatedAt: '刚刚',
      },
    ]
    projectForm.value = {
      title: '',
      goal: '',
      description: '',
      file: null,
    }
    fileInputKey.value += 1
    feedback.value = `已创建项目 ${created.title}`
    await props.refreshClassroom()
  } catch (err) {
    error.value = toErrorMessage(err, '创建项目失败')
  }
}

watch(() => props.classroomId, () => {
  void loadProjects()
}, { immediate: true })
</script>

<template>
  <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>
  <p v-else-if="feedback" role="status" class="feedback feedback--success">{{ feedback }}</p>

  <section class="panel">
    <div class="panel__head">
      <div>
        <h2 class="panel__title">创建项目</h2>
        <p class="panel__meta">项目表单和 <code>sb3</code> 上传集中在一个区域，下面只保留项目列表。</p>
      </div>
    </div>

    <form class="form-grid form-grid--two-column" @submit.prevent="submitCreateProject">
      <label class="field">
        <span>项目标题</span>
        <input v-model="projectForm.title" class="input" placeholder="迷宫项目" />
      </label>
      <label class="field">
        <span>教学目标</span>
        <input v-model="projectForm.goal" class="input" placeholder="让角色按事件响应" />
      </label>
      <label class="field">
        <span>项目说明</span>
        <input v-model="projectForm.description" class="input" placeholder="第一节课项目" />
      </label>

      <div class="field">
        <span>上传 sb3</span>
        <label class="file-picker">
          <input
            :key="fileInputKey"
            class="sr-only"
            type="file"
            accept=".sb3"
            @change="onFileChange"
          />
          <span class="file-picker__value">{{ projectForm.file?.name || '选择 .sb3 文件' }}</span>
          <span class="file-picker__action">浏览文件</span>
        </label>
      </div>

      <button class="button button--primary" type="submit">创建项目</button>
    </form>
  </section>

  <section class="panel">
    <div class="panel__head">
      <div>
        <h2 class="panel__title">项目列表</h2>
        <p class="panel__meta">创建完成后，再从列表进入项目详情查看分析结果与学生实时进度。</p>
      </div>
    </div>

    <div v-if="loading && !projects.length" class="empty-state">正在拉取项目列表…</div>
    <div v-else-if="!projects.length" class="empty-state">当前班级还没有项目，先上传一个参考项目。</div>

    <div v-else class="card-grid">
      <article v-for="project in projects" :key="project.id" class="release-card">
        <div class="release-card__head">
          <div>
            <h2>{{ project.title }}</h2>
            <p>{{ project.goal }}</p>
          </div>
          <StatusBadge :tone="projectTone(project)">
            {{ project.status }}
          </StatusBadge>
        </div>

        <dl class="release-card__meta">
          <div>
            <dt>项目说明</dt>
            <dd>{{ project.description || '—' }}</dd>
          </div>
          <div>
            <dt>分析状态</dt>
            <dd>{{ project.analysisStatus }}</dd>
          </div>
          <div>
            <dt>最近更新</dt>
            <dd>{{ project.updatedAt }}</dd>
          </div>
        </dl>

        <RouterLink class="button button--primary" :to="projectDetailRoute(project)">查看项目详情</RouterLink>
      </article>
    </div>
  </section>
</template>
