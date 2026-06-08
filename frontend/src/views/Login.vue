<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-tr from-slate-900 via-sky-950 to-slate-900 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8 bg-slate-900/60 backdrop-blur-xl border border-slate-700/50 p-8 rounded-2xl shadow-2xl">
      <!-- Title -->
      <div class="text-center">
        <i class="pi pi-shield text-sky-400 text-5xl mb-2"></i>
        <h2 class="text-3xl font-extrabold text-white tracking-tight">ERP Veterinario</h2>
        <p class="mt-2 text-sm text-slate-400">
          Administración Multisucursal & Facturación
        </p>
      </div>

      <!-- Tabs Toggle -->
      <div class="flex border-b border-slate-700/50">
        <button
          @click="activeTab = 'login'"
          :class="['w-1/2 py-2.5 text-sm font-medium transition-colors border-b-2', 
            activeTab === 'login' ? 'border-sky-500 text-sky-400' : 'border-transparent text-slate-400 hover:text-slate-200']"
        >
          Iniciar Sesión
        </button>
        <button
          @click="activeTab = 'register'"
          :class="['w-1/2 py-2.5 text-sm font-medium transition-colors border-b-2', 
            activeTab === 'register' ? 'border-sky-500 text-sky-400' : 'border-transparent text-slate-400 hover:text-slate-200']"
        >
          Registrar Empresa
        </button>
      </div>

      <!-- Error Message -->
      <div v-if="error || authStore.error" class="bg-red-500/10 border border-red-500/50 text-red-200 p-3.5 rounded-lg text-sm flex items-center gap-2">
        <i class="pi pi-exclamation-triangle"></i>
        <span>{{ error || authStore.error }}</span>
      </div>

      <!-- Success Message -->
      <div v-if="successMsg" class="bg-emerald-500/10 border border-emerald-500/50 text-emerald-200 p-3.5 rounded-lg text-sm flex items-center gap-2">
        <i class="pi pi-check-circle"></i>
        <span>{{ successMsg }}</span>
      </div>

      <!-- Tab Content: LOGIN -->
      <form v-if="activeTab === 'login'" class="mt-8 space-y-6" @submit.prevent="handleLogin">
        <div class="rounded-md space-y-4">
          <div>
            <label for="username" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">Usuario o Correo</label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-400">
                <i class="pi pi-user"></i>
              </span>
              <input
                id="username"
                v-model="loginData.username"
                type="text"
                required
                class="block w-full pl-10 pr-3 py-2.5 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-500 transition-all text-sm"
                placeholder="Ej. admin"
              />
            </div>
          </div>

          <div>
            <label for="password" class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">Contraseña</label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-400">
                <i class="pi pi-lock"></i>
              </span>
              <input
                id="password"
                v-model="loginData.password"
                type="password"
                required
                class="block w-full pl-10 pr-3 py-2.5 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-500 transition-all text-sm"
                placeholder="••••••••"
              />
            </div>
          </div>
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full flex justify-center py-2.5 px-4 border border-transparent rounded-lg shadow-sm text-sm font-semibold text-white bg-sky-600 hover:bg-sky-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500 focus:ring-offset-slate-900 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span v-if="authStore.loading"><i class="pi pi-spin pi-spinner mr-2"></i>Cargando...</span>
          <span v-else>Entrar</span>
        </button>

        <div class="bg-slate-800/40 p-3 rounded-lg border border-slate-700/30 text-center">
          <p class="text-xs text-slate-400">
            Credenciales de prueba:
            <code class="text-sky-300 font-mono ml-1">admin / admin123</code>
          </p>
        </div>
      </form>

      <!-- Tab Content: REGISTER -->
      <form v-else class="mt-8 space-y-4" @submit.prevent="handleRegister">
        <h3 class="text-xs font-bold text-sky-400 uppercase tracking-wider border-b border-slate-800 pb-1.5 mb-2">Datos de la Empresa</h3>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-slate-300 mb-1">RUC (11 dígitos)</label>
            <input
              v-model="regData.ruc"
              type="text"
              required
              maxlength="11"
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="Ej. 20123456789"
            />
          </div>
          <div>
            <label class="block text-xs text-slate-300 mb-1">Razón Social</label>
            <input
              v-model="regData.razon_social"
              type="text"
              required
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="Nombre legal"
            />
          </div>
        </div>

        <div>
          <label class="block text-xs text-slate-300 mb-1">Dirección Fiscal</label>
          <input
            v-model="regData.direccion"
            type="text"
            class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
            placeholder="Dirección comercial"
          />
        </div>

        <h3 class="text-xs font-bold text-sky-400 uppercase tracking-wider border-b border-slate-800 pb-1.5 mt-4 mb-2">Usuario Administrador</h3>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-slate-300 mb-1">Nombre Completo</label>
            <input
              v-model="regData.admin_name"
              type="text"
              required
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="Juan Perez"
            />
          </div>
          <div>
            <label class="block text-xs text-slate-300 mb-1">Correo Electrónico</label>
            <input
              v-model="regData.admin_email"
              type="email"
              required
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="juan@correo.com"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-slate-300 mb-1">Usuario de Acceso</label>
            <input
              v-model="regData.admin_username"
              type="text"
              required
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="ej. juan_admin"
            />
          </div>
          <div>
            <label class="block text-xs text-slate-300 mb-1">Contraseña</label>
            <input
              v-model="regData.admin_password"
              type="password"
              required
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="Mínimo 6 caracteres"
            />
          </div>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full mt-4 flex justify-center py-2.5 px-4 border border-transparent rounded-lg shadow-sm text-sm font-semibold text-white bg-teal-600 hover:bg-teal-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-teal-500 focus:ring-offset-slate-900 transition-all disabled:opacity-50"
        >
          <span v-if="loading"><i class="pi pi-spin pi-spinner mr-2"></i>Registrando...</span>
          <span v-else>Crear Cuenta e Inicializar ERP</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('login')
const error = ref('')
const successMsg = ref('')
const loading = ref(false)

const loginData = reactive({
  username: '',
  password: ''
})

const regData = reactive({
  ruc: '',
  razon_social: '',
  nombre_comercial: '',
  direccion: '',
  telefono: '',
  email: '',
  admin_name: '',
  admin_email: '',
  admin_username: '',
  admin_password: ''
})

async function handleLogin() {
  error.value = ''
  successMsg.value = ''
  const success = await authStore.login(loginData.username, loginData.password)
  if (success) {
    router.push('/')
  }
}

async function handleRegister() {
  error.value = ''
  successMsg.value = ''
  loading.value = true
  try {
    const response = await axios.post('/companies/register', regData)
    if (response.data.success) {
      successMsg.value = '¡Empresa y cuenta de administrador creadas con éxito! Por favor inicie sesión.'
      activeTab.value = 'login'
      loginData.username = regData.admin_username
      loginData.password = regData.admin_password
      // Reset form
      Object.assign(regData, {
        ruc: '',
        razon_social: '',
        nombre_comercial: '',
        direccion: '',
        telefono: '',
        email: '',
        admin_name: '',
        admin_email: '',
        admin_username: '',
        admin_password: ''
      })
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Error al registrar la empresa'
  } finally {
    loading.value = false
  }
}
</script>
