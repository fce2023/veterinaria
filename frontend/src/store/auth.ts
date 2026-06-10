import { defineStore } from 'pinia'
import axios from 'axios'

// Set axios default configuration
const apiBase = import.meta.env.VITE_API_BASE_URL || '/api/v1'
axios.defaults.baseURL = apiBase

// Add authorization header if token exists
const token = localStorage.getItem('token')
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
}

export interface User {
  id: string
  nombre: string
  email: string
  username: string
  company_id: string
  branch_id: string
  role_type: string
  roles?: Array<{ nombre: string }>
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null') as User | null,
    loading: false,
    error: ''
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    isSuperAdmin: (state) => state.user?.role_type === 'SUPER_ADMIN',
    isCompanyAdmin: (state) => state.user?.role_type === 'COMPANY_ADMIN',
    isBranchAdmin: (state) => state.user?.role_type === 'BRANCH_ADMIN',
    isCashier: (state) => state.user?.role_type === 'BRANCH_USER' || state.user?.role_type === 'CASHIER',
    isAdmin: (state) => {
      if (!state.user) return false
      return state.user.role_type === 'SUPER_ADMIN' || state.user.role_type === 'COMPANY_ADMIN'
    }
  },
  actions: {
    async login(username: string, password: string): Promise<boolean> {
      this.loading = true
      this.error = ''
      try {
        const response = await axios.post('/auth/login', { username, password })
        if (response.data.success) {
          const { token, user } = response.data.data
          this.token = token
          this.user = user
          localStorage.setItem('token', token)
          localStorage.setItem('user', JSON.stringify(user))
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
          this.loading = false
          return true
        } else {
          this.error = response.data.error || 'Fallo al iniciar sesión'
          this.loading = false
          return false
        }
      } catch (err: any) {
        this.loading = false
        this.error = err.response?.data?.error || 'Credenciales inválidas o error de conexión'
        return false
      }
    },
    async switchBranch(branchId: string) {
      try {
        const response = await axios.post('/auth/switch-branch', { branch_id: branchId })
        if (response.data.success) {
          const { token, user } = response.data.data
          this.token = token
          this.user = user
          localStorage.setItem('token', token)
          localStorage.setItem('user', JSON.stringify(user))
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
          // Reload page to refresh all data with new branch context
          window.location.reload()
          return true
        }
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Error al cambiar de sucursal'
        return false
      }
    },
    async fetchUser() {
      if (!this.token) return
      try {
        const response = await axios.get('/auth/me')
        if (response.data.success) {
          this.user = response.data.data
          localStorage.setItem('user', JSON.stringify(this.user))
        } else {
          this.logout()
        }
      } catch (err) {
        this.logout()
      }
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      delete axios.defaults.headers.common['Authorization']
      window.location.href = '/login'
    }
  }
})
