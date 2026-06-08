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
  roles?: Array<{ nombre: string }>
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null as User | null,
    loading: false,
    error: ''
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => {
      if (!state.user || !state.user.roles) return false
      return state.user.roles.some(role => role.nombre === 'Administrador')
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
    async fetchUser() {
      if (!this.token) return
      try {
        const response = await axios.get('/auth/me')
        if (response.data.success) {
          this.user = response.data.data
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
      delete axios.defaults.headers.common['Authorization']
      window.location.href = '/login'
    }
  }
})
