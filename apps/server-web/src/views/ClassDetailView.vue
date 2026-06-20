<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import {
  buildStudentBatchCreateInputs,
  studentBatchTemplate,
} from '@/services/studentBatchImport'
import {
  useTeacherApiClient,
  type TeacherClassroomDetail,
  type TeacherRelease,
  type TeacherStudent,
} from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const route = useRoute()
const apiClient = useTeacherApiClient()
const classroomId = computed(() => String(route.params.id ?? ''))

const classroom = ref<TeacherClassroomDetail | null>(null)
const students = ref<TeacherStudent[]>([])
const projects = ref<TeacherRelease[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const createStudentForm = ref({
  username: '',
  displayName: '',
  initialPassword: '',
})
const batchForm = ref({
  defaultPassword: '',
  pastedText: '',
})
const projectForm = ref({
  title: '',
  goal: '',
  description: '',
  file: null as File | null,
})

async function loadClassroomData() {
  if (!classroomId.value || !apiClient.getClassroomDetail) {
    return
  }

  loading.value = true
  error.value = null
  try {
    const [detail, classroomStudents, classroomProjects] = await Promise.all([
      apiClient.getClassroomDetail(classroomId.value),
      apiClient.listClassroomStudents?.(classroomId.value) ?? [],
      apiClient.listClassroomProjects?.(classroomId.value) ?? [],
    ])
    classroom.value = detail
    students.value = classroomStudents
    projects.value = classroomProjects
  } catch (err) {
    error.value = toErrorMessage(err, '班级详情加载失败')
  } finally {
    loading.value = false
  }
}

async function submitCreateStudent() {
  if (!apiClient.createClassroomStudent || !classroomId.value) {
    return
  }

  try {
    const created = await apiClient.createClassroomStudent(classroomId.value, createStudentForm.value)
    students.value = [...students.value, created]
    createStudentForm.value = { username: '', displayName: '', initialPassword: '' }
  } catch (err) {
    error.value = toErrorMessage(err, '创建学生失败')
  }
}

async function submitBatchCreateStudents() {
  if (!apiClient.batchCreateClassroomStudents || !classroomId.value) {
    return
  }

  try {
    const inputs = buildStudentBatchCreateInputs({
      pastedText: batchForm.value.pastedText,
      defaultPassword: batchForm.value.defaultPassword,
      existingUsernames: students.value.map((student) => student.username),
    })
    const result = await apiClient.batchCreateClassroomStudents(classroomId.value, inputs)
    students.value = [...students.value, ...result.created]
    batchForm.value.pastedText = ''
  } catch (err) {
    error.value = toErrorMessage(err, '批量导入学生失败')
  }
}

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  projectForm.value.file = target.files?.[0] ?? null
}

async function submitCreateProject() {
  if (!apiClient.createClassroomProject || !classroomId.value || !projectForm.value.file) {
    return
  }

  try {
    const created = await apiClient.createClassroomProject(classroomId.value, {
      title: projectForm.value.title,
      goal: projectForm.value.goal,
      description: projectForm.value.description,
      file: projectForm.value.file,
    })
    projects.value = [
      ...projects.value,
      {
        id: created.id,
        classroomId: classroomId.value,
        title: created.title,
        goal: projectForm.value.goal,
        description: projectForm.value.description,
        className: classroom.value?.name ?? '',
        status: created.status,
        analysisStatus: created.analysisStatus,
        studentCount: 0,
        updatedAt: '刚刚',
      },
    ]
    projectForm.value = { title: '', goal: '', description: '', file: null }
  } catch (err) {
    error.value = toErrorMessage(err, '创建项目失败')
  }
}

onMounted(() => {
  void loadClassroomData()
})
</script>

<template>
  <AppShell
    :title="classroom?.name || '班级详情'"
    description="默认先看学生管理，再在同一页管理班级项目。"
  >
    <template #actions>
      <StatusBadge :tone="loading ? 'warning' : 'info'">
        {{ loading ? '加载中' : `${students.length} 名学生 · ${projects.length} 个项目` }}
      </StatusBadge>
    </template>

    <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">学生管理</h2>
          <p class="panel__meta">班级详情默认先展示学生，可单个新增或沿用 Excel 模板批量导入。</p>
        </div>
      </div>

      <form class="form-grid" @submit.prevent="submitCreateStudent">
        <label class="field">
          <span>学生账号</span>
          <input v-model="createStudentForm.username" class="input" placeholder="student-01" />
        </label>
        <label class="field">
          <span>显示名</span>
          <input v-model="createStudentForm.displayName" class="input" placeholder="小明" />
        </label>
        <label class="field">
          <span>初始密码</span>
          <input v-model="createStudentForm.initialPassword" class="input" type="password" placeholder="abc12345" />
        </label>
        <button class="button button--primary" type="submit">创建学生</button>
      </form>

      <div class="batch-import-layout">
        <div class="batch-import-guide">
          <p class="batch-import-guide__eyebrow">批量导入</p>
          <p class="batch-import-guide__note">继续沿用 Excel 模板粘贴导入方式。</p>
          <a class="button button--ghost" :href="studentBatchTemplate.href" :download="studentBatchTemplate.downloadName">
            下载 Excel 模板
          </a>
        </div>
        <form class="stack" @submit.prevent="submitBatchCreateStudents">
          <label class="field">
            <span>统一初始密码</span>
            <input v-model="batchForm.defaultPassword" class="input" type="password" />
          </label>
          <label class="field">
            <span>粘贴 Excel 内容</span>
            <textarea v-model="batchForm.pastedText" class="input" rows="8" />
          </label>
          <button class="button button--primary" type="submit">批量导入学生</button>
        </form>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>账号</th>
              <th>姓名</th>
              <th>当前提示</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="student in students" :key="student.id">
              <td>{{ student.username }}</td>
              <td>{{ student.name }}</td>
              <td>{{ student.latestAiHint }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">项目管理</h2>
          <p class="panel__meta">班级里的每个 Scratch3 都是一个小项目，进入项目详情后看每个学生进度和当前提示。</p>
        </div>
      </div>

      <form class="form-grid" @submit.prevent="submitCreateProject">
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
        <label class="field">
          <span>上传 sb3</span>
          <input class="input" type="file" accept=".sb3" @change="onFileChange" />
        </label>
        <button class="button button--primary" type="submit">创建项目</button>
      </form>

      <div class="card-grid">
        <article v-for="project in projects" :key="project.id" class="release-card">
          <div class="release-card__head">
            <div>
              <h2>{{ project.title }}</h2>
              <p>{{ project.goal }}</p>
            </div>
            <StatusBadge :tone="project.status === 'published' ? 'success' : 'warning'">
              {{ project.status }}
            </StatusBadge>
          </div>
          <p class="cell-subtle">分析状态：{{ project.analysisStatus }}</p>
          <RouterLink class="button button--primary" :to="`/projects/${project.id}`">查看项目详情</RouterLink>
        </article>
      </div>
    </section>
  </AppShell>
</template>
