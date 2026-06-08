<template>
  <div class="min-h-screen bg-slate-100 flex flex-col md:flex-row">
    <!-- Sidebar -->
    <aside class="w-full md:w-64 bg-slate-900 text-white flex flex-col shadow-lg">
      <div class="p-6 border-b border-slate-800 flex items-center gap-3">
        <i class="pi pi-shield text-sky-400 text-2xl"></i>
        <div>
          <h1 class="font-extrabold text-sm tracking-wide">ERP CORE</h1>
          <p class="text-xs text-sky-400 font-mono">v1.0.0 (SaaS)</p>
        </div>
      </div>

      <!-- User Profile Card -->
      <div class="p-4 border-b border-slate-800 bg-slate-800/40">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-full bg-sky-500/20 text-sky-400 flex items-center justify-center font-bold text-sm">
            {{ authStore.user?.nombre?.charAt(0) || 'U' }}
          </div>
          <div class="overflow-hidden">
            <h2 class="text-xs font-bold text-slate-200 truncate">{{ authStore.user?.nombre }}</h2>
            <p class="text-[10px] text-slate-400 truncate">{{ authStore.user?.email }}</p>
          </div>
        </div>
      </div>

      <!-- Navigation Menu -->
      <nav class="flex-1 p-4 space-y-1 overflow-y-auto max-h-[60vh]">
        <!-- Dashboard / Stats -->
        <button
          @click="currentSection = 'stats'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'stats' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-chart-bar"></i>
          <span>Dashboard</span>
        </button>

        <!-- Branches -->
        <button
          @click="currentSection = 'branches'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'branches' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-sitemap"></i>
          <span>Sucursales</span>
        </button>

        <!-- Categories & Brands -->
        <button
          @click="currentSection = 'categories'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'categories' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-tags"></i>
          <span>Categorías / Marcas</span>
        </button>

        <!-- Products -->
        <button
          @click="currentSection = 'products'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'products' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-box"></i>
          <span>Productos</span>
        </button>

        <!-- Stocks & Kardex -->
        <button
          @click="currentSection = 'kardex'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'kardex' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-history"></i>
          <span>Inventario / Kardex</span>
        </button>

        <!-- Suppliers -->
        <button
          @click="currentSection = 'suppliers'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'suppliers' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-truck"></i>
          <span>Proveedores</span>
        </button>

        <!-- Purchases -->
        <button
          @click="currentSection = 'purchases'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'purchases' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-shopping-cart"></i>
          <span>Compras</span>
        </button>

        <!-- Customers -->
        <button
          @click="currentSection = 'customers'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'customers' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-users"></i>
          <span>Clientes</span>
        </button>

        <!-- Sales / POS -->
        <button
          @click="currentSection = 'sales'"
          :class="['w-full flex items-center gap-3 px-4 py-2 rounded-lg text-xs font-semibold tracking-wide transition-all',
            currentSection === 'sales' ? 'bg-sky-600 text-white shadow-md' : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200']"
        >
          <i class="pi pi-desktop"></i>
          <span>Ventas / POS</span>
        </button>
      </nav>

      <!-- Logout Footer -->
      <div class="p-4 border-t border-slate-800">
        <button
          @click="handleLogout"
          class="w-full flex items-center justify-center gap-2 py-2 px-4 bg-red-950/40 hover:bg-red-900/40 border border-red-900/40 text-red-200 rounded-lg text-xs font-bold transition-all"
        >
          <i class="pi pi-sign-out"></i>
          <span>Cerrar Sesión</span>
        </button>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="flex-1 flex flex-col min-w-0">
      <!-- Navbar / Top Header -->
      <header class="bg-white border-b border-slate-200 px-6 py-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 class="text-xl font-extrabold text-slate-800">Panel de Control</h2>
          <p class="text-xs text-slate-500">
            Multi-Tenant / Empresa: <strong class="text-slate-700 font-semibold">{{ authStore.user?.company_id }}</strong>
          </p>
        </div>
        <div class="flex items-center gap-2 bg-sky-50 text-sky-700 py-1.5 px-3 rounded-full text-xs font-semibold border border-sky-100">
          <i class="pi pi-building"></i>
          <span>Sucursal de Acceso ID: {{ authStore.user?.branch_id }}</span>
        </div>
      </header>

      <!-- Content Views -->
      <section class="p-6 flex-1">
        <!-- View 0: DASHBOARD STATS -->
        <StatsSection v-if="currentSection === 'stats'" @navigate="currentSection = $event" />

        <!-- View 1: BRANCHES -->
        <div v-if="currentSection === 'branches'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- Create Branch Form -->
          <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-4">
            <h4 class="text-sm font-bold text-slate-800 border-b border-slate-100 pb-2 flex items-center gap-2">
              <i class="pi pi-plus text-sky-500"></i>
              <span>Nueva Sucursal</span>
            </h4>
            <form @submit.prevent="handleCreateBranch" class="space-y-3">
              <div>
                <label class="block text-xs text-slate-600 mb-1">Nombre</label>
                <input
                  v-model="branchForm.nombre"
                  type="text"
                  required
                  class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-sky-500 focus:border-sky-500"
                  placeholder="Ej. Sede San Borja"
                />
              </div>
              <div>
                <label class="block text-xs text-slate-600 mb-1">Dirección</label>
                <input
                  v-model="branchForm.direccion"
                  type="text"
                  class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-sky-500 focus:border-sky-500"
                  placeholder="Calle Las Flores 123"
                />
              </div>
              <div>
                <label class="block text-xs text-slate-600 mb-1">Teléfono</label>
                <input
                  v-model="branchForm.telefono"
                  type="text"
                  class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-sky-500 focus:border-sky-500"
                  placeholder="987654321"
                />
              </div>
              <button
                type="submit"
                class="w-full py-2 bg-sky-600 hover:bg-sky-500 text-white rounded-lg text-xs font-bold transition-all shadow-sm"
              >
                Agregar Sucursal
              </button>
            </form>
          </div>

          <!-- Branches List Table -->
          <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm lg:col-span-2 flex flex-col">
            <h4 class="text-sm font-bold text-slate-800 border-b border-slate-100 pb-2 mb-4">Lista de Sucursales</h4>
            <div class="overflow-x-auto flex-1">
              <table class="w-full text-left border-collapse">
                <thead>
                  <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
                    <th class="py-3 px-4">Nombre</th>
                    <th class="py-3 px-4">Dirección</th>
                    <th class="py-3 px-4">Teléfono</th>
                    <th class="py-3 px-4">Estado</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
                  <tr v-for="item in branches" :key="item.id" class="hover:bg-slate-50/50">
                    <td class="py-3 px-4 font-semibold text-slate-900">{{ item.nombre }}</td>
                    <td class="py-3 px-4 text-slate-500">{{ item.direccion || '-' }}</td>
                    <td class="py-3 px-4 text-slate-500">{{ item.telefono || '-' }}</td>
                    <td class="py-3 px-4">
                      <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-55/10 text-emerald-600 border border-emerald-200">
                        {{ item.estado }}
                      </span>
                    </td>
                  </tr>
                  <tr v-if="branches.length === 0">
                    <td colspan="4" class="text-center py-6 text-slate-400">No hay sucursales registradas</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- View 2: PRODUCTS -->
        <div v-if="currentSection === 'products'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- Create Product Form -->
          <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-4">
            <h4 class="text-sm font-bold text-slate-800 border-b border-slate-100 pb-2 flex items-center gap-2">
              <i class="pi pi-plus text-teal-500"></i>
              <span>Nuevo Producto</span>
            </h4>
            <form @submit.prevent="handleCreateProduct" class="space-y-3">
              <div>
                <label class="block text-xs text-slate-600 mb-1">Nombre</label>
                <input
                  v-model="productForm.nombre"
                  type="text"
                  required
                  class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-teal-500 focus:border-teal-500"
                  placeholder="Ej. Antipulgas Nexgard"
                />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs text-slate-600 mb-1">Código</label>
                  <input
                    v-model="productForm.codigo"
                    type="text"
                    class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-teal-500"
                    placeholder="P-001"
                  />
                </div>
                <div>
                  <label class="block text-xs text-slate-600 mb-1">Código Barras</label>
                  <input
                    v-model="productForm.codigo_barras"
                    type="text"
                    class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-teal-500"
                    placeholder="775..."
                  />
                </div>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs text-slate-600 mb-1">Precio Compra</label>
                  <input
                    v-model.number="productForm.precio_compra"
                    type="number"
                    step="0.01"
                    class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-teal-500"
                    placeholder="15.00"
                  />
                </div>
                <div>
                  <label class="block text-xs text-slate-600 mb-1">Precio Venta</label>
                  <input
                    v-model.number="productForm.precio_venta"
                    type="number"
                    step="0.01"
                    required
                    class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg text-xs focus:ring-teal-500"
                    placeholder="25.00"
                  />
                </div>
              </div>
              <button
                type="submit"
                class="w-full py-2 bg-teal-600 hover:bg-teal-500 text-white rounded-lg text-xs font-bold transition-all shadow-sm"
              >
                Agregar Producto
              </button>
            </form>
          </div>

          <!-- Products List Table -->
          <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm lg:col-span-2 flex flex-col">
            <h4 class="text-sm font-bold text-slate-800 border-b border-slate-100 pb-2 mb-4">Lista de Productos</h4>
            <div class="overflow-x-auto flex-1">
              <table class="w-full text-left border-collapse">
                <thead>
                  <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
                    <th class="py-3 px-4">Código</th>
                    <th class="py-3 px-4">Nombre</th>
                    <th class="py-3 px-4">Cost. Compra</th>
                    <th class="py-3 px-4">Precio Venta</th>
                    <th class="py-3 px-4">Estado</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
                  <tr v-for="item in products" :key="item.id" class="hover:bg-slate-50/50">
                    <td class="py-3 px-4 font-mono text-slate-500 text-[10px]">{{ item.codigo || '-' }}</td>
                    <td class="py-3 px-4 font-semibold text-slate-900">{{ item.nombre }}</td>
                    <td class="py-3 px-4 text-slate-500">S/. {{ item.precio_compra.toFixed(2) }}</td>
                    <td class="py-3 px-4 font-bold text-slate-900">S/. {{ item.precio_venta.toFixed(2) }}</td>
                    <td class="py-3 px-4">
                      <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-55/10 text-emerald-600 border border-emerald-200">
                        {{ item.estado }}
                      </span>
                    </td>
                  </tr>
                  <tr v-if="products.length === 0">
                    <td colspan="5" class="text-center py-6 text-slate-400">No hay productos registrados</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- View 3: CATEGORIES -->
        <CategoriesSection v-if="currentSection === 'categories'" />

        <!-- View 4: INVENTORY / KARDEX -->
        <KardexSection v-if="currentSection === 'kardex'" />

        <!-- View 5: SUPPLIERS -->
        <SuppliersSection v-if="currentSection === 'suppliers'" />

        <!-- View 6: PURCHASES -->
        <PurchasesSection v-if="currentSection === 'purchases'" />

        <!-- View 7: CUSTOMERS -->
        <CustomersSection v-if="currentSection === 'customers'" />

        <!-- View 8: SALES / POS -->
        <SalesSection v-if="currentSection === 'sales'" />
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

// Component Sections
import StatsSection from '../components/StatsSection.vue'
import CategoriesSection from '../components/CategoriesSection.vue'
import KardexSection from '../components/KardexSection.vue'
import SuppliersSection from '../components/SuppliersSection.vue'
import PurchasesSection from '../components/PurchasesSection.vue'
import CustomersSection from '../components/CustomersSection.vue'
import SalesSection from '../components/SalesSection.vue'

const authStore = useAuthStore()
const currentSection = ref('stats')

// Lists
const branches = ref<any[]>([])
const products = ref<any[]>([])

// Forms reactivity
const branchForm = reactive({
  nombre: '',
  direccion: '',
  telefono: '',
  estado: 'active'
})

const productForm = reactive({
  nombre: '',
  codigo: '',
  codigo_barras: '',
  precio_compra: 0,
  precio_venta: 0,
  estado: 'active'
})

onMounted(async () => {
  await authStore.fetchUser()
  await loadBranches()
  await loadProducts()
})

async function loadBranches() {
  try {
    const res = await axios.get('/branches')
    if (res.data.success) {
      branches.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading branches', err)
  }
}

async function loadProducts() {
  try {
    const res = await axios.get('/products')
    if (res.data.success) {
      products.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading products', err)
  }
}

async function handleCreateBranch() {
  try {
    const res = await axios.post('/branches', branchForm)
    if (res.data.success) {
      branches.value.push(res.data.data)
      // Reset form
      Object.assign(branchForm, {
        nombre: '',
        direccion: '',
        telefono: '',
        estado: 'active'
      })
    }
  } catch (err) {
    alert('Error al crear la sucursal')
  }
}

async function handleCreateProduct() {
  try {
    const res = await axios.post('/products', productForm)
    if (res.data.success) {
      products.value.push(res.data.data)
      // Reset form
      Object.assign(productForm, {
        nombre: '',
        codigo: '',
        codigo_barras: '',
        precio_compra: 0,
        precio_venta: 0,
        estado: 'active'
      })
    }
  } catch (err) {
    alert('Error al crear el producto')
  }
}

function handleLogout() {
  authStore.logout()
}
</script>

<style>
/* Smooth fade-in animation for stats */
.animate-fade-in {
  animation: fadeIn 0.4s ease-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
