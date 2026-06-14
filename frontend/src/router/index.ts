import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import SaaSAdmin from '../views/SaaSAdmin.vue'
import SaaSLogin from '../views/SaaSLogin.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { guest: true }
  },
  {
    path: '/saas-admin/login',
    name: 'SaaSLogin',
    component: SaaSLogin,
    meta: { guest: true }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/saas-admin',
    name: 'SaaSAdmin',
    component: SaaSAdmin,
    meta: { requiresAuth: true, requiresSuperAdmin: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation Guard
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const user = JSON.parse(localStorage.getItem('user') || 'null')
  const isSuperAdmin = user && user.role_type === 'SUPER_ADMIN'

  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!token) {
      if (to.matched.some(record => record.meta.requiresSuperAdmin)) {
        next('/saas-admin/login')
      } else {
        next('/login')
      }
    } else {
      // Role checking and Isolation
      if (isSuperAdmin) {
        // SuperAdmin should ONLY be in /saas-admin
        if (to.path !== '/saas-admin') {
          next('/saas-admin')
        } else {
          next()
        }
      } else {
        // Regular users should NEVER be in /saas-admin
        if (to.matched.some(record => record.meta.requiresSuperAdmin)) {
          next('/')
        } else {
          next()
        }
      }
    }
  } else if (to.matched.some(record => record.meta.guest)) {
    if (token) {
      if (isSuperAdmin) {
        next('/saas-admin')
      } else {
        if (to.path === '/saas-admin/login') {
          next()
        } else {
          next('/')
        }
      }
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
