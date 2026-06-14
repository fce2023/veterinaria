<template>
  <div class="min-h-screen bg-slate-50 flex flex-col md:flex-row">
    <!-- Mobile Header (Visible only on small screens) -->
    <header class="md:hidden bg-white border-b border-slate-200 px-4 py-3 flex items-center justify-between sticky top-0 z-40">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center shadow-indigo-200 shadow-lg">
          <i class="pi pi-shield text-white text-sm"></i>
        </div>
        <h1 class="font-black text-sm tracking-tight text-slate-800 uppercase">ERP Core</h1>
      </div>
      <button @click="isSidebarOpen = !isSidebarOpen" class="p-2 text-slate-600 hover:bg-slate-100 rounded-full transition-all">
        <i :class="isSidebarOpen ? 'pi pi-times' : 'pi pi-bars'"></i>
      </button>
    </header>

    <!-- Sidebar / Drawer -->
    <aside 
      :class="[
        'fixed inset-y-0 left-0 z-50 w-72 bg-slate-900 text-white flex flex-col shadow-2xl transition-transform duration-300 ease-in-out md:relative md:translate-x-0',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]"
    >
      <div class="p-6 border-b border-slate-800 hidden md:flex items-center gap-3">
        <div class="w-10 h-10 bg-indigo-500 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20">
          <i class="pi pi-shield text-white text-xl"></i>
        </div>
        <div>
          <h1 class="font-black text-sm tracking-widest text-white uppercase">ERP Core</h1>
          <p class="text-[10px] text-indigo-400 font-mono font-bold">v1.0.0 PREMIUM</p>
        </div>
      </div>

      <!-- User Profile Card -->
      <div class="p-6 border-b border-slate-800">
        <div class="flex items-center gap-4">
          <div class="relative">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center font-black text-lg shadow-lg">
              {{ authStore.user?.nombre?.charAt(0) || 'U' }}
            </div>
            <span class="absolute -bottom-1 -right-1 w-4 h-4 bg-emerald-500 border-2 border-slate-900 rounded-full"></span>
          </div>
          <div class="overflow-hidden">
            <h2 class="text-sm font-bold text-white truncate">{{ authStore.user?.nombre }}</h2>
            <p class="text-[10px] text-slate-400 font-bold uppercase tracking-tighter">{{ authStore.user?.role_type }}</p>
          </div>
        </div>
      </div>

      <!-- Navigation Menu -->
      <nav class="flex-1 p-4 space-y-1.5 overflow-y-auto">
        <p class="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] px-4 mb-2 mt-4">Principal</p>
        
        <button 
          v-if="!authStore.isCashier"
          @click="selectSection('stats')" 
          :class="[navBtnClass, currentSection === 'stats' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-chart-bar text-sm"></i>
          <span>Dashboard</span>
        </button>

        <button 
          v-if="authStore.isCompanyAdmin"
          @click="selectSection('branches')" 
          :class="[navBtnClass, currentSection === 'branches' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-sitemap text-sm"></i>
          <span>Sucursales</span>
        </button>

        <button 
          v-if="authStore.isCompanyAdmin || authStore.isBranchAdmin"
          @click="selectSection('users')" 
          :class="[navBtnClass, currentSection === 'users' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-users text-sm"></i>
          <span>Personal y Roles</span>
        </button>

        <p class="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] px-4 mb-2 mt-6">Inventario</p>
        
        <button 
          v-if="authStore.isCompanyAdmin"
          @click="selectSection('categories')" 
          :class="[navBtnClass, currentSection === 'categories' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-tags text-sm"></i>
          <span>Categorías</span>
        </button>
        
        <button 
          @click="selectSection('products')" 
          :class="[navBtnClass, currentSection === 'products' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-box text-sm"></i>
          <span>Productos</span>
        </button>
        
        <button 
          v-if="!authStore.isCashier"
          @click="selectSection('kardex')" 
          :class="[navBtnClass, currentSection === 'kardex' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-history text-sm"></i>
          <span>Movimientos</span>
        </button>

        <p class="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] px-4 mb-2 mt-6">Operaciones</p>
        
        <button 
          v-if="authStore.isCompanyAdmin"
          @click="selectSection('suppliers')" 
          :class="[navBtnClass, currentSection === 'suppliers' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-truck text-sm"></i>
          <span>Proveedores</span>
        </button>
        
        <button 
          v-if="!authStore.isCashier"
          @click="selectSection('purchases')" 
          :class="[navBtnClass, currentSection === 'purchases' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-shopping-cart text-sm"></i>
          <span>Compras</span>
        </button>
        
        <button 
          @click="selectSection('customers')" 
          :class="[navBtnClass, currentSection === 'customers' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-users text-sm"></i>
          <span>Clientes</span>
        </button>
        
        <button 
          @click="selectSection('sales')" 
          :class="[navBtnClass, currentSection === 'sales' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-desktop text-sm"></i>
          <span>POS / Ventas</span>
        </button>

        <button 
          @click="selectSection('cash')" 
          :class="[navBtnClass, currentSection === 'cash' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-wallet text-sm"></i>
          <span>Caja / Arqueo</span>
        </button>

        <!-- Sistema Section -->
        <p v-if="authStore.isCompanyAdmin" class="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] px-4 mb-2 mt-6">Sistema</p>

        <button 
          v-if="authStore.isCompanyAdmin"
          @click="selectSection('business')" 
          :class="[navBtnClass, currentSection === 'business' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-building text-sm"></i>
          <span>Configuración</span>
        </button>

        <button 
          v-if="authStore.isCompanyAdmin && authStore.hasModule('facturacion')"
          @click="selectSection('billing')" 
          :class="[navBtnClass, currentSection === 'billing' ? navBtnActive : navBtnInactive]"
        >
          <i class="pi pi-file-export text-sm"></i>
          <span>Facturación</span>
        </button>
      </nav>

      <!-- Logout Footer -->
      <div class="p-6 border-t border-slate-800">
        <button
          @click="handleLogout"
          class="w-full flex items-center justify-center gap-2 py-3 px-4 bg-slate-800 hover:bg-red-900/40 text-slate-300 hover:text-red-200 rounded-xl text-xs font-black transition-all border border-slate-700 hover:border-red-900/50"
        >
          <i class="pi pi-sign-out"></i>
          <span>CERRAR SESIÓN</span>
        </button>
      </div>
    </aside>

    <!-- Overlay for mobile sidebar -->
    <div 
      v-if="isSidebarOpen" 
      @click="isSidebarOpen = false" 
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-40 md:hidden"
    ></div>

    <!-- Main Content Area -->
    <main class="flex-1 flex flex-col min-w-0 pb-20 md:pb-0 overflow-y-auto max-h-screen">
      <!-- Navbar / Top Header (Desktop Only) -->
      <header class="hidden md:flex bg-white border-b border-slate-200 px-8 py-5 items-center justify-between sticky top-0 z-30">
        <div class="flex items-center gap-6">
          <div class="flex flex-col">
            <h2 class="text-xl font-black text-slate-900 leading-tight">{{ companyName }}</h2>
            <div class="flex items-center gap-2 mt-1">
              <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
              <p class="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Sucursal Activa</p>
            </div>
          </div>
        </div>
        
        <div class="flex items-center gap-4">
          <BranchSelector />
          <div class="h-8 w-px bg-slate-200 mx-2"></div>
          <div class="flex flex-col items-end">
            <p class="text-[9px] text-slate-400 font-black uppercase tracking-tighter">Plan de Servicio</p>
            <p class="text-xs font-black text-indigo-600">PROFESSIONAL SAAS</p>
          </div>
        </div>
      </header>

      <!-- Page Title (Mobile) -->
      <div class="md:hidden px-4 pt-6 pb-2">
        <h2 class="text-2xl font-black text-slate-900">{{ getSectionLabel(currentSection) }}</h2>
      </div>

      <!-- Content Views -->
      <section class="p-4 md:p-8 flex-1 animate-fade-in">
        <StatsSection v-if="currentSection === 'stats'" @navigate="currentSection = $event" />
        <CategoriesSection v-if="currentSection === 'categories'" />
        <ProductsSection v-if="currentSection === 'products'" />
        <CustomersSection v-if="currentSection === 'customers'" />
        <SuppliersSection v-if="currentSection === 'suppliers'" />
        <PurchasesSection v-if="currentSection === 'purchases'" />
        <KardexSection v-if="currentSection === 'kardex'" />
        <SalesSection v-if="currentSection === 'sales'" @navigate="currentSection = $event" />
        <CashSection v-if="currentSection === 'cash'" />
        <BillingSection v-if="currentSection === 'billing'" />
        <BusinessSection v-if="currentSection === 'business'" />
        <BranchesSection v-if="currentSection === 'branches'" />
        <UsersSection v-if="currentSection === 'users'" />
      </section>

      <!-- Mobile Bottom Tab Bar -->
      <nav class="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-slate-200 px-2 py-3 flex items-center justify-around z-40 shadow-[0_-4px_20px_rgba(0,0,0,0.05)]">
        <button @click="selectSection('stats')" :class="['flex flex-col items-center justify-center p-2 rounded-2xl transition-all duration-300', currentSection === 'stats' ? 'text-indigo-600 transform -translate-y-1' : 'text-slate-400']">
          <i class="pi pi-home text-xl"></i>
          <span v-if="currentSection === 'stats'" class="w-1 h-1 bg-indigo-600 rounded-full mt-1"></span>
        </button>
        <button @click="selectSection('products')" :class="['flex flex-col items-center justify-center p-2 rounded-2xl transition-all duration-300', currentSection === 'products' ? 'text-indigo-600 transform -translate-y-1' : 'text-slate-400']">
          <i class="pi pi-box text-xl"></i>
          <span v-if="currentSection === 'products'" class="w-1 h-1 bg-indigo-600 rounded-full mt-1"></span>
        </button>
        <button @click="selectSection('sales')" :class="['flex flex-col items-center justify-center p-2 rounded-2xl transition-all duration-300', currentSection === 'sales' ? 'text-indigo-600 transform -translate-y-1' : 'text-slate-400']">
          <i class="pi pi-desktop text-xl"></i>
          <span v-if="currentSection === 'sales'" class="w-1 h-1 bg-indigo-600 rounded-full mt-1"></span>
        </button>
        <button @click="selectSection('customers')" :class="['flex flex-col items-center justify-center p-2 rounded-2xl transition-all duration-300', currentSection === 'customers' ? 'text-indigo-600 transform -translate-y-1' : 'text-slate-400']">
          <i class="pi pi-users text-xl"></i>
          <span v-if="currentSection === 'customers'" class="w-1 h-1 bg-indigo-600 rounded-full mt-1"></span>
        </button>
        <button @click="isSidebarOpen = true" class="flex flex-col items-center justify-center p-2 rounded-2xl text-slate-400">
          <i class="pi pi-bars text-xl"></i>
        </button>
      </nav>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

// Component Sections
import StatsSection from '../components/StatsSection.vue'
import CategoriesSection from '../components/CategoriesSection.vue'
import ProductsSection from '../components/ProductsSection.vue'
import CustomersSection from '../components/CustomersSection.vue'
import SuppliersSection from '../components/SuppliersSection.vue'
import PurchasesSection from '../components/PurchasesSection.vue'
import KardexSection from '../components/KardexSection.vue'
import SalesSection from '../components/SalesSection.vue'
import BillingSection from '../components/BillingSection.vue'
import BusinessSection from '../components/BusinessSection.vue'
import CashSection from '../components/CashSection.vue'
import BranchesSection from '../components/BranchesSection.vue'
import UsersSection from '../components/UsersSection.vue'
import BranchSelector from '../components/BranchSelector.vue'

const authStore = useAuthStore()
const currentSection = ref('stats')
const isSidebarOpen = ref(false)

const companyName = computed(() => authStore.user?.company?.nombre_comercial || 'Mi Empresa')

// CSS Classes for cleaner template
const navBtnClass = 'w-full flex items-center gap-3 px-4 py-3 rounded-xl text-xs font-bold tracking-tight transition-all duration-200'
const navBtnActive = 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/30 transform scale-[1.02]'
const navBtnInactive = 'text-slate-400 hover:bg-slate-800/50 hover:text-slate-200'

onMounted(async () => {
  await authStore.fetchUser()
})

function selectSection(section: string) {
  currentSection.value = section
  isSidebarOpen.value = false
}

function getSectionLabel(section: string) {
  const labels: any = {
    'stats': 'Dashboard',
    'branches': 'Sucursales',
    'categories': 'Categorías',
    'products': 'Productos',
    'kardex': 'Movimientos',
    'suppliers': 'Proveedores',
    'purchases': 'Compras',
    'customers': 'Clientes',
    'sales': 'Ventas / POS',
    'cash': 'Caja / Arqueo',
    'billing': 'Facturación Electrónica',
    'business': 'Configuración del Negocio',
    'users': 'Personal y Roles'
  }
  return labels[section] || 'Inicio'
}

function handleLogout() {
  authStore.logout()
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap');

:root {
  font-family: 'Inter', sans-serif;
}

.animate-fade-in {
  animation: fadeIn 0.3s ease-out forwards;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Custom scrollbar for sidebar */
aside::-webkit-scrollbar {
  width: 4px;
}
aside::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 10px;
}
</style>
