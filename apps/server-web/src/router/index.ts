import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import type { Pinia } from 'pinia'
import { useSessionStore } from '@/stores/session'
import AdminOverviewView from '@/views/AdminOverviewView.vue'
import AdminStudentsView from '@/views/AdminStudentsView.vue'
import AdminTeachersView from '@/views/AdminTeachersView.vue'
import DashboardView from '@/views/DashboardView.vue'
import LiveReleaseView from '@/views/LiveReleaseView.vue'
import LoginView from '@/views/LoginView.vue'
import ReleasesView from '@/views/ReleasesView.vue'
import StudentsView from '@/views/StudentsView.vue'

export function createTeacherRouter(pinia: Pinia) {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: '/',
        redirect: '/dashboard',
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
        path: '/students',
        name: 'students',
        component: StudentsView,
        meta: {
          requiresTeacher: true,
        },
      },
      {
        path: '/releases',
        name: 'releases',
        component: ReleasesView,
        meta: {
          requiresTeacher: true,
        },
      },
      {
        path: '/releases/:id/live',
        name: 'release-live',
        component: LiveReleaseView,
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
        path: '/:pathMatch(.*)*',
        redirect: '/dashboard',
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
      return '/dashboard'
    }

    if (to.meta.requiresTeacher && sessionStore.isAdmin) {
      return '/admin'
    }

    return true
  })

  return router
}
