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
            <div class="relative flex rounded-lg shadow-sm">
              <input
                v-model="regData.ruc"
                type="text"
                required
                maxlength="11"
                class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-l-lg text-xs focus:ring-sky-500/50 pr-8"
                placeholder="Ej. 20123456789"
              />
              <button
                type="button"
                @click="buscarRUC"
                :disabled="searchingRuc || regData.ruc.length !== 11"
                class="inline-flex items-center px-2.5 py-1.5 border border-l-0 border-slate-700/80 rounded-r-lg bg-slate-800 text-slate-300 text-xs font-medium hover:bg-slate-700 focus:outline-none focus:ring-1 focus:ring-sky-500 disabled:opacity-50 transition-colors"
                title="Consultar RUC"
              >
                <i v-if="searchingRuc" class="pi pi-spin pi-spinner text-sky-400"></i>
                <i v-else class="pi pi-search text-sky-400"></i>
              </button>
            </div>
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

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-slate-300 mb-1">Dirección Fiscal</label>
            <input
              v-model="regData.direccion"
              type="text"
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
              placeholder="Dirección comercial"
            />
            
            <!-- Conditional Selector for Locales Anexos -->
            <div v-if="localesAnexos.length > 0" class="mt-1.5">
              <label class="block text-[10px] text-slate-400 mb-0.5">Establecimientos anexos detectados:</label>
              <select
                @change="seleccionarDireccionAnexa"
                class="block w-full px-2 py-1 bg-slate-800/50 border border-slate-750 text-slate-300 rounded text-[11px] focus:ring-sky-500/50"
              >
                <option value="">-- Usar Dirección Principal --</option>
                <option
                  v-for="(loc, index) in localesAnexos"
                  :key="index"
                  :value="index"
                >
                  {{ loc.direccion }} ({{ loc.distrito }})
                </option>
              </select>
            </div>
          </div>
          <div>
            <label class="block text-xs text-slate-300 mb-1">Rubro de Negocio (Módulo Demo)</label>
            <select
              v-model="regData.sector"
              class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-lg text-xs focus:ring-sky-500/50"
            >
              <option value="core">Comercio General (Core)</option>
              <option value="veterinaria">Veterinaria (Mascotas & HC)</option>
              <option value="vidrieria">Vidriería & Aluminio</option>
            </select>
          </div>
        </div>

        <h3 class="text-xs font-bold text-sky-400 uppercase tracking-wider border-b border-slate-800 pb-1.5 mt-4 mb-2">Usuario Administrador</h3>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-slate-300 mb-1">DNI (8 dígitos)</label>
            <div class="relative flex rounded-lg shadow-sm">
              <input
                v-model="regData.admin_dni"
                type="text"
                maxlength="8"
                class="block w-full px-3 py-2 bg-slate-800/80 border border-slate-700/80 text-white rounded-l-lg text-xs focus:ring-sky-500/50 pr-8"
                placeholder="Ej. 12345678"
              />
              <button
                type="button"
                @click="buscarDNI"
                :disabled="searchingDni || regData.admin_dni.length !== 8"
                class="inline-flex items-center px-2.5 py-1.5 border border-l-0 border-slate-700/80 rounded-r-lg bg-slate-800 text-slate-300 text-xs font-medium hover:bg-slate-700 focus:outline-none focus:ring-1 focus:ring-sky-500 disabled:opacity-50 transition-colors"
                title="Consultar DNI"
              >
                <i v-if="searchingDni" class="pi pi-spin pi-spinner text-sky-400"></i>
                <i v-else class="pi pi-search text-sky-400"></i>
              </button>
            </div>
          </div>
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
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div class="sm:col-span-2">
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
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('login')
const error = ref('')
const successMsg = ref('')
const loading = ref(false)
const searchingRuc = ref(false)
const searchingDni = ref(false)

const localesAnexos = ref<any[]>([])
const mainRucAddress = ref('')
const mainRucUbigeo = ref('')

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
  sector: 'core',
  ubigeo: '',
  admin_dni: '',
  admin_name: '',
  admin_email: '',
  admin_username: '',
  admin_password: ''
})

// Auto-trigger query when RUC reaches exactly 11 digits
watch(() => regData.ruc, (newVal) => {
  // Clean non-digits
  const digitsOnly = newVal.replace(/\D/g, '')
  if (newVal !== digitsOnly) {
    regData.ruc = digitsOnly
  }
  if (regData.ruc.length === 11) {
    buscarRUC()
  }
})

// Auto-trigger query when DNI reaches exactly 8 digits
watch(() => regData.admin_dni, (newVal) => {
  if (!newVal) return
  // Clean non-digits
  const digitsOnly = newVal.replace(/\D/g, '')
  if (newVal !== digitsOnly) {
    regData.admin_dni = digitsOnly
  }
  if (regData.admin_dni.length === 8) {
    buscarDNI()
  }
})

async function buscarRUC() {
  if (regData.ruc.length !== 11) return
  error.value = ''
  searchingRuc.value = true
  localesAnexos.value = []
  mainRucAddress.value = ''
  mainRucUbigeo.value = ''
  try {
    const response = await axios.get(`/public/ruc/${regData.ruc}`)
    if (response.data.success && response.data.data) {
      const info = response.data.data
      regData.razon_social = info.razon_social || info.nombre_o_razon_social || ''
      
      // Determine fiscal address
      let mappedAddress = info.direccion || info.direccion_completa || ''
      // If address is empty or is a hyphen (typical of RUC 10 sometimes)
      if (!mappedAddress || mappedAddress.trim() === '-') {
        // Construct from department, province, district
        const parts = []
        if (info.departamento && info.departamento !== '-') parts.push(info.departamento)
        if (info.provincia && info.provincia !== '-') parts.push(info.provincia)
        if (info.distrito && info.distrito !== '-') parts.push(info.distrito)
        mappedAddress = parts.join(' - ')
      }
      regData.direccion = mappedAddress
      mainRucAddress.value = mappedAddress
      regData.nombre_comercial = info.nombre_comercial || regData.razon_social || ''

      // Parse Ubigeo
      let ubi = ''
      if (info.ubigeo_sunat) {
        ubi = info.ubigeo_sunat
      } else if (Array.isArray(info.ubigeo) && info.ubigeo.length > 0) {
        // Ubigeo is array like ["15", "1501", "150101"], pick last
        ubi = info.ubigeo[info.ubigeo.length - 1]
      } else if (typeof info.ubigeo === 'string') {
        ubi = info.ubigeo
      }
      regData.ubigeo = ubi
      mainRucUbigeo.value = ubi

      // Set locales anexos if any
      if (Array.isArray(info.locales_anexos) && info.locales_anexos.length > 0) {
        localesAnexos.value = info.locales_anexos
      }
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'No se pudo encontrar el RUC ingresado'
  } finally {
    searchingRuc.value = false
  }
}

function seleccionarDireccionAnexa(event: any) {
  const index = event.target.value
  if (index === '') {
    // Revert to main
    regData.direccion = mainRucAddress.value
    regData.ubigeo = mainRucUbigeo.value
  } else {
    const selected = localesAnexos.value[Number(index)]
    if (selected) {
      regData.direccion = selected.direccion || ''
      regData.ubigeo = selected.ubigeo || ''
    }
  }
}

async function buscarDNI() {
  if (regData.admin_dni.length !== 8) return
  error.value = ''
  searchingDni.value = true
  try {
    const response = await axios.get(`/public/dni/${regData.admin_dni}`)
    if (response.data.success && response.data.data) {
      const info = response.data.data
      if (info.full_name) {
        regData.admin_name = info.full_name
      } else if (info.nombre_completa) {
        regData.admin_name = info.nombre_completa
      } else if (info.nombre_completo) {
        regData.admin_name = info.nombre_completo
      } else if (info.nombres) {
        regData.admin_name = `${info.nombres} ${info.apellido_paterno || ''} ${info.apellido_materno || ''}`.trim()
      } else if (info.first_name) {
        regData.admin_name = `${info.first_name} ${info.first_last_name || ''} ${info.second_last_name || ''}`.trim()
      }
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'No se pudo encontrar el DNI ingresado'
  } finally {
    searchingDni.value = false
  }
}

async function handleLogin() {
  error.value = ''
  successMsg.value = ''
  const success = await authStore.login(loginData.username, loginData.password)
  if (success) {
    router.push(authStore.homeRoute)
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
        sector: 'core',
        ubigeo: '',
        admin_dni: '',
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
