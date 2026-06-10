<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h2 class="text-2xl font-bold text-white flex items-center gap-2">
          <i class="pi pi-users text-sky-400"></i>
          Personal y Roles
        </h2>
        <p class="text-sm text-slate-400">Administre el acceso, roles y asignaciones del equipo</p>
      </div>
      <button 
        v-if="authStore.user?.role_type === 'COMPANY_ADMIN' || authStore.user?.role_type === 'BRANCH_ADMIN'"
        @click="openNewModal" 
        class="bg-sky-500 hover:bg-sky-400 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-colors flex items-center gap-2"
      >
        <i class="pi pi-user-plus"></i>
        Registrar Empleado
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

    <!-- Data Table -->
    <div v-else class="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-xl overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm text-slate-300">
          <thead class="text-xs text-slate-400 uppercase bg-slate-900/50 border-b border-slate-700">
            <tr>
              <th class="px-6 py-4 font-semibold">Empleado</th>
              <th class="px-6 py-4 font-semibold">Usuario (Login)</th>
              <th class="px-6 py-4 font-semibold">Sucursal</th>
              <th class="px-6 py-4 font-semibold">Rol Asignado</th>
              <th class="px-6 py-4 font-semibold">Estado</th>
              <th v-if="canEdit" class="px-6 py-4 text-right font-semibold">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/50">
            <tr v-for="user in users" :key="user.id" class="hover:bg-slate-800/80 transition-colors group">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-sky-600 to-indigo-600 flex items-center justify-center text-white font-bold text-xs uppercase shadow-lg shadow-sky-900/20">
                    {{ user.nombre.charAt(0) }}
                  </div>
                  <div>
                    <div class="font-semibold text-white">{{ user.nombre }}</div>
                    <div class="text-xs text-slate-400">{{ user.email }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 font-mono text-xs text-sky-300">{{ user.username }}</td>
              <td class="px-6 py-4">
                <span v-if="user.branch_id" class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-slate-800 border border-slate-700 text-xs">
                  <i class="pi pi-building text-slate-400"></i>
                  {{ getBranchName(user.branch_id) }}
                </span>
                <span v-else class="text-slate-500 italic">No asignada</span>
              </td>
              <td class="px-6 py-4">
                <span :class="[
                  'px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider',
                  user.role_type === 'COMPANY_ADMIN' ? 'bg-purple-500/20 text-purple-400 border border-purple-500/30' : 
                  user.role_type === 'BRANCH_ADMIN' ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30' : 
                  'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                ]">
                  {{ formatRole(user.role_type) }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span :class="[
                  'px-2 py-1 rounded-md text-xs font-medium',
                  user.estado === 'active' ? 'text-emerald-400 bg-emerald-400/10' : 'text-red-400 bg-red-400/10'
                ]">
                  {{ user.estado === 'active' ? 'Activo' : 'Inactivo' }}
                </span>
              </td>
              <td v-if="canEdit" class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button v-if="authStore.user?.role_type === 'COMPANY_ADMIN'" @click="openEditModal(user)" class="p-1.5 text-slate-400 hover:text-sky-400 bg-slate-800 rounded-md transition-colors" title="Editar">
                    <i class="pi pi-pencil"></i>
                  </button>
                  <button v-if="authStore.user?.role_type === 'COMPANY_ADMIN' && user.id !== authStore.user?.id" @click="deleteUser(user.id)" class="p-1.5 text-slate-400 hover:text-red-400 bg-slate-800 rounded-md transition-colors" title="Eliminar">
                    <i class="pi pi-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="users.length === 0">
              <td colspan="6" class="px-6 py-12 text-center text-slate-400">
                No hay empleados registrados en esta sucursal.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal"></div>
      
      <div class="bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl w-full max-w-2xl relative z-10 animate-fade-in-up flex flex-col max-h-[90vh]">
        <div class="p-6 border-b border-slate-800 flex justify-between items-center shrink-0">
          <h3 class="text-xl font-bold text-white flex items-center gap-2">
            <i :class="isEditing ? 'pi pi-user-edit' : 'pi pi-user-plus'" class="text-sky-400"></i>
            {{ isEditing ? 'Editar Empleado' : 'Registrar Nuevo Empleado' }}
          </h3>
          <button @click="closeModal" class="text-slate-400 hover:text-white transition-colors">
            <i class="pi pi-times text-xl"></i>
          </button>
        </div>

        <div class="p-6 overflow-y-auto grow">
          <form @submit.prevent="saveUser" class="space-y-6">
            <div v-if="formError" class="bg-red-500/10 border border-red-500/50 text-red-400 p-3 rounded-lg text-sm flex items-center gap-2">
              <i class="pi pi-exclamation-circle"></i>
              {{ formError }}
            </div>

            <!-- Datos Personales -->
            <div>
              <h4 class="text-xs font-bold text-sky-500 uppercase tracking-wider mb-4 border-b border-slate-800 pb-2">1. Datos Personales</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="md:col-span-2">
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Nombre Completo *</label>
                  <input v-model="formData.nombre" type="text" required class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm" placeholder="Ej. Juan Pérez">
                </div>
                <div>
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Correo Electrónico *</label>
                  <input v-model="formData.email" type="email" required class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm" placeholder="juan@empresa.com">
                </div>
                <div>
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Usuario de Acceso (Login) *</label>
                  <input v-model="formData.username" type="text" required class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm" placeholder="Ej. jperez">
                </div>
                <div v-if="!isEditing" class="md:col-span-2">
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Contraseña Inicial *</label>
                  <input v-model="formData.password" type="password" :required="!isEditing" minlength="8" class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm" placeholder="Mínimo 8 caracteres">
                  <p class="text-xs text-slate-500 mt-1">El empleado usará esta contraseña para ingresar al sistema.</p>
                </div>
              </div>
            </div>

            <!-- Asignación y Permisos -->
            <div>
              <h4 class="text-xs font-bold text-sky-500 uppercase tracking-wider mb-4 border-b border-slate-800 pb-2">2. Asignación y Permisos</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Sucursal Asignada *</label>
                  <select v-model="formData.branch_id" required :disabled="isEditing || authStore.user?.role_type === 'BRANCH_ADMIN'" class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm appearance-none cursor-pointer disabled:opacity-50">
                    <option value="" disabled>Seleccione una sucursal</option>
                    <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.nombre }}</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Rol en el Sistema *</label>
                  <select v-model="formData.role_type" required :disabled="authStore.user?.role_type === 'BRANCH_ADMIN'" class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm appearance-none cursor-pointer disabled:opacity-50">
                    <option value="" disabled>Seleccione el rol</option>
                    <option v-if="authStore.user?.role_type === 'COMPANY_ADMIN'" value="COMPANY_ADMIN">Administrador de Empresa</option>
                    <option v-if="authStore.user?.role_type === 'COMPANY_ADMIN'" value="BRANCH_ADMIN">Encargado de Sucursal</option>
                    <option value="CASHIER">Cajero (Ventas y Stock)</option>
                  </select>
                </div>
                <div v-if="isEditing" class="md:col-span-2">
                  <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">Estado de la Cuenta</label>
                  <select v-model="formData.estado" required class="w-full bg-slate-800 border border-slate-700 text-white px-4 py-2.5 rounded-lg focus:ring-2 focus:ring-sky-500 transition-all text-sm appearance-none cursor-pointer">
                    <option value="active">Activo (Permitir acceso)</option>
                    <option value="inactive">Inactivo (Bloquear acceso)</option>
                  </select>
                </div>
              </div>
            </div>
          </form>
        </div>
        
        <div class="p-6 border-t border-slate-800 shrink-0 flex justify-end gap-3 bg-slate-900/50 rounded-b-2xl">
          <button type="button" @click="closeModal" class="px-5 py-2.5 text-sm font-semibold text-slate-300 hover:text-white transition-colors">Cancelar</button>
          <button type="submit" @click="saveUser" :disabled="saving" class="bg-sky-500 hover:bg-sky-400 text-white px-6 py-2.5 rounded-lg text-sm font-semibold transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
            <i v-if="saving" class="pi pi-spin pi-spinner"></i>
            {{ isEditing ? 'Guardar Cambios' : 'Crear Empleado' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

const authStore = useAuthStore()

const users = ref<any[]>([])
const branches = ref<any[]>([])
const loading = ref(true)
const error = ref('')

const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const formError = ref('')

const canEdit = computed(() => {
  return authStore.user?.role_type === 'COMPANY_ADMIN'
})

const formData = ref({
  id: '',
  nombre: '',
  email: '',
  username: '',
  password: '',
  role_type: '',
  branch_id: '',
  estado: 'active'
})

async function fetchData() {
  loading.value = true
  error.value = ''
  try {
    const [resUsers, resBranches] = await Promise.all([
      axios.get('/users'),
      axios.get('/branches')
    ])
    
    if (resUsers.data.success) {
      users.value = resUsers.data.data
    }
    if (resBranches.data.success) {
      branches.value = resBranches.data.data
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Error al cargar los datos del personal'
  } finally {
    loading.value = false
  }
}

function getBranchName(branchId: string) {
  const branch = branches.value.find(b => b.id === branchId)
  return branch ? branch.nombre : 'Desconocida'
}

function formatRole(role: string) {
  switch(role) {
    case 'COMPANY_ADMIN': return 'Admin. Empresa'
    case 'BRANCH_ADMIN': return 'Encargado'
    case 'BRANCH_USER': return 'Usuario'
    case 'CASHIER': return 'Cajero'
    default: return role
  }
}

function openNewModal() {
  isEditing.value = false
  formError.value = ''
  formData.value = {
    id: '',
    nombre: '',
    email: '',
    username: '',
    password: '',
    role_type: authStore.user?.role_type === 'BRANCH_ADMIN' ? 'CASHIER' : '',
    branch_id: authStore.user?.role_type === 'BRANCH_ADMIN' ? (authStore.user?.branch_id || '') : '',
    estado: 'active'
  }
  showModal.value = true
}

function openEditModal(user: any) {
  isEditing.value = true
  formError.value = ''
  formData.value = {
    id: user.id,
    nombre: user.nombre,
    email: user.email,
    username: user.username,
    password: '', // No password on edit
    role_type: user.role_type,
    branch_id: user.branch_id,
    estado: user.estado || 'active'
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function saveUser() {
  formError.value = ''
  saving.value = true
  
  try {
    const payload = { ...formData.value }
    if (!isEditing.value) {
      delete payload.id
    }

    if (isEditing.value) {
      await axios.put(`/users/${payload.id}`, {
        nombre: payload.nombre,
        email: payload.email,
        username: payload.username,
        role_type: payload.role_type,
        estado: payload.estado
      })
    } else {
      await axios.post('/users', payload)
    }
    
    closeModal()
    fetchData()
  } catch (err: any) {
    formError.value = err.response?.data?.error || 'Error al guardar el empleado'
  } finally {
    saving.value = false
  }
}

async function deleteUser(id: string) {
  if (confirm('¿Está seguro de eliminar o desactivar este empleado?')) {
    try {
      await axios.delete(`/users/${id}`)
      fetchData()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Error al eliminar')
    }
  }
}

onMounted(() => {
  fetchData()
})
</script>
