import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import type { Pinia } from 'pinia'
import { useSessionStore } from '@/stores/session'
import AdminOverviewView from '@/views/AdminOverviewView.vue'
import AdminAuditLogsView from '@/views/AdminAuditLogsView.vue'
import AdminStudentsView from '@/views/AdminStudentsView.vue'
import AdminTeachersView from '@/views/AdminTeachersView.vue'
import ClassDetailView from '@/views/ClassDetailView.vue'
import ClassesView from '@/views/ClassesView.vue'
import DashboardView from '@/views/DashboardView.vue'
import LoginView from '@/views/LoginView.vue'
import ProjectDetailView from '@/views/ProjectDetailView.vue'

export function createTeacherRouter(pinia: Pinia) {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: '/',
        redirect: '/classes',
      },
      {
        path: '/login',
        name: 'login',
        component: LoginView,
        meta: {
          publicRoute: true,
        },
      },
      {
        path: '/dashboard',
        name: 'dashboard',
        component: DashboardView,
        meta: {
          requiresTeacher: true,
        },
      },
      {
        path: '/classes',
        name: 'classes',
        component: ClassesView,
        meta: {
          requiresTeacher: true,
        },
      },
      {
        path: '/classes/:id',
        name: 'class-detail',
        component: ClassDetailView,
        props: true,
        meta: {
          requiresTeacher: true,
        },
      },
      {
        path: '/projects/:id',
        name: 'project-detail',
        component: ProjectDetailView,
        props: true,
        meta: {
          requiresTeacher: true,
        },
      },
      {
        path: '/admin',
        name: 'admin-overview',
        component: AdminOverviewView,
        meta: {
          requiresAdmin: true,
        },
      },
      {
        path: '/admin/teachers',
        name: 'admin-teachers',
        component: AdminTeachersView,
        meta: {
          requiresAdmin: true,
        },
      },
      {
        path: '/admin/students',
        name: 'admin-students',
        component: AdminStudentsView,
        meta: {
          requiresAdmin: true,
        },
      },
      {
        path: '/admin/audit-logs',
        name: 'admin-audit-logs',
        component: AdminAuditLogsView,
        meta: {
          requiresAdmin: true,
        },
      },
      {
        path: '/:pathMatch(.*)*',
        redirect: '/classes',
      },
    ],
  })

  router.beforeEach((to: RouteLocationNormalized) => {
    const sessionStore = useSessionStore(pinia)

    if (to.meta.publicRoute) {
      if (sessionStore.isAuthenticated) {
        return sessionStore.landingPath
      }
      return true
    }

    if (!sessionStore.isAuthenticated) {
      return {
        path: '/login',
        query: {
          redirect: to.fullPath,
        },
      }
    }

    if (to.meta.requiresAdmin && !sessionStore.isAdmin) {
      return '/classes'
    }

    if (to.meta.requiresTeacher && sessionStore.isAdmin) {
      return '/admin'
    }

    return true
  })

  return router
}
