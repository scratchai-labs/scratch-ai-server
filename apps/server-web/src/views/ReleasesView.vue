<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppShell from '@/components/AppShell.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import {
  useTeacherApiClient,
  type TeacherRelease,
  type TeacherReleaseAnalysis,
  type TeacherReleaseAssignedStudent,
  type TeacherReleaseDetail,
  type TeacherReleaseMutationResult,
  type TeacherReleaseStatus,
  type TeacherStudent,
} from '@/services/teacherApi'
import { toErrorMessage } from '@/stores/storeUtils'
import { useTeacherDirectoryStore } from '@/stores/teacherDirectory'

const apiClient = useTeacherApiClient()
const directoryStore = useTeacherDirectoryStore()

const createForm = reactive({
  title: '',
  goal: '',
  description: '',
  file: null as File | null,
})

const createError = ref<string | null>(null)
const createFeedback = ref('')
const createLoading = ref(false)
const detailError = ref<string | null>(null)
const detailFeedback = ref('')
const detailLoading = ref(false)
const detailSaving = ref(false)
const fileInputKey = ref(0)
const selectedReleaseId = ref('')
const selectedReleaseDetail = ref<TeacherReleaseDetail | null>(null)
const selectedReleaseAnalysis = ref<TeacherReleaseAnalysis | null>(null)
const students = ref<TeacherStudent[]>([])
const assignedStudentIds = ref<string[]>([])

const releases = computed(() => directoryStore.releases)
const selectedReleaseTitle = computed(() => selectedReleaseDetail.value?.title ?? '发布单详情')

function releaseTone(status: TeacherReleaseStatus) {
  if (status === 'published') {
    return 'success'
  }

  if (status === 'draft') {
    return 'warning'
  }

  return 'muted'
}

function releaseLabel(status: TeacherReleaseStatus) {
  if (status === 'published') {
    return '已发布'
  }

  if (status === 'draft') {
    return '草稿'
  }

  return '已归档'
}

function analysisLabel(status: string) {
  if (status === 'ready') {
    return '分析完成'
  }

  if (status === 'pending') {
    return '分析中'
  }

  if (status === 'failed') {
    return '分析失败'
  }

  return status || '未知'
}

function analysisTone(status: string) {
  if (status === 'ready') {
    return 'success'
  }

  if (status === 'pending') {
    return 'warning'
  }

  if (status === 'failed') {
    return 'danger'
  }

  return 'muted'
}

function compareStudentsByName(left: TeacherStudent, right: TeacherStudent) {
  return left.name.localeCompare(right.name, 'zh-Hans-CN')
}

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  createForm.file = target.files?.[0] ?? null
}

function normalizeAssignedStudents(selectedIds: string[]): TeacherReleaseAssignedStudent[] {
  return students.value
    .filter((student) => selectedIds.includes(student.id))
    .map((student) => ({
      id: student.id,
      username: student.username,
      displayName: student.name,
      status: student.status || 'assigned',
    }))
}

function mergeReleaseMutation(result: TeacherReleaseMutationResult) {
  directoryStore.releases = directoryStore.releases.map((release) =>
    release.id === result.id
      ? {
          ...release,
          status: result.status,
          analysisStatus: result.analysisStatus,
        }
      : release,
  )
}

async function reloadReleases() {
  await directoryStore.loadReleases(apiClient)
}

async function loadReleaseDetail(releaseId: string) {
  detailLoading.value = true
  detailError.value = null
  detailFeedback.value = ''

  try {
    const [detail, analysis, nextStudents] = await Promise.all([
      apiClient.getReleaseDetail?.(releaseId),
      apiClient.getReleaseAnalysis?.(releaseId),
      apiClient.listStudents(),
    ])

    if (!detail) {
      throw new Error('发布单详情接口未提供')
    }
    if (!analysis) {
      throw new Error('发布单分析接口未提供')
    }

    selectedReleaseId.value = releaseId
    selectedReleaseDetail.value = detail
    selectedReleaseAnalysis.value = analysis
    students.value = [...nextStudents].sort(compareStudentsByName)
    assignedStudentIds.value = detail.assignedStudents.map((student) => student.id)
  } catch (error) {
    detailError.value = toErrorMessage(error, '发布单详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

async function submitCreateRelease() {
  if (
    !createForm.title.trim()
    || !createForm.goal.trim()
    || !createForm.description.trim()
    || !createForm.file
  ) {
    return
  }

  createLoading.value = true
  createError.value = null
  createFeedback.value = ''

  try {
    const createdRelease = await apiClient.createRelease?.({
      title: createForm.title.trim(),
      goal: createForm.goal.trim(),
      description: createForm.description.trim(),
      file: createForm.file,
    })

    if (!createdRelease) {
      throw new Error('发布单创建接口未提供')
    }

    createFeedback.value = `已上传发布单 ${createdRelease.title}`
    createForm.title = ''
    createForm.goal = ''
    createForm.description = ''
    createForm.file = null
    fileInputKey.value += 1
    await reloadReleases()
    await loadReleaseDetail(createdRelease.id)
  } catch (error) {
    createError.value = toErrorMessage(error, '发布单上传失败')
  } finally {
    createLoading.value = false
  }
}

async function assignStudents(releaseId: string) {
  detailSaving.value = true
  detailError.value = null
  detailFeedback.value = ''

  try {
    const result = await apiClient.assignStudentsToRelease?.(
      releaseId,
      [...assignedStudentIds.value],
    )

    if (!result) {
      throw new Error('发布单分配接口未提供')
    }

    if (selectedReleaseDetail.value?.id === releaseId) {
      selectedReleaseDetail.value = {
        ...selectedReleaseDetail.value,
        assignedStudents: normalizeAssignedStudents(result.studentIds),
      }
    }

    detailFeedback.value = `已分配 ${result.assignedCount} 名学生`
  } catch (error) {
    detailError.value = toErrorMessage(error, '学生分配失败')
  } finally {
    detailSaving.value = false
  }
}

async function changeReleaseStatus(
  releaseId: string,
  action: 'publish' | 'archive',
) {
  detailSaving.value = true
  detailError.value = null
  detailFeedback.value = ''

  try {
    const result =
      action === 'publish'
        ? await apiClient.publishRelease?.(releaseId)
        : await apiClient.archiveRelease?.(releaseId)

    if (!result) {
      throw new Error(action === 'publish' ? '发布接口未提供' : '归档接口未提供')
    }

    mergeReleaseMutation(result)

    if (selectedReleaseDetail.value?.id === releaseId) {
      selectedReleaseDetail.value = {
        ...selectedReleaseDetail.value,
        status: result.status,
        analysisStatus: result.analysisStatus,
      }
    }

    detailFeedback.value = action === 'publish'
      ? `已发布 ${result.title}`
      : `已归档 ${result.title}`
  } catch (error) {
    detailError.value = toErrorMessage(
      error,
      action === 'publish' ? '发布失败' : '归档失败',
    )
  } finally {
    detailSaving.value = false
  }
}

onMounted(() => {
  void reloadReleases()
})
</script>

<template>
  <AppShell
    title="发布单管理"
    description="把上传、详情、分配和状态流转都收在一页，方便老师直接操作。"
  >
    <template #actions>
      <StatusBadge :tone="directoryStore.releasesLoading ? 'warning' : 'success'">
        {{ directoryStore.releasesLoading ? '加载中' : `${directoryStore.releaseCount} 个发布单` }}
      </StatusBadge>
      <button class="button button--ghost" type="button" :disabled="directoryStore.releasesLoading" @click="reloadReleases">
        刷新发布单
      </button>
    </template>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">上传参考 sb3</h2>
          <p class="panel__meta">创建草稿发布单后，会自动拉取详情和分析摘要，方便继续分配学生与发布。</p>
        </div>
      </div>

      <form class="form-grid" data-testid="create-release-form" @submit.prevent="submitCreateRelease">
        <label class="field">
          <span>发布单标题</span>
          <input
            v-model="createForm.title"
            class="input"
            name="release-title"
            placeholder="迷宫任务"
          />
        </label>

        <label class="field">
          <span>教学目标</span>
          <input
            v-model="createForm.goal"
            class="input"
            name="release-goal"
            placeholder="让角色移动起来"
          />
        </label>

        <label class="field">
          <span>任务说明</span>
          <textarea
            v-model="createForm.description"
            class="input"
            name="release-description"
            rows="4"
            placeholder="第一课任务"
          />
        </label>

        <label class="field">
          <span>参考 sb3</span>
          <input
            :key="fileInputKey"
            class="input"
            name="release-sb3"
            type="file"
            accept=".sb3,application/zip"
            @change="onFileChange"
          />
        </label>

        <button class="button button--primary" type="submit" :disabled="createLoading">
          上传并创建发布单
        </button>
      </form>

      <p v-if="createError" role="alert" class="feedback feedback--error">
        {{ createError }}
      </p>
      <p v-else-if="createFeedback" role="status" class="feedback feedback--success">
        {{ createFeedback }}
      </p>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">发布单列表</h2>
          <p class="panel__meta">查看详情后，可继续完成学生分配、发布和归档；实时看板入口仍保留在卡片里。</p>
        </div>
      </div>

      <p v-if="directoryStore.releasesError" role="alert" class="feedback feedback--error">
        {{ directoryStore.releasesError }}
      </p>

      <div v-if="!directoryStore.releasesLoading && !releases.length" class="empty-state">
        暂无发布单数据
      </div>

      <div v-else class="card-grid">
        <article v-for="release in releases" :key="release.id" class="release-card">
          <div class="release-card__head">
            <div>
              <h2>{{ release.title }}</h2>
              <p>{{ release.className }}</p>
            </div>
            <StatusBadge :tone="releaseTone(release.status)">
              {{ releaseLabel(release.status) }}
            </StatusBadge>
          </div>

          <dl class="release-card__meta">
            <div>
              <dt>学生数</dt>
              <dd>{{ release.studentCount }}</dd>
            </div>
            <div>
              <dt>分析状态</dt>
              <dd>{{ analysisLabel(release.analysisStatus) }}</dd>
            </div>
            <div>
              <dt>更新时间</dt>
              <dd>{{ release.updatedAt }}</dd>
            </div>
          </dl>

          <div class="stack">
            <button
              class="button button--ghost"
              type="button"
              :data-testid="`view-release-${release.id}`"
              :disabled="detailLoading"
              @click="loadReleaseDetail(release.id)"
            >
              查看详情
            </button>
            <RouterLink class="button button--primary" :to="`/releases/${release.id}/live`">
              查看实时看板
            </RouterLink>
          </div>
        </article>
      </div>
    </section>

    <section class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">{{ selectedReleaseTitle }}</h2>
          <p class="panel__meta">这里展示任务详情、分析摘要与学生分配结果，方便老师快速完成发布流转。</p>
        </div>
        <StatusBadge
          v-if="selectedReleaseDetail"
          :tone="releaseTone(selectedReleaseDetail.status)"
        >
          {{ releaseLabel(selectedReleaseDetail.status) }}
        </StatusBadge>
      </div>

      <p v-if="detailError" role="alert" class="feedback feedback--error">
        {{ detailError }}
      </p>
      <p v-else-if="detailFeedback" role="status" class="feedback feedback--success">
        {{ detailFeedback }}
      </p>

      <div v-if="detailLoading" class="empty-state">
        正在加载发布单详情
      </div>

      <div v-else-if="selectedReleaseDetail && selectedReleaseAnalysis" class="stack">
        <div class="table-wrap">
          <table class="data-table">
            <tbody>
              <tr>
                <th>标题</th>
                <td>{{ selectedReleaseDetail.title }}</td>
              </tr>
              <tr>
                <th>目标</th>
                <td>{{ selectedReleaseDetail.goal }}</td>
              </tr>
              <tr>
                <th>说明</th>
                <td>{{ selectedReleaseDetail.description }}</td>
              </tr>
              <tr>
                <th>分析状态</th>
                <td>
                  <StatusBadge :tone="analysisTone(selectedReleaseAnalysis.analysisStatus)">
                    {{ analysisLabel(selectedReleaseAnalysis.analysisStatus) }}
                  </StatusBadge>
                </td>
              </tr>
              <tr>
                <th>更新时间</th>
                <td>{{ selectedReleaseDetail.updatedAt }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="stack">
          <div>
            <strong>教学提示</strong>
            <ul>
              <li v-for="point in selectedReleaseAnalysis.teachingPoints" :key="point">
                {{ point }}
              </li>
            </ul>
          </div>

          <div>
            <strong>角色</strong>
            <ul>
              <li v-for="roleName in selectedReleaseAnalysis.roleNames" :key="roleName">
                {{ roleName }}
              </li>
            </ul>
          </div>

          <div>
            <strong>广播 / 变量 / 扩展</strong>
            <ul>
              <li>广播：{{ selectedReleaseAnalysis.broadcastMessages.join('、') || '无' }}</li>
              <li>变量：{{ selectedReleaseAnalysis.variableNames.join('、') || '无' }}</li>
              <li>列表：{{ selectedReleaseAnalysis.listNames.join('、') || '无' }}</li>
              <li>扩展：{{ selectedReleaseAnalysis.extensions.join('、') || '无' }}</li>
            </ul>
          </div>
        </div>

        <div class="stack">
          <div>
            <strong>已分配学生</strong>
            <p class="cell-subtle">
              {{ selectedReleaseDetail.assignedStudents.length ? `${selectedReleaseDetail.assignedStudents.length} 名学生已绑定` : '当前还没有分配学生' }}
            </p>
            <ul v-if="selectedReleaseDetail.assignedStudents.length">
              <li v-for="student in selectedReleaseDetail.assignedStudents" :key="student.id">
                {{ student.displayName }}（{{ student.username }}）
              </li>
            </ul>
          </div>

          <div>
            <strong>分配学生</strong>
            <div class="stack">
              <label v-for="student in students" :key="student.id">
                <input
                  v-model="assignedStudentIds"
                  :name="`assign-student-${student.id}`"
                  type="checkbox"
                  :value="student.id"
                />
                {{ student.name }}（{{ student.username }}）
              </label>
            </div>
          </div>

          <div class="stack">
            <button
              class="button button--primary"
              type="button"
              :data-testid="`assign-release-${selectedReleaseId}`"
              :disabled="detailSaving"
              @click="assignStudents(selectedReleaseId)"
            >
              保存分配
            </button>
            <button
              class="button button--ghost"
              type="button"
              :data-testid="`publish-release-${selectedReleaseId}`"
              :disabled="detailSaving"
              @click="changeReleaseStatus(selectedReleaseId, 'publish')"
            >
              发布发布单
            </button>
            <button
              class="button button--ghost"
              type="button"
              :data-testid="`archive-release-${selectedReleaseId}`"
              :disabled="detailSaving"
              @click="changeReleaseStatus(selectedReleaseId, 'archive')"
            >
              归档发布单
            </button>
          </div>
        </div>
      </div>

      <div v-else class="empty-state">
        先从发布单列表选择一个任务查看详情
      </div>
    </section>
  </AppShell>
</template>
