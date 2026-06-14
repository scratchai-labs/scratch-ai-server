<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { studentStatusLabel, studentStatusTone } from '@/presenters/studentStatus'
import { useTeacherApiClient } from '@/services/teacherApi'
import { useTeacherDirectoryStore } from '@/stores/teacherDirectory'
import { toErrorMessage } from '@/stores/storeUtils'

const apiClient = useTeacherApiClient()
const directoryStore = useTeacherDirectoryStore()

const students = computed(() => directoryStore.students)
const createForm = ref({
  username: '',
  displayName: '',
  initialPassword: '',
})
const resetPasswords = ref<Record<string, string>>({})
const saving = ref(false)
const actionError = ref<string | null>(null)
const actionFeedback = ref('')

async function reloadStudents() {
  await directoryStore.loadStudents(apiClient)
}

async function submitCreateStudent() {
  if (!apiClient.createStudent) {
    actionError.value = '当前环境暂不支持创建学生'
    actionFeedback.value = ''
    return
  }

  saving.value = true
  actionError.value = null
  actionFeedback.value = ''

  try {
    const createdStudent = await apiClient.createStudent({
      username: createForm.value.username.trim(),
      displayName: createForm.value.displayName.trim(),
      initialPassword: createForm.value.initialPassword,
    })

    directoryStore.students = [...directoryStore.students, createdStudent]
    createForm.value = {
      username: '',
      displayName: '',
      initialPassword: '',
    }
    actionFeedback.value = `已创建学生账号 ${createdStudent.username}`
  } catch (error) {
    actionError.value = toErrorMessage(error, '创建学生失败')
  } finally {
    saving.value = false
  }
}

async function submitResetStudentPassword(studentId: string) {
  const newPassword = resetPasswords.value[studentId] ?? ''
  if (!newPassword.trim()) {
    return
  }

  if (!apiClient.resetStudentPassword) {
    actionError.value = '当前环境暂不支持重置学生密码'
    actionFeedback.value = ''
    return
  }

  saving.value = true
  actionError.value = null
  actionFeedback.value = ''

  try {
    const updatedStudent = await apiClient.resetStudentPassword(studentId, newPassword)
    directoryStore.students = directoryStore.students.map((student) =>
      student.id === studentId ? { ...student, ...updatedStudent } : student,
    )
    resetPasswords.value = {
      ...resetPasswords.value,
      [studentId]: '',
    }
    actionFeedback.value = `已重置 ${updatedStudent.username} 的密码`
  } catch (error) {
    actionError.value = toErrorMessage(error, '重置学生密码失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void reloadStudents()
})
</script>

<template>
  <AppShell
    title="学生管理"
    description="展示学生最新进度、AI 提示和更新时间。"
  >
    <template #actions>
      <StatusBadge :tone="directoryStore.studentsLoading ? 'warning' : 'success'">
        {{ directoryStore.studentsLoading ? '加载中' : `${directoryStore.studentCount} 名学生` }}
      </StatusBadge>
      <button class="button button--ghost" type="button" :disabled="directoryStore.studentsLoading" @click="reloadStudents">
        刷新列表
      </button>
    </template>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">新建学生</h2>
          <p class="panel__meta">支持教师直接补录学生账号，便于课前准备或临时加人。</p>
        </div>
      </div>

      <form class="form-grid" data-testid="create-student-form" @submit.prevent="submitCreateStudent">
        <label class="field">
          <span>学生账号</span>
          <input
            v-model="createForm.username"
            class="input"
            name="student-username"
            autocomplete="username"
            placeholder="student-01"
          />
        </label>

        <label class="field">
          <span>显示名</span>
          <input
            v-model="createForm.displayName"
            class="input"
            name="student-display-name"
            placeholder="小明"
          />
        </label>

        <label class="field">
          <span>初始密码</span>
          <input
            v-model="createForm.initialPassword"
            class="input"
            name="student-password"
            type="password"
            autocomplete="new-password"
            placeholder="abc12345"
          />
        </label>

        <button class="button button--primary" type="submit" :disabled="saving">
          创建学生
        </button>
      </form>

      <p v-if="actionError" role="alert" class="feedback feedback--error">
        {{ actionError }}
      </p>
      <p v-else-if="actionFeedback" role="status" class="feedback feedback--success">
        {{ actionFeedback }}
      </p>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">学生列表</h2>
          <p class="panel__meta">列表会结合学生基础档案和最近一次学习历史，显示每名学生当前状态。</p>
        </div>
      </div>

      <p v-if="directoryStore.studentsError" role="alert" class="feedback feedback--error">
        {{ directoryStore.studentsError }}
      </p>

      <div v-if="!directoryStore.studentsLoading && !students.length" class="empty-state">
        暂无学生数据
      </div>

      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>学生</th>
              <th>账号</th>
              <th>班级</th>
              <th>最新进度</th>
              <th>最新 AI 提示</th>
              <th>更新时间</th>
              <th>重置密码</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="student in students" :key="student.id">
              <td>{{ student.name }}</td>
              <td>{{ student.username || '—' }}</td>
              <td>{{ student.className }}</td>
              <td>
                <template v-if="student.progress > 0">
                  <div class="progress-track" :aria-label="`${student.name} 进度 ${student.progress}%`">
                    <div class="progress-bar" :style="{ width: `${student.progress}%` }" />
                  </div>
                  <span class="cell-subtle">{{ student.progress }}%</span>
                </template>
                <template v-else>
                  <StatusBadge :tone="studentStatusTone(student.status)">
                    {{ studentStatusLabel(student) }}
                  </StatusBadge>
                </template>
                <div v-if="student.currentTarget" class="cell-subtle">
                  当前目标：{{ student.currentTarget }}
                </div>
                <div v-if="student.stepSummary" class="cell-subtle">
                  当前步骤：{{ student.stepSummary }}
                </div>
              </td>
              <td>{{ student.latestAiHint }}</td>
              <td>{{ student.updatedAt }}</td>
              <td>
                <div class="stack">
                  <input
                    v-model="resetPasswords[student.id]"
                    class="input"
                    :name="`reset-student-password-${student.id}`"
                    type="password"
                    autocomplete="new-password"
                    placeholder="输入新密码"
                  />
                  <button
                    class="button button--ghost"
                    type="button"
                    :data-testid="`student-reset-${student.id}`"
                    :disabled="saving || !resetPasswords[student.id]?.trim()"
                    @click="submitResetStudentPassword(student.id)"
                  >
                    重置密码
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </AppShell>
</template>
