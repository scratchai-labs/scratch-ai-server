<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { TeacherClassroomDetail } from '@/services/teacherApi'

defineProps<{
  classroom: TeacherClassroomDetail | null
  classroomId: string
  refreshClassroom?: () => Promise<void>
}>()
</script>

<template>
  <section class="quick-links">
    <article class="quick-link-card">
      <div class="stack">
        <h2>学生工作区</h2>
        <p>单个创建、批量导入和学生列表拆成独立节奏，适合老师先处理账号，再回头检查列表。</p>
      </div>
      <RouterLink class="button button--primary" :to="`/classes/${classroomId}/students`">进入学生页</RouterLink>
    </article>

    <article class="quick-link-card">
      <div class="stack">
        <h2>项目工作区</h2>
        <p>项目标题、目标、说明和 <code>sb3</code> 上传统一放在项目页，列表与创建表单不再紧贴在一起。</p>
      </div>
      <RouterLink class="button button--primary" :to="`/classes/${classroomId}/projects`">进入项目页</RouterLink>
    </article>
  </section>

  <section class="section-grid section-grid--aside">
    <div class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">班级摘要</h2>
          <p class="panel__meta">概览页只保留班级层级的信息，不直接堆表单。</p>
        </div>
      </div>

      <dl class="release-card__meta">
        <div>
          <dt>班级名称</dt>
          <dd>{{ classroom?.name || '—' }}</dd>
        </div>
        <div>
          <dt>学生数量</dt>
          <dd>{{ classroom?.studentCount ?? 0 }}</dd>
        </div>
        <div>
          <dt>项目数量</dt>
          <dd>{{ classroom?.projectCount ?? 0 }}</dd>
        </div>
        <div>
          <dt>创建时间</dt>
          <dd>{{ classroom?.createdAt || '—' }}</dd>
        </div>
      </dl>
    </div>

    <div class="panel">
      <div class="panel__head">
        <div>
          <h2 class="panel__title">建议操作顺序</h2>
          <p class="panel__meta">把老师最常用的动作拆成清晰的两步，避免跨区来回滚动。</p>
        </div>
      </div>

      <div class="section-stack">
        <article class="quick-link-card">
          <h2>1. 先处理学生</h2>
          <p>如果要新开课或补录，先进入学生页完成创建或批量导入，再检查学生列表是否完整。</p>
        </article>
        <article class="quick-link-card">
          <h2>2. 再处理项目</h2>
          <p>确认学生后，再到项目页创建任务、上传 <code>sb3</code>，最后从项目列表进入详情。</p>
        </article>
      </div>
    </div>
  </section>
</template>
