<template>
  <div class="flex flex-col h-full space-y-4 md:space-y-6">
    <!-- Header with Tabs -->
    <div class="flex items-center justify-between bg-white p-2 md:p-3 rounded-3xl border border-slate-200 shadow-sm sticky top-0 z-20">
      <div class="flex gap-1 md:gap-2">
        <button
          @click="activeSubTab = 'pos'; loadStocks()"
          :class="[tabClass, activeSubTab === 'pos' ? tabActive : tabInactive]"
        >
          <i class="pi pi-desktop md:mr-2"></i>
          <span class="hidden md:inline">Venta Nueva</span>
          <span class="md:hidden">POS</span>
        </button>
        <button
          @click="activeSubTab = 'history'; loadSales()"
          :class="[tabClass, activeSubTab === 'history' ? tabActive : tabInactive]"
        >
          <i class="pi pi-history md:mr-2"></i>
          <span class="hidden md:inline">Historial</span>
          <span class="md:hidden">Histor.</span>
        </button>
      </div>
      <div v-if="activeSubTab === 'pos'" class="px-4 py-2 bg-indigo-50 text-indigo-700 rounded-2xl text-[10px] md:text-xs font-black uppercase tracking-widest hidden sm:flex items-center gap-2">
        <span class="w-2 h-2 rounded-full bg-indigo-500 animate-pulse"></span>
        Caja Abierta
      </div>
    </div>

    <!-- Mode 1: POS Screen -->
    <div v-if="activeSubTab === 'pos'" class="flex flex-col lg:flex-row gap-6 h-full flex-1 min-h-0">
      
      <!-- Catalog Area -->
      <div class="flex-[1.5] flex flex-col space-y-4 min-h-0">
        <!-- Search & Customer Select -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div class="relative group">
            <i class="pi pi-search absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"></i>
            <input 
              v-model="productSearch"
              type="text" 
              placeholder="Buscar producto..."
              class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none transition-all shadow-sm"
            />
          </div>
          <div class="relative">
            <i class="pi pi-user absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"></i>
            <select
              v-model="saleForm.customer_id"
              class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none transition-all shadow-sm appearance-none"
            >
              <option value="">Cliente: Público General</option>
              <option v-for="cust in customers" :key="cust.id" :value="cust.id">{{ cust.nombre }}</option>
            </select>
          </div>
        </div>

        <!-- Product Grid -->
        <div class="flex-1 overflow-y-auto pr-1 grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-3 md:gap-4 pb-20 lg:pb-0">
          <button 
            v-for="item in filteredStocks" 
            :key="item.product_id"
            @click="quickAddItem(item)"
            :disabled="item.stock_actual <= 0"
            class="flex flex-col bg-white p-3 md:p-4 rounded-[2rem] border border-slate-200 shadow-sm hover:border-indigo-400 hover:shadow-lg transition-all active:scale-95 group relative text-left"
          >
            <div class="flex-1">
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest block mb-1">{{ item.category_nombre || 'General' }}</span>
              <h4 class="text-xs md:text-sm font-black text-slate-900 leading-tight line-clamp-2 mb-2 group-hover:text-indigo-600">{{ item.product_nombre }}</h4>
              <div class="flex items-center justify-between mt-auto">
                <span class="text-sm md:text-lg font-black text-slate-950">S/. {{ item.precio_venta.toFixed(2) }}</span>
                <span :class="['text-[9px] font-black px-1.5 py-0.5 rounded-lg', item.stock_actual > 5 ? 'bg-emerald-50 text-emerald-600' : 'bg-red-50 text-red-600']">
                  {{ item.stock_actual }} un.
                </span>
              </div>
            </div>
            <!-- Quantity hint if already in cart -->
            <div v-if="getItemQtyInCart(item.product_id)" class="absolute -top-2 -right-2 w-7 h-7 bg-indigo-600 text-white rounded-full flex items-center justify-center text-[10px] font-black shadow-lg shadow-indigo-200">
              {{ getItemQtyInCart(item.product_id) }}
            </div>
          </button>
        </div>
      </div>

      <!-- Cart / Summary Area (Floating on mobile, sidebar on desktop) -->
      <div 
        :class="[
          'lg:flex-1 bg-white lg:rounded-[2.5rem] border border-slate-200 shadow-2xl lg:shadow-sm flex flex-col overflow-hidden transition-all duration-300 z-30',
          isCartOpen ? 'fixed inset-0 z-50 rounded-none' : 'hidden lg:flex'
        ]"
      >
        <!-- Cart Header -->
        <div class="p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h3 class="text-lg font-black text-slate-900 flex items-center gap-2">
              <i class="pi pi-shopping-cart text-indigo-600"></i>
              <span>Carrito de Venta</span>
            </h3>
            <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-1">{{ saleForm.items.length }} ítems seleccionados</p>
          </div>
          <button @click="isCartOpen = false" class="lg:hidden w-10 h-10 flex items-center justify-center bg-white border border-slate-200 rounded-2xl text-slate-400">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <!-- Cart Items -->
        <div class="flex-1 overflow-y-auto p-4 space-y-3">
          <div v-for="(item, idx) in saleForm.items" :key="idx" class="bg-white p-4 rounded-3xl border border-slate-100 shadow-sm flex items-center gap-4 group">
            <div class="w-10 h-10 bg-slate-50 rounded-2xl flex items-center justify-center text-slate-400 font-black text-sm">
              {{ item.cantidad }}
            </div>
            <div class="flex-1">
              <h5 class="text-xs font-black text-slate-900 leading-tight">{{ getProductName(item.product_id) }}</h5>
              <p class="text-[10px] text-slate-400 font-bold mt-1">S/. {{ item.precio_unitario.toFixed(2) }} c/u</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-black text-slate-950">S/. {{ (item.cantidad * item.precio_unitario).toFixed(2) }}</p>
              <button @click="removeItem(idx)" class="text-red-400 hover:text-red-600 transition-colors opacity-0 group-hover:opacity-100">
                <i class="pi pi-trash text-xs"></i>
              </button>
            </div>
          </div>

          <!-- Empty Cart -->
          <div v-if="saleForm.items.length === 0" class="h-full flex flex-col items-center justify-center text-center p-12 gap-4">
            <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center text-slate-200">
              <i class="pi pi-shopping-bag text-4xl"></i>
            </div>
            <div>
              <p class="text-sm font-black text-slate-800 uppercase tracking-tight">El carrito está vacío</p>
              <p class="text-xs text-slate-400 font-medium mt-1">Selecciona productos del catálogo para comenzar la venta.</p>
            </div>
          </div>
        </div>

        <!-- Summary & Checkout -->
        <div class="p-6 bg-slate-900 text-white rounded-t-[2.5rem] lg:rounded-none space-y-4">
          <div class="flex justify-between items-center px-2">
            <span class="text-xs font-bold text-slate-400 uppercase tracking-[0.2em]">Total a Pagar</span>
            <span class="text-3xl font-black tracking-tighter">S/. {{ calculatedTotal.toFixed(2) }}</span>
          </div>
          <div class="flex gap-3">
            <button @click="clearForm" class="flex-1 py-4 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all">
              Limpiar
            </button>
            <button 
              @click="submitSale"
              :disabled="saleForm.items.length === 0"
              class="flex-[2] py-4 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-700 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-xl shadow-indigo-900/40 transition-all flex items-center justify-center gap-2"
            >
              <i class="pi pi-check-circle text-base"></i>
              Procesar Pago
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Mode 2: Sales History (Modern List) -->
    <div v-else class="flex-1 overflow-y-auto space-y-4 pb-20">
      <div v-for="sale in sales" :key="sale.id" class="bg-white p-5 rounded-[2.5rem] border border-slate-200 shadow-sm flex items-center gap-4 hover:border-indigo-300 transition-all">
        <div class="w-14 h-14 bg-indigo-50 text-indigo-600 rounded-3xl flex items-center justify-center flex-shrink-0">
          <i class="pi pi-receipt text-xl"></i>
        </div>
        <div class="flex-1">
          <div class="flex items-center justify-between mb-1">
            <h4 class="text-sm font-black text-slate-900">VENTA-{{ sale.id.substring(0, 8).toUpperCase() }}</h4>
            <span class="text-sm font-black text-slate-950">S/. {{ sale.total.toFixed(2) }}</span>
          </div>
          <p class="text-xs text-slate-500 font-bold tracking-tight">{{ sale.customer?.nombre || 'Público General' }}</p>
          <p class="text-[10px] text-slate-400 font-medium mt-1 uppercase">{{ formatDate(sale.created_at) }}</p>
        </div>
        <button @click="viewDetails(sale.id)" class="w-10 h-10 flex items-center justify-center bg-slate-50 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-2xl transition-all">
          <i class="pi pi-chevron-right text-xs"></i>
        </button>
      </div>
      
      <div v-if="sales.length === 0" class="text-center py-20 text-slate-300">
        <i class="pi pi-inbox text-5xl mb-4"></i>
        <p class="text-sm font-black uppercase">Sin ventas registradas</p>
      </div>
    </div>

    <!-- Mobile Floating Cart Button -->
    <button 
      v-if="activeSubTab === 'pos' && !isCartOpen" 
      @click="isCartOpen = true"
      class="lg:hidden fixed bottom-24 right-6 w-16 h-16 bg-indigo-600 text-white rounded-full shadow-2xl shadow-indigo-400 z-40 flex items-center justify-center animate-bounce-slow"
    >
      <i class="pi pi-shopping-cart text-xl"></i>
      <span v-if="saleForm.items.length > 0" class="absolute -top-1 -right-1 w-6 h-6 bg-red-500 text-white rounded-full flex items-center justify-center text-[10px] font-black border-2 border-white">
        {{ saleForm.items.length }}
      </span>
    </button>

    <!-- Details Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-slate-900/70 backdrop-blur-md flex items-end sm:items-center justify-center p-0 sm:p-4 z-[70] animate-fade-in">
      <div @click="showModal = false" class="absolute inset-0"></div>
      <div class="relative bg-white rounded-t-[3rem] sm:rounded-[3rem] w-full max-w-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <div class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h3 class="text-xl font-black text-slate-900 tracking-tight">Venta #{{ selectedSale?.sale?.id.substring(0, 8).toUpperCase() }}</h3>
            <p class="text-[10px] text-indigo-500 font-black uppercase tracking-widest mt-1">Resumen detallado</p>
          </div>
          <button @click="showModal = false" class="w-10 h-10 flex items-center justify-center bg-white border border-slate-200 text-slate-400 rounded-2xl">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-8 space-y-8">
          <!-- Info Grid -->
          <div class="grid grid-cols-2 gap-8">
            <div class="space-y-1">
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Cliente</p>
              <p class="text-sm font-black text-slate-900">{{ selectedSale?.sale?.customer?.nombre || 'Público General' }}</p>
              <p class="text-xs text-slate-500">{{ selectedSale?.sale?.customer?.tipo_documento }} {{ selectedSale?.sale?.customer?.numero_documento }}</p>
            </div>
            <div class="text-right space-y-1">
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Emisión</p>
              <p class="text-sm font-black text-slate-900">{{ formatDate(selectedSale?.sale?.created_at) }}</p>
            </div>
          </div>

          <!-- Items Table -->
          <div class="space-y-4">
            <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Ítems</p>
            <div class="space-y-3">
              <div v-for="item in selectedSale?.items" :key="item.id" class="flex items-center gap-4 py-3 border-b border-slate-50">
                <div class="w-10 h-10 bg-slate-50 rounded-2xl flex items-center justify-center text-slate-400 font-black text-xs">
                  {{ item.cantidad }}
                </div>
                <div class="flex-1">
                  <h5 class="text-xs font-black text-slate-900 leading-tight">{{ item.product_nombre }}</h5>
                </div>
                <div class="text-right">
                  <p class="text-xs font-black text-slate-950">S/. {{ item.total.toFixed(2) }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-8 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
           <div class="flex flex-col">
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Total Transacción</p>
              <p class="text-2xl font-black text-slate-900 tracking-tighter">S/. {{ selectedSale?.sale?.total.toFixed(2) }}</p>
           </div>
           <button @click="showModal = false" class="px-8 py-3 bg-slate-900 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg active:scale-95 transition-all">
              Cerrar
           </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import axios from 'axios'

const activeSubTab = ref('pos')
const customers = ref<any[]>([])
const stocks = ref<any[]>([])
const sales = ref<any[]>([])
const productSearch = ref('')
const isCartOpen = ref(false)

// CSS Classes
const tabClass = 'flex items-center gap-2 px-6 py-3 rounded-2xl text-[10px] md:text-xs font-black uppercase tracking-widest transition-all'
const tabActive = 'bg-slate-900 text-white shadow-xl shadow-slate-200'
const tabInactive = 'text-slate-400 hover:bg-slate-50'

const saleForm = reactive({
  customer_id: '',
  items: [] as any[]
})

const showModal = ref(false)
const selectedSale = ref<any>(null)

const filteredStocks = computed(() => {
  const query = productSearch.value.toLowerCase()
  return stocks.value.filter(s => 
    s.product_nombre.toLowerCase().includes(query) || 
    (s.product_codigo && s.product_codigo.toLowerCase().includes(query))
  )
})

const calculatedTotal = computed(() => {
  return saleForm.items.reduce((sum, item) => sum + (item.cantidad * item.precio_unitario) - item.descuento, 0)
})

onMounted(() => {
  loadCustomers()
  loadStocks()
})

async function loadCustomers() {
  try {
    const res = await axios.get('/customers')
    if (res.data.success) customers.value = res.data.data
  } catch (err) {
    console.error('Error loading customers', err)
  }
}

async function loadStocks() {
  try {
    const res = await axios.get('/stocks')
    if (res.data.success) stocks.value = res.data.data || []
  } catch (err) {
    console.error('Error loading stocks', err)
  }
}

async function loadSales() {
  try {
    const res = await axios.get('/sales')
    if (res.data.success) sales.value = res.data.data || []
  } catch (err) {
    console.error('Error loading sales', err)
  }
}

function quickAddItem(stock: any) {
  const alreadyAdded = saleForm.items.find(item => item.product_id === stock.product_id)
  const currentQty = alreadyAdded ? alreadyAdded.cantidad : 0

  if (currentQty + 1 > stock.stock_actual) {
    alert('Stock insuficiente')
    return
  }

  if (alreadyAdded) {
    alreadyAdded.cantidad += 1
  } else {
    saleForm.items.push({
      product_id: stock.product_id,
      cantidad: 1,
      precio_unitario: stock.precio_venta,
      descuento: 0
    })
  }
}

function getItemQtyInCart(pId: string) {
  const item = saleForm.items.find(i => i.product_id === pId)
  return item ? item.cantidad : 0
}

function getProductName(pId: string) {
  const stockObj = stocks.value.find(s => s.product_id === pId)
  return stockObj ? stockObj.product_nombre : 'Producto'
}

function removeItem(idx: number) {
  saleForm.items.splice(idx, 1)
}

function clearForm() {
  saleForm.customer_id = ''
  saleForm.items = []
  isCartOpen.value = false
}

async function submitSale() {
  if (saleForm.items.length === 0) return

  try {
    const res = await axios.post('/sales', saleForm)
    if (res.data.success) {
      alert('¡Venta completada!')
      clearForm()
      await loadStocks()
      activeSubTab.value = 'history'
      loadSales()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error en la venta')
  }
}

async function viewDetails(id: string) {
  try {
    const res = await axios.get(`/sales/${id}`)
    if (res.data.success) {
      selectedSale.value = res.data.data
      showModal.value = true
    }
  } catch (err) {
    alert('Error al cargar detalle')
  }
}

function formatDate(dStr: string) {
  if (!dStr) return '-'
  const d = new Date(dStr)
  return d.toLocaleString('es-PE', { 
    day: '2-digit', 
    month: 'short', 
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.animate-bounce-slow {
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}
</style>
