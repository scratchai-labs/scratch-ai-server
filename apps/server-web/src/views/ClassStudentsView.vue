<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  buildStudentBatchCreateInputs,
  studentBatchTemplate,
} from '@/services/studentBatchImport'
import {
  useTeacherApiClient,
  type TeacherClassroomDetail,
  type TeacherStudent,
} from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const props = defineProps<{
  classroom: TeacherClassroomDetail | null
  classroomId: string
  refreshClassroom: () => Promise<void>
}>()

const apiClient = useTeacherApiClient()

const students = ref<TeacherStudent[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const feedback = ref<string | null>(null)

const createStudentForm = ref({
  username: '',
  displayName: '',
  initialPassword: '',
})
const batchForm = ref({
  defaultPassword: '',
  pastedText: '',
})

async function loadStudents() {
  if (!props.classroomId || !apiClient.listClassroomStudents) {
    students.value = []
    return
  }

  loading.value = true
  error.value = null
  try {
    students.value = await apiClient.listClassroomStudents(props.classroomId)
  } catch (err) {
    error.value = toErrorMessage(err, '学生列表加载失败')
  } finally {
    loading.value = false
  }
}

async function submitCreateStudent() {
  if (!apiClient.createClassroomStudent || !props.classroomId) {
    return
  }

  error.value = null
  feedback.value = null
  try {
    const created = await apiClient.createClassroomStudent(props.classroomId, createStudentForm.value)
    students.value = [...students.value, created]
    createStudentForm.value = { username: '', displayName: '', initialPassword: '' }
    feedback.value = `已创建学生账号 ${created.username}`
    await props.refreshClassroom()
  } catch (err) {
    error.value = toErrorMessage(err, '创建学生失败')
  }
}

async function submitBatchCreateStudents() {
  if (!apiClient.batchCreateClassroomStudents || !props.classroomId) {
    return
  }

  error.value = null
  feedback.value = null
  try {
    const inputs = buildStudentBatchCreateInputs({
      pastedText: batchForm.value.pastedText,
      defaultPassword: batchForm.value.defaultPassword,
      existingUsernames: students.value.map((student) => student.username),
    })
    const result = await apiClient.batchCreateClassroomStudents(props.classroomId, inputs)
    students.value = [...students.value, ...result.created]
    batchForm.value.pastedText = ''
    feedback.value = result.conflicts.length
      ? `已批量创建 ${result.created.length} 名学生，冲突账号：${result.conflicts.join('、')}`
      : `已批量创建 ${result.created.length} 名学生`
    await props.refreshClassroom()
  } catch (err) {
    error.value = toErrorMessage(err, '批量导入学生失败')
  }
}

watch(() => props.classroomId, () => {
  void loadStudents()
}, { immediate: true })
</script>

<template>
  <p v-if="error" role="alert" class="feedback feedback--error">{{ error }}</p>
  <p v-else-if="feedback" role="status" class="feedback feedback--success">{{ feedback }}</p>

  <section class="section-grid section-grid--aside">
    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">创建学生</h2>
          <p class="panel__meta">单个创建适合补录、临时加人或课前准备少量账号。</p>
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
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">批量导入学生</h2>
          <p class="panel__meta">下载模板、按列填写，再把表格内容粘贴回来批量创建。</p>
        </div>
      </div>

      <div class="batch-import-layout">
        <div class="batch-import-guide">
          <p class="batch-import-guide__eyebrow">Batch import</p>
          <p class="batch-import-guide__note">先下载模板，再从第 8 行开始录入，最后复制 A 到 C 列粘贴回来。</p>
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
    </section>
  </section>

  <section class="panel">
    <div class="panel__head">
      <div>
        <h2 class="panel__title">学生列表</h2>
        <p class="panel__meta">导入与创建完成后，再集中检查学生账号、显示名和当前提示。</p>
      </div>
    </div>

    <div v-if="loading && !students.length" class="empty-state">正在拉取学生列表…</div>
    <div v-else-if="!students.length" class="empty-state">当前班级还没有学生，先创建一个账号或批量导入。</div>

    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>账号</th>
            <th>姓名</th>
            <th>当前目标</th>
            <th>当前提示</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="student in students" :key="student.id">
            <td>{{ student.username }}</td>
            <td>{{ student.name }}</td>
            <td>{{ student.currentTarget || '—' }}</td>
            <td>{{ student.latestAiHint }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
