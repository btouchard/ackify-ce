// SPDX-License-Identifier: AGPL-3.0-or-later
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const SignPage = () => import('@/pages/SignPage.vue')
const SignaturesPage = () => import('@/pages/SignaturesPage.vue')
const AdminDashboard = () => import('@/pages/admin/AdminDashboard.vue')
const AdminDocumentDetail = () => import('@/pages/admin/AdminDocumentDetail.vue')
const EmbedPage = () => import('@/pages/EmbedPage.vue')
const NotFoundPage = () => import('@/pages/NotFoundPage.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'sign',
    component: SignPage,
    meta: { requiresAuth: false }
  },
  {
    path: '/signatures',
    name: 'signatures',
    component: SignaturesPage,
    meta: { requiresAuth: true }
  },
  {
    path: '/admin',
    name: 'admin',
    component: AdminDashboard,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/admin/docs/:docId',
    name: 'admin-document',
    component: AdminDocumentDetail,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/embed',
    name: 'embed',
    component: EmbedPage,
    meta: { requiresAuth: false, isEmbed: true }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: NotFoundPage
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    return { top: 0, behavior: 'smooth' }
  }
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  try {
    // Only check auth if the route requires it
    if (to.meta.requiresAuth || to.meta.requiresAdmin) {
      if (!authStore.initialized) {
        await authStore.checkAuth()
      }

      if (!authStore.isAuthenticated) {
        sessionStorage.setItem('redirectAfterLogin', to.fullPath)
        await authStore.startOAuthLogin(to.fullPath)
        return false
      }

      if (to.meta.requiresAdmin && !authStore.isAdmin) {
        next({ name: 'sign' })
        return
      }
    }

    if (from.path === '/api/v1/auth/callback') {
      const redirectPath = sessionStorage.getItem('redirectAfterLogin')
      if (redirectPath) {
        sessionStorage.removeItem('redirectAfterLogin')
        next(redirectPath)
        return
      }
    }

    next()
  } catch (error) {
    console.error('Navigation guard error:', error)
    next()
  }
})

export default router