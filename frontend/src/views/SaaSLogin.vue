<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-tr from-slate-900 via-indigo-950 to-slate-900 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8 bg-slate-900/60 backdrop-blur-xl border border-indigo-500/30 p-8 rounded-3xl shadow-2xl">
      <!-- Title -->
      <div class="text-center">
        <div class="inline-flex w-16 h-16 bg-indigo-500/10 border border-indigo-500/30 text-indigo-400 rounded-2xl items-center justify-center mb-4 shadow-lg shadow-indigo-500/10">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-9 w-9" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
        <h2 class="text-2xl font-black text-white tracking-tight uppercase">SaaS Console</h2>
        <p class="mt-2 text-xs text-slate-400 font-bold uppercase tracking-widest">
          Administración Global de Inquilinos
        </p>
      </div>

      <!-- Warnings/Errors -->
      <div v-if="error || authStore.error" class="bg-red-500/10 border border-red-500/30 text-red-200 p-4 rounded-2xl text-xs font-bold flex items-center gap-2">
        <i class="pi pi-exclamation-triangle"></i>
        <span>{{ error || authStore.error }}</span>
      </div>

      <!-- Login Form -->
      <form class="mt-8 space-y-5" @submit.prevent="handleLogin">
        <div class="space-y-4">
          <div>
            <label for="saas_username" class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-2">Usuario Administrador</label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400">
                <i class="pi pi-user"></i>
              </span>
              <input
                id="saas_username"
                v-model="loginData.username"
                type="text"
                required
                class="block w-full pl-11 pr-4 py-3.5 bg-slate-800/60 border border-slate-700/60 text-white rounded-2xl focus:outline-none focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500/50 transition-all text-sm font-bold placeholder-slate-500"
                placeholder="Ingrese su usuario"
              />
            </div>
          </div>

          <div>
            <label for="saas_password" class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-2">Contraseña</label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400">
                <i class="pi pi-lock"></i>
              </span>
              <input
                id="saas_password"
                v-model="loginData.password"
                type="password"
                required
                class="block w-full pl-11 pr-4 py-3.5 bg-slate-800/60 border border-slate-700/60 text-white rounded-2xl focus:outline-none focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500/50 transition-all text-sm font-bold placeholder-slate-500"
                placeholder="••••••••"
              />
            </div>
          </div>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full mt-6 py-4 bg-indigo-600 hover:bg-indigo-500 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-indigo-500/10 hover:shadow-indigo-500/20 active:scale-[0.98] transition-all disabled:opacity-50"
        >
          <span v-if="loading"><i class="pi pi-spin pi-spinner mr-2"></i>Iniciando sesión...</span>
          <span v-else>Acceder a la Consola</span>
        </button>
      </form>

      <div class="text-center pt-4">
        <a href="/login" class="text-[10px] text-slate-500 hover:text-slate-300 font-bold uppercase tracking-widest transition-colors">Volver al portal de clientes</a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const router = useRouter()
const authStore = useAuthStore()

const error = ref('')
const loading = ref(false)

const loginData = reactive({
  username: '',
  password: ''
})

async function handleLogin() {
  error.value = ''
  loading.value = true
  const success = await authStore.login(loginData.username, loginData.password)
  loading.value = false
  
  if (success) {
    if (authStore.isSuperAdmin) {
      router.push('/saas-admin')
    } else {
      // Force exit if non-superadmin logs in here
      authStore.logout()
      error.value = 'Acceso denegado: esta consola es de uso exclusivo para SuperAdmin.'
    }
  }
}
</script>
