<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h2 class="text-2xl font-bold text-white flex items-center gap-2">
          <i class="pi pi-building text-sky-400"></i>
          Sucursales
        </h2>
        <p class="text-sm text-slate-400">Gestione las ubicaciones físicas de su empresa</p>
      </div>
      <button 
        v-if="authStore.user?.role_type === 'COMPANY_ADMIN'"
        @click="openNewModal" 
        class="bg-sky-500 hover:bg-sky-400 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-colors flex items-center gap-2"
      >
        <i class="pi pi-plus"></i>
        Nueva Sucursal
      </button>
    </div>

    <!-- Error/Loading -->
    <div v-if="loading" class="flex justify-center py-12">
      <i class="pi pi-spin pi-spinner text-4xl text-sky-500"></i>
    </div>
    <div v-else-if="error" class="bg-red-500/10 border border-red-500/50 text-red-400 p-4 rounded-xl flex items-center gap-3">
      <i class="pi pi-exclamation-triangle text-xl"></i>
      <p>{{ error }}</p>
    </div>

    <!-- Data Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div 
        v-for="branch in branches" 
        :key="branch.id"
        class="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-xl p-5 hover:border-sky-500/50 transition-colors group relative"
      >
        <div class="flex justify-between items-start mb-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-sky-500/20 flex items-center justify-center text-sky-400">
              <i class="pi pi-map-marker text-xl"></i>
            </div>
            <div>
              <h3 class="font-bold text-white text-lg">{{ branch.nombre }}</h3>
              <p class="text-xs text-slate-400">{{ branch.id.substring(0,8) }}...</p>
            </div>
          </div>
          <!-- Action Buttons -->
          <div v-if="authStore.user?.role_type === 'COMPANY_ADMIN'" class="opacity-0 group-hover:opacity-100 transition-opacity flex gap-2">
            <button @click="openEditModal(branch)" class="p-1.5 text-slate-400 hover:text-sky-400 bg-slate-800 rounded-lg transition-colors" title="Editar">
              <i class="pi pi-pencil"></i>
            </button>
          </div>
        </div>

        <div class="space-y-2 text-sm text-slate-300">
          <div class="flex items-start gap-2">
            <i class="pi pi-directions mt-1 text-slate-500"></i>
            <span>{{ branch.direccion || 'Sin dirección registrada' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <i class="pi pi-phone text-slate-500"></i>
            <span>{{ branch.telefono || 'Sin teléfono' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <i class="pi pi-envelope text-slate-500"></i>
            <span>{{ branch.email || 'Sin correo' }}</span>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="branches.length === 0" class="col-span-full bg-slate-800/30 border border-slate-700/50 rounded-2xl p-12 text-center">
        <div class="w-16 h-16 bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4">
          <i class="pi pi-building text-2xl text-slate-400"></i>
        </div>
        <h3 class="text-lg font-bold text-white mb-2">No hay sucursales</h3>
        <p class="text-slate-400 mb-6 max-w-md mx-auto">Comience registrando la sede principal de la empresa.</p>
        <button 
          v-if="authStore.user?.role_type === 'COMPANY_ADMIN'"
          @click="openNewModal" 
          class="bg-sky-500 hover:bg-sky-400 text-white px-6 py-2.5 rounded-lg text-sm font-semibold transition-colors inline-flex items-center gap-2"
        >
          <i class="pi pi-plus"></i>
          Crear Primera Sucursal
        </button>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal"></div>
      
      <div class="bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl w-full max-w-lg relative z-10 animate-fade-in-up">
        <div class="p-6 border-b border-slate-800 flex justify-between items-center">
          <h3 class="text-xl font-bold text-white">
            {{ isEditing ? 'Editar Sucursal' : 'Nueva Sucursal' }}
          </h3>
          <button @click="closeModal" class="text-slate-400 hover:text-white transition-colors">
            <i class="pi pi-times text-xl"></i>
          </button>
        </div>

        <form @submit.prevent="saveBranch" class="p-6 space-y-4">
          <div v-if="formError" class="bg-red-500/10 border border-red-500/50 text-red-400 p-3 rounded-lg text-sm flex items-center gap-2">
            <i class="pi pi-exclamation-circle"></i>
            {{ formError }}
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Nombre / Alias *</label>
            <input 
              v-model="formData.nombre" 
              type="text" 
              required
              class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 focus:border-sky-500 transition-all text-sm"
              placeholder="Ej. Sede Principal"
            >
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Dirección Física</label>
            <input 
              v-model="formData.direccion" 
              type="text" 
              class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 focus:border-sky-500 transition-all text-sm"
              placeholder="Ej. Av. Central 123"
            >
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Teléfono</label>
              <input 
                v-model="formData.telefono" 
                type="text" 
                class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 focus:border-sky-500 transition-all text-sm"
                placeholder="Ej. +51 987654321"
              >
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Email Público</label>
              <input 
                v-model="formData.email" 
                type="email" 
                class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 focus:border-sky-500 transition-all text-sm"
                placeholder="contacto@sede.com"
              >
            </div>
          </div>

          <div class="pt-4 flex justify-end gap-3">
            <button 
              type="button" 
              @click="closeModal"
              class="px-5 py-2.5 text-sm font-semibold text-slate-300 hover:text-white transition-colors"
            >
              Cancelar
            </button>
            <button 
              type="submit" 
              :disabled="saving"
              class="bg-sky-500 hover:bg-sky-400 text-white px-6 py-2.5 rounded-lg text-sm font-semibold transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <i v-if="saving" class="pi pi-spin pi-spinner"></i>
              {{ isEditing ? 'Guardar Cambios' : 'Crear Sucursal' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

const authStore = useAuthStore()

const branches = ref<any[]>([])
const loading = ref(true)
const error = ref('')

const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const formError = ref('')

const formData = ref({
  id: '',
  nombre: '',
  direccion: '',
  telefono: '',
  email: ''
})

async function fetchBranches() {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get('/branches')
    if (res.data.success) {
      branches.value = res.data.data
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Error al cargar las sucursales'
  } finally {
    loading.value = false
  }
}

function openNewModal() {
  isEditing.value = false
  formError.value = ''
  formData.value = {
    id: '',
    nombre: '',
    direccion: '',
    telefono: '',
    email: ''
  }
  showModal.value = true
}

function openEditModal(branch: any) {
  isEditing.value = true
  formError.value = ''
  formData.value = {
    id: branch.id,
    nombre: branch.nombre,
    direccion: branch.direccion || '',
    telefono: branch.telefono || '',
    email: branch.email || ''
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function saveBranch() {
  formError.value = ''
  saving.value = true
  
  try {
    const payload = { ...formData.value }
    if (!isEditing.value) {
      delete payload.id
    }

    if (isEditing.value) {
      await axios.put(`/branches/${payload.id}`, payload)
    } else {
      await axios.post('/branches', payload)
    }
    
    closeModal()
    fetchBranches()
  } catch (err: any) {
    formError.value = err.response?.data?.error || 'Error al guardar la sucursal'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchBranches()
})
</script>
