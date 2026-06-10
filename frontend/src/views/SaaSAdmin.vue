<template>
  <div class="min-h-screen bg-gray-100">
    <!-- Navbar -->
    <nav class="bg-indigo-700 text-white shadow-lg">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16 items-center">
          <div class="flex items-center space-x-3">
            <div class="bg-white p-1 rounded-lg">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-indigo-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            </div>
            <span class="text-xl font-bold tracking-tight">ERP Veterinaria SaaS Panel</span>
          </div>
          <div class="flex items-center space-x-4">
            <span class="text-sm font-medium bg-indigo-800 px-3 py-1 rounded-full">{{ authStore.user?.nombre }} (SuperAdmin)</span>
            <button @click="authStore.logout()" class="text-sm bg-white text-indigo-700 hover:bg-indigo-50 px-3 py-2 rounded-md font-semibold transition-colors duration-200 shadow-sm">
              Cerrar Sesión
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex items-center space-x-4">
          <div class="p-3 bg-blue-100 rounded-lg text-blue-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">Total Empresas</p>
            <p class="text-2xl font-bold text-gray-900">{{ stats.total_companies || 0 }}</p>
          </div>
        </div>
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex items-center space-x-4">
          <div class="p-3 bg-green-100 rounded-lg text-green-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">Suscripciones Activas</p>
            <p class="text-2xl font-bold text-gray-900">{{ stats.active_companies || 0 }}</p>
          </div>
        </div>
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex items-center space-x-4">
          <div class="p-3 bg-purple-100 rounded-lg text-purple-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">Total Sedes</p>
            <p class="text-2xl font-bold text-gray-900">{{ stats.total_branches || 0 }}</p>
          </div>
        </div>
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex items-center space-x-4">
          <div class="p-3 bg-yellow-100 rounded-lg text-yellow-600">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">Ventas Totales (S/.)</p>
            <p class="text-2xl font-bold text-gray-900">{{ stats.total_sales_value?.toLocaleString() || 0 }}</p>
          </div>
        </div>
      </div>

      <!-- Companies Table -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div class="px-6 py-4 border-b border-gray-100 bg-gray-50 flex justify-between items-center">
          <h2 class="text-lg font-bold text-gray-800">Inquilinos (Empresas)</h2>
          <button @click="fetchData" class="text-indigo-600 hover:text-indigo-800 text-sm font-medium">Actualizar Lista</button>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead>
              <tr class="bg-gray-50">
                <th class="px-6 py-3 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Empresa / RUC</th>
                <th class="px-6 py-3 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Plan / Sedes</th>
                <th class="px-6 py-3 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Expiración</th>
                <th class="px-6 py-3 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider">Estado</th>
                <th class="px-6 py-3 text-right text-xs font-semibold text-gray-500 uppercase tracking-wider">Acciones</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="company in companies" :key="company.id" class="hover:bg-gray-50 transition-colors duration-150">
                <td class="px-6 py-4">
                  <div class="text-sm font-bold text-gray-900">{{ company.razon_social }}</div>
                  <div class="text-sm text-gray-500">{{ company.ruc }}</div>
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 text-xs font-bold rounded-full" 
                    :class="company.plan_type === 'Premium' ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'">
                    {{ company.plan_type }}
                  </span>
                  <div class="text-xs text-gray-500 mt-1">Sedes: {{ company.branches_count }} / {{ company.max_branches }}</div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500">
                  {{ new Date(company.subscription_expires_at).toLocaleDateString() }}
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 text-xs font-bold rounded-full" 
                    :class="new Date(company.subscription_expires_at) > new Date() ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
                    {{ new Date(company.subscription_expires_at) > new Date() ? 'Activo' : 'Vencido' }}
                  </span>
                </td>
                <td class="px-6 py-4 text-right">
                  <button @click="openEditModal(company)" class="text-indigo-600 hover:text-indigo-900 font-bold text-sm">Gestionar</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </main>

    <!-- Edit Subscription Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" @click="showModal = false"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
        <div class="inline-block align-bottom bg-white rounded-xl text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <h3 class="text-lg font-bold text-gray-900 mb-4">Gestionar Suscripción: {{ selectedCompany?.razon_social }}</h3>
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">Tipo de Plan</label>
                <select v-model="editForm.plan_type" class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500">
                  <option value="Basic">Basic</option>
                  <option value="Premium">Premium</option>
                  <option value="Enterprise">Enterprise</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">Límite de Sedes</label>
                <input type="number" v-model.number="editForm.max_branches" class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">Fecha de Expiración</label>
                <input type="date" v-model="editForm.expires_at" class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button @click="updateSubscription" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 sm:ml-3 sm:w-auto sm:text-sm">Guardar Cambios</button>
            <button @click="showModal = false" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">Cancelar</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const companies = ref<any[]>([])
const stats = ref<any>({})
const showModal = ref(false)
const selectedCompany = ref<any>(null)
const editForm = ref({
  plan_type: '',
  max_branches: 0,
  expires_at: ''
})

const fetchData = async () => {
  try {
    const [compRes, statsRes] = await Promise.all([
      axios.get('/saas-admin/companies'),
      axios.get('/saas-admin/stats')
    ])
    companies.value = compRes.data.data
    stats.value = statsRes.data.data
  } catch (err) {
    console.error('Error fetching SaaS data', err)
  }
}

const openEditModal = (company: any) => {
  selectedCompany.value = company
  editForm.value = {
    plan_type: company.plan_type,
    max_branches: company.max_branches,
    expires_at: new Date(company.subscription_expires_at).toISOString().split('T')[0]
  }
  showModal.value = true
}

const updateSubscription = async () => {
  try {
    const res = await axios.put(`/saas-admin/companies/${selectedCompany.value.id}/subscription`, {
      plan_type: editForm.value.plan_type,
      max_branches: editForm.value.max_branches,
      subscription_expires_at: new Date(editForm.value.expires_at).toISOString()
    })
    if (res.data.success) {
      showModal.value = false
      fetchData()
    }
  } catch (err) {
    alert('Error al actualizar suscripción')
  }
}

onMounted(fetchData)
</script>
