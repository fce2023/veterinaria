<template>
  <div class="overlay" :class="{ show: isSidebarOpen }" @click="isSidebarOpen = false"></div>

  <div class="app-layout">
    <!-- ─── Sidebar ─── -->
    <aside class="sidebar" :class="{ open: isSidebarOpen }">
      <div class="sidebar-logo">
        <div class="logo-row">
          <div class="logo-mark"><i class="ti ti-shield-check"></i></div>
          <div>
            <div class="logo-text">ERP CORE</div>
            <div class="logo-version">v1.0.0 PREMIUM</div>
          </div>
        </div>
      </div>

      <div class="user-card">
        <div class="avatar">{{ authStore.user?.nombre?.charAt(0) || 'U' }}</div>
        <div class="user-info">
          <div class="user-name">{{ authStore.user?.nombre }}</div>
          <div class="user-role">{{ authStore.user?.role_type }}</div>
        </div>
        <div class="online-dot"></div>
      </div>

      <nav class="nav">
        <div class="nav-section">
          <div class="nav-label">PRINCIPAL</div>
          
          <div v-if="!authStore.isCashier" 
               class="nav-item" :class="{ active: currentSection === 'stats' }"
               @click="selectSection('stats')">
            <i class="ti ti-chart-bar"></i><span>Dashboard</span>
          </div>
          
          <div v-if="authStore.isCompanyAdmin"
               class="nav-item" :class="{ active: currentSection === 'branches' }"
               @click="selectSection('branches')">
            <i class="ti ti-building-store"></i><span>Sucursales</span>
          </div>

          <div v-if="authStore.isCompanyAdmin || authStore.isBranchAdmin"
               class="nav-item" :class="{ active: currentSection === 'users' }"
               @click="selectSection('users')">
            <i class="ti ti-users"></i><span>Personal y Roles</span>
          </div>
        </div>

        <div class="nav-section">
          <div class="nav-label">INVENTARIO</div>
          
          <div v-if="authStore.isCompanyAdmin"
               class="nav-item" :class="{ active: currentSection === 'categories' }"
               @click="selectSection('categories')">
            <i class="ti ti-tag"></i><span>Categorías</span>
          </div>
          
          <div class="nav-item" :class="{ active: currentSection === 'products' }"
               @click="selectSection('products')">
            <i class="ti ti-box"></i><span>Productos</span>
          </div>

          <div v-if="!authStore.isCashier"
               class="nav-item" :class="{ active: currentSection === 'kardex' }"
               @click="selectSection('kardex')">
            <i class="ti ti-arrows-exchange"></i><span>Movimientos</span>
          </div>
        </div>

        <div class="nav-section">
          <div class="nav-label">OPERACIONES</div>
          
          <div v-if="authStore.isCompanyAdmin"
               class="nav-item" :class="{ active: currentSection === 'suppliers' }"
               @click="selectSection('suppliers')">
            <i class="ti ti-truck"></i><span>Proveedores</span>
          </div>

          <div v-if="!authStore.isCashier"
               class="nav-item" :class="{ active: currentSection === 'purchases' }"
               @click="selectSection('purchases')">
            <i class="ti ti-shopping-cart"></i><span>Compras</span>
          </div>

          <div v-if="authStore.hasModule('facturacion')"
               class="nav-item" :class="{ active: currentSection === 'documents' }"
               @click="selectSection('documents', 'documents')">
            <i class="ti ti-file-description"></i><span>Comprobantes</span>
          </div>

          <div class="nav-item" :class="{ active: currentSection === 'customers' }"
               @click="selectSection('customers')">
            <i class="ti ti-user-circle"></i><span>Clientes</span>
          </div>

          <div class="nav-item" :class="{ active: currentSection === 'sales' }"
               @click="selectSection('sales')">
            <i class="ti ti-device-desktop"></i><span>POS / Ventas</span>
          </div>

          <div class="nav-item" :class="{ active: currentSection === 'cash' }"
               @click="selectSection('cash')">
            <i class="ti ti-cash"></i><span>Caja / Arqueo</span>
          </div>
        </div>

        <div v-if="authStore.isCompanyAdmin" class="nav-section">
          <div class="nav-label">SISTEMA</div>
          
          <div class="nav-item" :class="{ active: currentSection === 'business' }"
               @click="selectSection('business')">
            <i class="ti ti-settings"></i><span>Configuración</span>
          </div>

          <div v-if="authStore.hasModule('facturacion')"
               class="nav-item" :class="{ active: currentSection === 'billing' }"
               @click="selectSection('billing', 'config')">
            <i class="ti ti-file-invoice"></i><span>Facturación API</span>
          </div>
        </div>
      </nav>

      <div class="sidebar-footer">
        <button @click="handleLogout" class="btn-logout">
          <i class="ti ti-logout"></i>
          <span>Cerrar Sesión</span>
        </button>
      </div>
    </aside>

    <!-- ─── Main ─── -->
    <main class="main">
      <header class="topbar">
        <button class="menu-btn" @click="isSidebarOpen = true"><i class="ti ti-menu-2"></i></button>
        <div class="topbar-left">
          <div class="topbar-title">{{ companyName }}</div>
          <div class="topbar-sub">
            <div class="branch-dot"></div>
            <span class="branch-label">Sucursal activa</span>
          </div>
        </div>
        <div class="topbar-right">
          <BranchSelector />
          <div class="plan-chip">PROFESSIONAL SAAS</div>
        </div>
      </header>

      <div class="content-view">
        <StatsSection v-if="currentSection === 'stats'" @navigate="selectSection($event)" />
        <CategoriesSection v-if="currentSection === 'categories'" />
        <ProductsSection v-if="currentSection === 'products'" />
        <CustomersSection v-if="currentSection === 'customers'" />
        <SuppliersSection v-if="currentSection === 'suppliers'" />
        <PurchasesSection v-if="currentSection === 'purchases'" />
        <KardexSection v-if="currentSection === 'kardex'" />
        <SalesSection v-if="currentSection === 'sales'" @navigate="selectSection($event)" />
        <CashSection v-if="currentSection === 'cash'" />
        <BillingSection v-if="currentSection === 'billing' || currentSection === 'documents'" 
                        :initialTab="currentBillingTab" 
                        :key="currentBillingTab"
                        @tabChange="handleBillingTabChange" />
        <BusinessSection v-if="currentSection === 'business'" />
        <BranchesSection v-if="currentSection === 'branches'" />
        <UsersSection v-if="currentSection === 'users'" />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '../store/auth'

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
const currentSection = ref(localStorage.getItem('erp_last_section') || (authStore.isCashier ? 'sales' : 'stats'))
const currentBillingTab = ref(localStorage.getItem('erp_last_billing_tab') || (currentSection.value === 'documents' ? 'documents' : 'config'))
const isSidebarOpen = ref(false)

const companyName = computed(() => authStore.user?.company?.nombre_comercial || 'Mi Empresa')

onMounted(async () => {
  await authStore.fetchUser()
  // Ensure cashier doesn't stay in restricted sections on refresh
  if (authStore.isCashier && !['sales', 'cash', 'customers', 'products'].includes(currentSection.value)) {
    currentSection.value = 'sales'
  }
})

function selectSection(section: string, billingTab: string = 'config') {
  currentSection.value = section
  currentBillingTab.value = billingTab
  localStorage.setItem('erp_last_section', section)
  
  // Persist billing tab if we're in billing/documents
  if (section === 'billing' || section === 'documents') {
    localStorage.setItem('erp_last_billing_tab', billingTab)
  }
  
  isSidebarOpen.value = false
}

function handleBillingTabChange(tab: string) {
  currentBillingTab.value = tab
  localStorage.setItem('erp_last_billing_tab', tab)
}

function handleLogout() {
  localStorage.removeItem('erp_last_section')
  localStorage.removeItem('erp_last_billing_tab')
  authStore.logout()
}
</script>

<style>
:root {
  --bg: #f8fafc;
  --surface: #ffffff;
  --surface2: #f1f5f9;
  --border: rgba(0,0,0,0.06);
  --border-hover: rgba(0,0,0,0.1);
  --accent: #4f46e5;
  --accent-dim: rgba(79, 70, 229, 0.1);
  --accent-hover: #4338ca;
  --green: #10b981;
  --green-dim: rgba(16, 185, 129, 0.12);
  --red: #ef4444;
  --red-dim: rgba(239, 68, 68, 0.12);
  --amber: #f59e0b;
  --amber-dim: rgba(245, 158, 11, 0.12);
  --text: #0f172a;
  --text2: #64748b;
  --text3: #94a3b8;
  --radius: 12px;
  --radius-sm: 8px;
  --sidebar-w: 240px;
  --transition: 0.18s ease;
}

/* Sidebar context */
.sidebar {
  --surface: #0f172a;
  --surface2: #1e293b;
  --text: #f8fafc;
  --text2: #94a3b8;
  --text3: #475569;
  --border: rgba(255,255,255,0.05);
  --accent: #6366f1;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: 'Inter', system-ui, sans-serif;
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
  line-height: 1.5;
  overflow: hidden;
}

/* ─── Layout ─── */
.app-layout { display: flex; height: 100vh; overflow: hidden; }

/* ─── Sidebar ─── */
.sidebar {
  width: var(--sidebar-w);
  flex-shrink: 0;
  background: var(--surface);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: transform var(--transition);
  z-index: 100;
}

.sidebar-logo {
  padding: 22px 18px 18px;
  border-bottom: 1px solid var(--border);
}
.logo-row { display: flex; align-items: center; gap: 10px; }
.logo-mark {
  width: 32px; height: 32px; border-radius: 9px;
  background: var(--accent);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 15px;
  flex-shrink: 0;
  box-shadow: 0 0 18px rgba(99, 102, 241, 0.4);
}
.logo-text { font-size: 13px; font-weight: 600; letter-spacing: 0.04em; color: var(--text); }
.logo-version { font-size: 10px; color: var(--accent); letter-spacing: 0.08em; margin-top: 1px; }

.user-card {
  padding: 14px 16px;
  display: flex; align-items: center; gap: 10px;
  border-bottom: 1px solid var(--border);
}
.avatar {
  width: 34px; height: 34px; border-radius: 50%;
  background: var(--accent-dim);
  border: 1.5px solid var(--accent);
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: 600; color: var(--accent);
  flex-shrink: 0;
}
.user-info { flex: 1; min-width: 0; }
.user-name { font-size: 12.5px; font-weight: 500; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.user-role { font-size: 10px; color: var(--text2); letter-spacing: 0.06em; }
.online-dot { width: 7px; height: 7px; background: var(--green); border-radius: 50%; flex-shrink: 0; box-shadow: 0 0 6px var(--green); }

.nav { flex: 1; overflow-y: auto; padding: 12px 10px; }
.nav::-webkit-scrollbar { width: 4px; }
.nav::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }

.nav-section { margin-bottom: 20px; }
.nav-label {
  font-size: 9.5px; font-weight: 600; letter-spacing: 0.14em;
  color: var(--text3); padding: 0 8px 8px;
}
.nav-item {
  display: flex; align-items: center; gap: 9px;
  padding: 9px 10px; border-radius: var(--radius-sm);
  color: var(--text2); font-size: 13px; cursor: pointer;
  transition: all var(--transition);
  margin-bottom: 2px;
}
.nav-item:hover { background: var(--surface2); color: var(--text); }
.nav-item.active { background: var(--accent-dim); color: var(--accent); font-weight: 600; }
.nav-item i { font-size: 16px; flex-shrink: 0; }

.sidebar-footer {
  padding: 14px 16px;
  border-top: 1px solid var(--border);
}
.btn-logout {
  width: 100%; display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border);
  background: var(--surface2); color: var(--text2); font-family: inherit;
  font-size: 12px; font-weight: 600; cursor: pointer; transition: all var(--transition);
}
.btn-logout:hover { background: var(--red-dim); color: var(--red); border-color: var(--red); }

/* ─── Main ─── */
.main { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; background: var(--bg); }

.topbar {
  padding: 14px 20px;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  display: flex; align-items: center; gap: 14px;
  flex-shrink: 0;
}
.topbar-left { flex: 1; min-width: 0; }
.topbar-title { font-size: 18px; font-weight: 800; color: var(--text); }
.topbar-sub { display: flex; align-items: center; gap: 6px; margin-top: 2px; }
.branch-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--green); }
.branch-label { font-size: 10px; font-weight: 700; color: var(--text2); text-transform: uppercase; letter-spacing: 0.05em; }

.topbar-right { display: flex; align-items: center; gap: 16px; }
.plan-chip {
  background: var(--accent-dim);
  color: var(--accent);
  border: 1px solid rgba(79, 70, 229, 0.2);
  font-size: 10px; font-weight: 700;
  padding: 4px 10px; border-radius: 20px;
  letter-spacing: 0.05em;
  white-space: nowrap;
}
.menu-btn {
  display: none; background: none; border: none; color: var(--text2);
  font-size: 22px; cursor: pointer; padding: 4px; line-height: 1;
}

.content-view { flex: 1; overflow-y: auto; padding: 24px; }
.content-view::-webkit-scrollbar { width: 6px; }
.content-view::-webkit-scrollbar-thumb { background: var(--border-hover); border-radius: 10px; }

/* Overlay */
.overlay { display: none; position: fixed; inset: 0; background: rgba(15, 23, 42, 0.5); z-index: 99; backdrop-filter: blur(4px); }
.overlay.show { display: block; }

/* Responsive */
@media (max-width: 900px) {
  .sidebar {
    position: fixed; left: 0; top: 0; bottom: 0; z-index: 150;
    width: 260px; transform: translateX(-100%);
  }
  .sidebar.open { transform: translateX(0); }
  .menu-btn { display: block; }
  .topbar-right { display: none; }
  .content-view { padding: 16px; }
}
</style>
