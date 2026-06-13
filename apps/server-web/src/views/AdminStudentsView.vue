<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useTeacherApiClient, type ManagedStudent } from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'

const apiClient = useTeacherApiClient()

const students = ref<ManagedStudent[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref<string | null>(null)
const feedback = ref('')
const resetPasswords = reactive<Record<string, string>>({})

const totalStudents = computed(() => students.value.length)

async function reloadStudents() {
  loading.value = true
  error.value = null

  try {
    const nextStudents = await apiClient.listManagedStudents?.()
    if (!nextStudents) {
      throw new Error('管理员学生列表接口未提供')
    }
    students.value = [...nextStudents].sort(compareStudentsByCreatedAt)
  } catch (nextError) {
    error.value = toErrorMessage(nextError, '学生列表加载失败')
  } finally {
    loading.value = false
  }
}

async function submitResetPassword(studentId: string) {
  const nextPassword = resetPasswords[studentId]?.trim()
  if (!nextPassword) {
    return
  }

  saving.value = true
  error.value = null

  try {
    const updatedStudent = await apiClient.resetManagedStudentPassword?.(studentId, nextPassword)
    if (!updatedStudent) {
      throw new Error('管理员学生密码重置接口未提供')
    }
    mergeStudent(updatedStudent)
    feedback.value = `已重置 ${updatedStudent.username} 的密码`
    resetPasswords[studentId] = ''
  } catch (nextError) {
    error.value = toErrorMessage(nextError, '学生密码重置失败')
  } finally {
    saving.value = false
  }
}

async function toggleStudentStatus(studentId: string, currentStatus: string) {
  saving.value = true
  error.value = null

  try {
    const updatedStudent =
      currentStatus === 'disabled'
        ? await apiClient.enableManagedStudent?.(studentId)
        : await apiClient.disableManagedStudent?.(studentId)
    if (!updatedStudent) {
      throw new Error('管理员学生状态接口未提供')
    }
    mergeStudent(updatedStudent)
    feedback.value = currentStatus === 'disabled'
      ? `已启用 ${updatedStudent.username}`
      : `已禁用 ${updatedStudent.username}`
  } catch (nextError) {
    error.value = toErrorMessage(nextError, '学生状态更新失败')
  } finally {
    saving.value = false
  }
}

function mergeStudent(updatedStudent: ManagedStudent) {
  students.value = students.value
    .map((student) => (student.id === updatedStudent.id ? updatedStudent : student))
    .sort(compareStudentsByCreatedAt)
}

onMounted(() => {
  void reloadStudents()
})

function compareStudentsByCreatedAt(left: ManagedStudent, right: ManagedStudent) {
  return Date.parse(left.createdAt) - Date.parse(right.createdAt)
}
</script>

<template>
  <AppShell
    title="学生管理"
    description="管理员统一查看学生归属、账号状态，并可直接执行密码重置和启停控制。"
  >
    <template #actions>
      <StatusBadge :tone="loading ? 'warning' : 'success'">
        {{ loading ? '加载中' : `${totalStudents} 个学生账号` }}
      </StatusBadge>
      <button class="button button--ghost" type="button" :disabled="loading" @click="reloadStudents">
        刷新列表
      </button>
    </template>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">学生账号总览</h2>
          <p class="panel__meta">学生账号仍由教师创建；管理员在这里统一做全局查询、停用与密码重置。</p>
        </div>
      </div>

      <p v-if="error" role="alert" class="feedback feedback--error">
        {{ error }}
      </p>
      <p v-else-if="feedback" role="status" class="feedback feedback--success">
        {{ feedback }}
      </p>

      <div v-if="!loading && !students.length" class="empty-state">
        暂无学生账号
      </div>

      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>学生账号</th>
              <th>显示名</th>
              <th>所属教师</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>重置密码</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="student in students" :key="student.id">
              <td>{{ student.username }}</td>
              <td>{{ student.displayName }}</td>
              <td>{{ student.teacherUsername }}</td>
              <td>{{ student.status }}</td>
              <td>{{ student.createdAt }}</td>
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
                    :disabled="saving"
                    @click="submitResetPassword(student.id)"
                  >
                    重置密码
                  </button>
                </div>
              </td>
              <td>
                <button
                  class="button button--ghost"
                  type="button"
                  :data-testid="student.status === 'disabled' ? `student-enable-${student.id}` : `student-disable-${student.id}`"
                  :disabled="saving"
                  @click="toggleStudentStatus(student.id, student.status)"
                >
                  {{ student.status === 'disabled' ? '启用' : '禁用' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </AppShell>
</template>
