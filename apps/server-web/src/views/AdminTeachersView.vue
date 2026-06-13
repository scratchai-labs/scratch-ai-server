<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useTeacherApiClient } from '@/services/teacherApi'
import { useAdminTeacherDirectoryStore } from '@/stores/adminTeacherDirectory'

const apiClient = useTeacherApiClient()
const directoryStore = useAdminTeacherDirectoryStore()

const createForm = reactive({
  username: '',
  initialPassword: '',
})

const resetPasswords = reactive<Record<string, string>>({})

const teachers = computed(() => directoryStore.teachers)

async function reloadTeachers() {
  await directoryStore.loadTeachers(apiClient)
}

async function submitCreateTeacher() {
  if (!createForm.username.trim() || !createForm.initialPassword.trim()) {
    return
  }

  await directoryStore.createTeacher(apiClient, {
    username: createForm.username.trim(),
    initialPassword: createForm.initialPassword,
  })

  createForm.username = ''
  createForm.initialPassword = ''
}

async function submitResetPassword(teacherId: string) {
  const nextPassword = resetPasswords[teacherId]?.trim()
  if (!nextPassword) {
    return
  }

  await directoryStore.resetTeacherPassword(apiClient, teacherId, nextPassword)
  resetPasswords[teacherId] = ''
}

async function toggleTeacherStatus(teacherId: string, currentStatus: string) {
  if (currentStatus === 'disabled') {
    await directoryStore.enableTeacher(apiClient, teacherId)
    return
  }

  await directoryStore.disableTeacher(apiClient, teacherId)
}

onMounted(() => {
  void reloadTeachers()
})
</script>

<template>
  <AppShell
    title="教师管理"
    description="管理员统一创建教师、重置密码，并控制教师账号启停。"
  >
    <template #actions>
      <StatusBadge :tone="directoryStore.loading ? 'warning' : 'success'">
        {{ directoryStore.loading ? '加载中' : `${teachers.length} 个账号` }}
      </StatusBadge>
      <button class="button button--ghost" type="button" :disabled="directoryStore.loading" @click="reloadTeachers">
        刷新列表
      </button>
    </template>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">新建教师</h2>
          <p class="panel__meta">创建后的教师账号可直接登录原有教师后台，管理自己的学生和发布单。</p>
        </div>
      </div>

      <form class="form-grid" data-testid="create-teacher-form" @submit.prevent="submitCreateTeacher">
        <label class="field">
          <span>教师账号</span>
          <input
            v-model="createForm.username"
            class="input"
            name="teacher-username"
            autocomplete="username"
            placeholder="teacher-01"
          />
        </label>

        <label class="field">
          <span>初始密码</span>
          <input
            v-model="createForm.initialPassword"
            class="input"
            name="teacher-password"
            type="password"
            autocomplete="new-password"
            placeholder="secret123"
          />
        </label>

        <button class="button button--primary" type="submit" :disabled="directoryStore.saving">
          创建教师
        </button>
      </form>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">教师账号列表</h2>
          <p class="panel__meta">支持查看角色、状态，并对单个教师执行密码重置和启停操作。</p>
        </div>
      </div>

      <p v-if="directoryStore.error" role="alert" class="feedback feedback--error">
        {{ directoryStore.error }}
      </p>
      <p v-else-if="directoryStore.feedback" role="status" class="feedback feedback--success">
        {{ directoryStore.feedback }}
      </p>

      <div v-if="!directoryStore.loading && !teachers.length" class="empty-state">
        暂无教师账号
      </div>

      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>账号</th>
              <th>角色</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>重置密码</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="teacher in teachers" :key="teacher.id">
              <td>{{ teacher.username }}</td>
              <td>{{ teacher.role }}</td>
              <td>{{ teacher.status }}</td>
              <td>{{ teacher.createdAt }}</td>
              <td>
                <div class="stack">
                  <input
                    v-model="resetPasswords[teacher.id]"
                    class="input"
                    :name="`reset-password-${teacher.id}`"
                    type="password"
                    autocomplete="new-password"
                    placeholder="输入新密码"
                  />
                  <button
                    class="button button--ghost"
                    type="button"
                    :disabled="directoryStore.saving"
                    @click="submitResetPassword(teacher.id)"
                  >
                    重置密码
                  </button>
                </div>
              </td>
              <td>
                <button
                  v-if="teacher.role !== 'admin'"
                  class="button button--ghost"
                  type="button"
                  :data-testid="teacher.status === 'disabled' ? `teacher-enable-${teacher.id}` : `teacher-disable-${teacher.id}`"
                  :disabled="directoryStore.saving"
                  @click="toggleTeacherStatus(teacher.id, teacher.status)"
                >
                  {{ teacher.status === 'disabled' ? '启用' : '禁用' }}
                </button>
                <span v-else class="cell-subtle">系统保留账号</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </AppShell>
</template>
