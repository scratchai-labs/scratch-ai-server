import { ref } from 'vue'
import { defineStore } from 'pinia'
import type {
  CreateManagedTeacherInput,
  ManagedTeacher,
  TeacherApiClient,
} from '@/services/teacherApi'
import { toErrorMessage } from './storeUtils'

export const useAdminTeacherDirectoryStore = defineStore('adminTeacherDirectory', () => {
  const teachers = ref<ManagedTeacher[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)
  const feedback = ref<string>('')

  async function loadTeachers(api: TeacherApiClient) {
    loading.value = true
    error.value = null

    try {
      teachers.value = await api.listTeachers?.() ?? []
    } catch (nextError) {
      error.value = toErrorMessage(nextError, '教师列表加载失败')
    } finally {
      loading.value = false
    }
  }

  async function createTeacher(api: TeacherApiClient, input: CreateManagedTeacherInput) {
    saving.value = true
    error.value = null

    try {
      const createdTeacher = await api.createTeacher?.(input)
      if (!createdTeacher) {
        throw new Error('教师创建接口未提供')
      }
      teachers.value = [...teachers.value, createdTeacher].sort(compareTeachersByCreatedAt)
      feedback.value = `已创建教师账号 ${createdTeacher.username}`
      return createdTeacher
    } catch (nextError) {
      error.value = toErrorMessage(nextError, '教师创建失败')
      throw nextError
    } finally {
      saving.value = false
    }
  }

  async function resetTeacherPassword(api: TeacherApiClient, teacherId: string, newPassword: string) {
    saving.value = true
    error.value = null

    try {
      const updatedTeacher = await api.resetTeacherPassword?.(teacherId, newPassword)
      if (!updatedTeacher) {
        throw new Error('教师密码重置接口未提供')
      }
      mergeTeacher(updatedTeacher)
      feedback.value = `已重置 ${updatedTeacher.username} 的密码`
      return updatedTeacher
    } catch (nextError) {
      error.value = toErrorMessage(nextError, '教师密码重置失败')
      throw nextError
    } finally {
      saving.value = false
    }
  }

  async function disableTeacher(api: TeacherApiClient, teacherId: string) {
    saving.value = true
    error.value = null

    try {
      const updatedTeacher = await api.disableTeacher?.(teacherId)
      if (!updatedTeacher) {
        throw new Error('教师禁用接口未提供')
      }
      mergeTeacher(updatedTeacher)
      feedback.value = `已禁用 ${updatedTeacher.username}`
      return updatedTeacher
    } catch (nextError) {
      error.value = toErrorMessage(nextError, '教师禁用失败')
      throw nextError
    } finally {
      saving.value = false
    }
  }

  async function enableTeacher(api: TeacherApiClient, teacherId: string) {
    saving.value = true
    error.value = null

    try {
      const updatedTeacher = await api.enableTeacher?.(teacherId)
      if (!updatedTeacher) {
        throw new Error('教师启用接口未提供')
      }
      mergeTeacher(updatedTeacher)
      feedback.value = `已启用 ${updatedTeacher.username}`
      return updatedTeacher
    } catch (nextError) {
      error.value = toErrorMessage(nextError, '教师启用失败')
      throw nextError
    } finally {
      saving.value = false
    }
  }

  function clearFeedback() {
    feedback.value = ''
  }

  function mergeTeacher(updatedTeacher: ManagedTeacher) {
    teachers.value = teachers.value
      .map((teacher) => (teacher.id === updatedTeacher.id ? updatedTeacher : teacher))
      .sort(compareTeachersByCreatedAt)
  }

  return {
    teachers,
    loading,
    saving,
    error,
    feedback,
    loadTeachers,
    createTeacher,
    resetTeacherPassword,
    disableTeacher,
    enableTeacher,
    clearFeedback,
  }
})

function compareTeachersByCreatedAt(left: ManagedTeacher, right: ManagedTeacher) {
  return Date.parse(left.createdAt) - Date.parse(right.createdAt)
}
