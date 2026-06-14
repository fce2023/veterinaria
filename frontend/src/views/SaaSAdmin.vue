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

      <!-- Navigation Tabs -->
      <div class="flex space-x-4 mb-6">
        <button 
          @click="activeTab = 'companies'"
          :class="[
            'px-4 py-2 text-sm font-bold rounded-lg transition-all',
            activeTab === 'companies' ? 'bg-indigo-600 text-white shadow-md' : 'bg-white border border-gray-200 text-gray-650 hover:bg-gray-50'
          ]"
        >
          Inquilinos (Empresas)
        </button>
        <button 
          @click="activeTab = 'factura_api'"
          :class="[
            'px-4 py-2 text-sm font-bold rounded-lg transition-all',
            activeTab === 'factura_api' ? 'bg-indigo-600 text-white shadow-md' : 'bg-white border border-gray-200 text-gray-650 hover:bg-gray-50'
          ]"
        >
          Configuración Global FacturaAPI
        </button>
      </div>

      <!-- Companies Table -->
      <div v-if="activeTab === 'companies'" class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden animate-fade-in">
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
                  <div v-if="company.ruc.startsWith('10') && company.nombre_comercial && company.nombre_comercial !== company.razon_social" class="text-xs text-indigo-600 font-semibold italic mt-0.5">
                    Comercial: {{ company.nombre_comercial }}
                  </div>
                  <div class="text-sm text-gray-500 mt-0.5">{{ company.ruc }}</div>
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 text-xs font-bold rounded-full" 
                    :class="company.plan_type === 'Premium' ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'">
                    {{ company.plan_type }}
                  </span>
                  <div class="text-xs text-gray-500 mt-1">Sedes: {{ company.branches_count }} / {{ company.max_branches }}</div>
                  <div class="flex flex-wrap gap-1 mt-2">
                    <span v-for="mod in company.modules" :key="mod" class="px-1.5 py-0.5 text-[9px] font-black uppercase tracking-wider bg-slate-100 text-slate-600 rounded-lg">
                      {{ mod }}
                    </span>
                  </div>
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
                <td class="px-6 py-4 text-right space-x-3">
                  <button @click="openDetailsModal(company)" class="text-sky-600 hover:text-sky-900 font-bold text-sm">Detalles</button>
                  <button @click="openEditModal(company)" class="text-indigo-600 hover:text-indigo-900 font-bold text-sm">Gestionar</button>
                  <button @click="deleteCompany(company)" class="text-red-600 hover:text-red-900 font-bold text-sm">Eliminar</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- FacturaAPI Global Configuration Form -->
      <div v-else-if="activeTab === 'factura_api'" class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden animate-fade-in max-w-2xl">
        <div class="px-6 py-4 border-b border-gray-100 bg-gray-50">
          <h2 class="text-lg font-bold text-gray-800">Configuración Global FacturaAPI</h2>
          <p class="text-xs text-gray-550 mt-1">Configura la API Key global del administrador del SaaS y el endpoint de FacturaAPI.</p>
        </div>
        <form @submit.prevent="saveGlobalSettings" class="p-6 space-y-6">
          <div>
            <label class="block text-xs font-bold text-gray-700 mb-1">URL Base del API (Global)</label>
            <input 
              v-model="globalSettings.factura_api_url"
              type="url" 
              placeholder="https://apifacturacion.sehuacho.com"
              class="w-full px-4 py-2 border border-gray-300 rounded-md text-xs focus:ring-indigo-500 focus:border-indigo-500 outline-none"
              required
            />
            <p class="text-[10px] text-amber-600 font-bold mt-1">
              ⚠️ IMPORTANTE: Ingresar únicamente el dominio base. NO incluyas "/api/v1" ni barras diagonales "/" al final (Ejemplo correcto: https://apifacturacion.sehuacho.com).
            </p>
          </div>

          <div>
            <label class="block text-xs font-bold text-gray-700 mb-1">API Key Global (SaaS Admin)</label>
            <div class="relative">
              <input 
                :type="showGlobalKey ? 'text' : 'password'"
                v-model="globalSettings.factura_api_key"
                placeholder="Ingresar API Key global..."
                class="w-full pl-4 pr-10 py-2 border border-gray-300 rounded-md text-xs focus:ring-indigo-500 focus:border-indigo-500 outline-none font-mono"
                required
              />
              <button 
                type="button"
                @click="showGlobalKey = !showGlobalKey"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-indigo-600"
              >
                <i :class="showGlobalKey ? 'pi pi-eye-slash' : 'pi pi-eye'"></i>
              </button>
            </div>
            <p class="text-[10px] text-gray-400 mt-1">Esta llave asegura que solo tu SaaS pueda consumir el microservicio.</p>
          </div>

          <div class="pt-4 border-t border-gray-100 flex justify-end">
            <button 
              type="submit"
              :disabled="savingGlobal"
              class="px-6 py-2 bg-indigo-650 hover:bg-indigo-700 text-white rounded-md text-xs font-bold shadow-md transition-all flex items-center gap-2"
            >
              <i v-if="savingGlobal" class="pi pi-spin pi-spinner"></i>
              <span>{{ savingGlobal ? 'Guardando...' : 'Guardar Configuración Global' }}</span>
            </button>
          </div>
        </form>
      </div>
    </main>

    <!-- Details Modal -->
    <div v-if="showDetailsModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" @click="showDetailsModal = false"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
        <div class="inline-block align-bottom bg-white rounded-xl text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-xl sm:w-full">
          <div class="bg-indigo-700 px-6 py-4 text-white">
            <h3 class="text-lg font-bold">Detalles de la Empresa</h3>
            <p class="text-xs text-indigo-200 mt-0.5">Información general y de suscripción registrada</p>
          </div>
          <div class="bg-white px-6 py-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Razón Social</label>
                <div class="text-sm font-bold text-gray-900 mt-0.5">{{ selectedCompanyDetails?.razon_social }}</div>
              </div>
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">RUC</label>
                <div class="text-sm font-mono text-gray-900 mt-0.5">{{ selectedCompanyDetails?.ruc }}</div>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Nombre Comercial</label>
                <div class="text-sm text-gray-900 mt-0.5">{{ selectedCompanyDetails?.nombre_comercial || '-' }}</div>
              </div>
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Estado de Empresa</label>
                <div class="mt-0.5">
                  <span class="px-2 py-0.5 text-xs font-bold rounded-full bg-green-100 text-green-700">
                    {{ selectedCompanyDetails?.estado }}
                  </span>
                </div>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Dirección Fiscal / Ubigeo</label>
              <div class="text-sm text-gray-900 mt-0.5">
                {{ selectedCompanyDetails?.direccion || '-' }} 
                <span v-if="selectedCompanyDetails?.ubigeo" class="text-xs font-mono bg-slate-100 text-slate-600 px-1.5 py-0.5 rounded ml-1.5">
                  Ubigeo: {{ selectedCompanyDetails.ubigeo }}
                </span>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Email de Contacto</label>
                <div class="text-sm text-gray-900 mt-0.5">{{ selectedCompanyDetails?.email || '-' }}</div>
              </div>
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Teléfono</label>
                <div class="text-sm text-gray-900 mt-0.5">{{ selectedCompanyDetails?.telefono || '-' }}</div>
              </div>
            </div>

            <div class="border-t border-gray-100 pt-4">
              <h4 class="text-xs font-black text-indigo-700 uppercase tracking-widest mb-3">Información del Plan y SaaS</h4>
              
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Plan</label>
                  <div class="text-sm font-bold text-gray-900 mt-0.5">{{ selectedCompanyDetails?.plan_type }}</div>
                </div>
                <div>
                  <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Sedes Registradas</label>
                  <div class="text-sm text-gray-900 mt-0.5">{{ selectedCompanyDetails?.branches_count }} / {{ selectedCompanyDetails?.max_branches }}</div>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-4 mt-3">
                <div>
                  <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400">Expiración Suscripción</label>
                  <div class="text-sm text-gray-900 mt-0.5">
                    {{ selectedCompanyDetails?.subscription_expires_at ? new Date(selectedCompanyDetails.subscription_expires_at).toLocaleDateString() : '-' }}
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold uppercase tracking-wider text-gray-400 font-bold">Módulos Habilitados</label>
                  <div class="flex flex-wrap gap-1.5 mt-1">
                    <span v-for="mod in selectedCompanyDetails?.modules" :key="mod" class="px-2 py-0.5 text-[10px] font-black uppercase tracking-wider bg-indigo-50 text-indigo-700 rounded-md">
                      {{ mod }}
                    </span>
                    <span v-if="!selectedCompanyDetails?.modules || selectedCompanyDetails.modules.length === 0" class="text-xs text-gray-400">Ninguno</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-6 py-3 sm:flex sm:flex-row-reverse border-t border-gray-100">
            <button @click="showDetailsModal = false" class="w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 sm:w-auto sm:text-sm">Cerrar</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Subscription Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" @click="showModal = false"></div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
        <div class="inline-block align-bottom bg-white rounded-xl text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-6 pt-5 pb-4">
            <h3 class="text-lg font-bold text-gray-900 mb-4">Gestionar Empresa y Suscripción: {{ selectedCompany?.razon_social }}</h3>
            
            <div class="space-y-4 max-h-[60vh] overflow-y-auto pr-1">
              <!-- General Info Section -->
              <h4 class="text-xs font-black text-indigo-750 uppercase tracking-widest border-b border-gray-150 pb-1 mt-2">Datos Básicos de Empresa</h4>
              
              <div>
                <label class="block text-xs font-medium text-gray-700">Razón Social</label>
                <input type="text" v-model="editForm.razon_social" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>

              <div>
                <label class="block text-xs font-medium text-gray-700">Nombre Comercial</label>
                <input type="text" v-model="editForm.nombre_comercial" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>

              <div>
                <label class="block text-xs font-medium text-gray-700">Dirección Fiscal</label>
                <input type="text" v-model="editForm.direccion" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-gray-700">Teléfono</label>
                  <input type="text" v-model="editForm.telefono" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-gray-700">Correo Contacto</label>
                  <input type="email" v-model="editForm.email" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
                </div>
              </div>

              <!-- Subscription Section -->
              <h4 class="text-xs font-black text-indigo-750 uppercase tracking-widest border-b border-gray-150 pb-1 pt-3">Configuración de Suscripción</h4>

              <div>
                <label class="block text-xs font-medium text-gray-700">Tipo de Plan</label>
                <select v-model="editForm.plan_type" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500">
                  <option value="Básico">Básico</option>
                  <option value="Premium">Premium</option>
                  <option value="Enterprise">Enterprise</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700">Límite de Sedes</label>
                <input type="number" v-model.number="editForm.max_branches" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700">Fecha de Expiración</label>
                <input type="date" v-model="editForm.expires_at" class="mt-1 block w-full text-xs border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" />
              </div>
              <div class="pt-2">
                <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-3">Módulos Habilitados</label>
                <div class="space-y-2.5">
                  <label class="flex items-center gap-3 cursor-pointer group">
                    <input type="checkbox" v-model="editForm.modules.veterinaria" class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500" />
                    <div>
                      <span class="text-sm font-bold text-gray-700 group-hover:text-indigo-600 transition-colors">Módulo Veterinaria</span>
                      <p class="text-[10px] text-gray-400">Mascotas, historial clínico</p>
                    </div>
                  </label>
                  <label class="flex items-center gap-3 cursor-pointer group">
                    <input type="checkbox" v-model="editForm.modules.vidrieria" class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500" />
                    <div>
                      <span class="text-sm font-bold text-gray-700 group-hover:text-indigo-600 transition-colors">Módulo Vidriería</span>
                      <p class="text-[10px] text-gray-400">Fórmula dimensional M, M2, M3</p>
                    </div>
                  </label>
                  <label class="flex items-center gap-3 cursor-pointer group">
                    <input type="checkbox" v-model="editForm.modules.facturacion" class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500" />
                    <div>
                      <span class="text-sm font-bold text-gray-700 group-hover:text-indigo-600 transition-colors">Facturación Electrónica</span>
                      <p class="text-[10px] text-gray-400">Integración SUNAT/FacturaAPI (certificado SOL, comprobantes)</p>
                    </div>
                  </label>
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-6 py-3 sm:flex sm:flex-row-reverse">
            <button @click="updateSubscription" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-sm font-semibold text-white hover:bg-indigo-700 sm:ml-3 sm:w-auto">Guardar Cambios</button>
            <button @click="showModal = false" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-sm font-semibold text-gray-700 hover:bg-gray-50 sm:mt-0 sm:ml-3 sm:w-auto">Cancelar</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const companies = ref<any[]>([])
const stats = ref<any>({})
const showModal = ref(false)
const selectedCompany = ref<any>(null)

// Navigation state
const activeTab = ref('companies')
const showGlobalKey = ref(false)
const savingGlobal = ref(false)
const globalSettings = ref({
  factura_api_url: '',
  factura_api_key: ''
})

const loadGlobalSettings = async () => {
  try {
    const res = await axios.get('/saas-admin/settings')
    if (res.data.success && res.data.data) {
      globalSettings.value.factura_api_url = res.data.data.factura_api_url || ''
      globalSettings.value.factura_api_key = res.data.data.factura_api_key || ''
    }
  } catch (err) {
    console.error('Error loading global settings', err)
  }
}

const saveGlobalSettings = async () => {
  savingGlobal.value = true
  try {
    const res = await axios.post('/saas-admin/settings', globalSettings.value)
    if (res.data.success) {
      alert('Configuración global de FacturaAPI guardada correctamente')
    }
  } catch (err) {
    alert('Error al guardar la configuración global')
  } finally {
    savingGlobal.value = false
  }
}

const showDetailsModal = ref(false)
const selectedCompanyDetails = ref<any>(null)

const editForm = ref({
  razon_social: '',
  nombre_comercial: '',
  direccion: '',
  telefono: '',
  email: '',
  plan_type: '',
  max_branches: 0,
  expires_at: '',
  modules: {
    veterinaria: false,
    vidrieria: false,
    facturacion: false
  }
})

const isInitializingModal = ref(false)

// Automatically suggest preset limits and modules when changing plans
watch(() => editForm.value.plan_type, (newPlan) => {
  if (isInitializingModal.value) return
  
  if (newPlan === 'Básico') {
    editForm.value.max_branches = 1
    editForm.value.modules.veterinaria = false
    editForm.value.modules.vidrieria = false
    editForm.value.modules.facturacion = false
  } else if (newPlan === 'Premium') {
    editForm.value.max_branches = 5
    editForm.value.modules.veterinaria = true
    editForm.value.modules.vidrieria = false
    editForm.value.modules.facturacion = true
  } else if (newPlan === 'Enterprise') {
    editForm.value.max_branches = 99
    editForm.value.modules.veterinaria = true
    editForm.value.modules.vidrieria = true
    editForm.value.modules.facturacion = true
  }
})

const openDetailsModal = (company: any) => {
  selectedCompanyDetails.value = company
  showDetailsModal.value = true
}

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
  const compModules = company.modules || []
  isInitializingModal.value = true
  editForm.value = {
    razon_social: company.razon_social || '',
    nombre_comercial: company.nombre_comercial || '',
    direccion: company.direccion || '',
    telefono: company.telefono || '',
    email: company.email || '',
    plan_type: company.plan_type,
    max_branches: company.max_branches,
    expires_at: company.subscription_expires_at ? new Date(company.subscription_expires_at).toISOString().split('T')[0] : '',
    modules: {
      veterinaria: compModules.includes('veterinaria'),
      vidrieria: compModules.includes('vidrieria'),
      facturacion: compModules.includes('facturacion')
    }
  }
  nextTick(() => {
    isInitializingModal.value = false
  })
  showModal.value = true
}

const updateSubscription = async () => {
  try {
    // 1. Update basic company details
    await axios.put(`/saas-admin/companies/${selectedCompany.value.id}`, {
      razon_social: editForm.value.razon_social,
      nombre_comercial: editForm.value.nombre_comercial,
      direccion: editForm.value.direccion,
      telefono: editForm.value.telefono,
      email: editForm.value.email
    })

    // 2. Update core subscription limits
    const res = await axios.put(`/saas-admin/companies/${selectedCompany.value.id}/subscription`, {
      plan_type: editForm.value.plan_type,
      max_branches: editForm.value.max_branches,
      subscription_expires_at: new Date(editForm.value.expires_at).toISOString()
    })

    if (res.data.success) {
      // 2. Update modules status by comparing differences
      const origModules = selectedCompany.value.modules || []
      const promises = []
      
      const targetVet = editForm.value.modules.veterinaria
      const hasVet = origModules.includes('veterinaria')
      if (targetVet !== hasVet) {
        promises.push(axios.post('/saas-admin/modules/toggle', {
          company_id: selectedCompany.value.id,
          module_key: 'veterinaria',
          is_active: targetVet
        }))
      }

      const targetVid = editForm.value.modules.vidrieria
      const hasVid = origModules.includes('vidrieria')
      if (targetVid !== hasVid) {
        promises.push(axios.post('/saas-admin/modules/toggle', {
          company_id: selectedCompany.value.id,
          module_key: 'vidrieria',
          is_active: targetVid
        }))
      }

      const targetFac = editForm.value.modules.facturacion
      const hasFac = origModules.includes('facturacion')
      if (targetFac !== hasFac) {
        promises.push(axios.post('/saas-admin/modules/toggle', {
          company_id: selectedCompany.value.id,
          module_key: 'facturacion',
          is_active: targetFac
        }))
      }

      if (promises.length > 0) {
        await Promise.all(promises)
      }

      showModal.value = false
      fetchData()
    }
  } catch (err) {
    alert('Error al actualizar la suscripción y módulos')
  }
}

const deleteCompany = async (company: any) => {
  if (!confirm(`¿Está seguro de que desea eliminar la empresa "${company.razon_social}" y todos sus usuarios y sucursales asociados? Esta acción no se puede deshacer.`)) {
    return
  }
  try {
    const res = await axios.delete(`/saas-admin/companies/${company.id}`)
    if (res.data.success) {
      alert(res.data.message || 'Empresa eliminada correctamente')
      fetchData()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al eliminar la empresa')
  }
}

onMounted(() => {
  fetchData()
  loadGlobalSettings()
})
</script>
