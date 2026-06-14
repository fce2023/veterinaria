<template>
  <div class="flex flex-col h-full space-y-4 md:space-y-6">
    <!-- Header with Tabs -->
    <div class="flex items-center justify-between bg-white p-2 md:p-3 rounded-3xl border border-slate-200 shadow-sm sticky top-0 z-20">
      <div class="flex gap-1 md:gap-2">
        <button
          @click="activeSubTab = 'pos'; checkCashSession(); loadStocks()"
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
      <div v-if="activeSubTab === 'pos'" :class="['px-4 py-2 rounded-2xl text-[10px] md:text-xs font-black uppercase tracking-widest hidden sm:flex items-center gap-2', hasActiveCashSession ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700']">
        <span :class="['w-2 h-2 rounded-full animate-pulse', hasActiveCashSession ? 'bg-emerald-500' : 'bg-red-500']"></span>
        {{ hasActiveCashSession ? 'Caja Abierta' : 'Caja Cerrada' }}
      </div>
    </div>

    <!-- Warning if Cash is Closed -->
    <div v-if="activeSubTab === 'pos' && !hasActiveCashSession" class="flex-1 bg-white rounded-[2.5rem] border border-slate-200 shadow-sm p-8 flex flex-col items-center justify-center text-center my-8">
      <div class="w-16 h-16 bg-red-50 text-red-500 rounded-3xl flex items-center justify-center mb-4">
        <i class="pi pi-lock text-3xl"></i>
      </div>
      <h3 class="text-xl font-black text-slate-900 mb-2">Caja de Turno Cerrada</h3>
      <p class="text-xs text-slate-500 font-bold max-w-sm mb-6">Debes abrir una sesión de caja antes de realizar ventas. Esto permite llevar un control y arqueo adecuado de los ingresos.</p>
      <button 
        @click="emit('navigate', 'cash')"
        class="px-8 py-3.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-indigo-100 transition-all flex items-center gap-2"
      >
        <i class="pi pi-wallet"></i> Ir a Apertura de Caja
      </button>
    </div>

    <!-- Main Content Area -->
    <div v-if="activeSubTab === 'pos' && hasActiveCashSession" class="flex-1 min-h-0">
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 h-full items-start">
        
        <!-- Left Side: Product Catalog -->
        <div class="lg:col-span-7 xl:col-span-8 flex flex-col space-y-4 min-h-0">
          <!-- Search & Customer -->
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
          <div class="flex-1 overflow-y-auto pr-1 grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3 md:gap-4 pb-24 lg:pb-8">
            <button 
              v-for="item in filteredStocks" 
              :key="item.product_id"
              @click="quickAddItem(item)"
              :disabled="item.stock_actual <= 0"
              class="flex flex-col bg-white p-3 rounded-2xl border border-slate-200 shadow-sm hover:border-indigo-400 hover:shadow-lg hover:shadow-indigo-50 transition-all active:scale-95 group relative text-left h-full"
            >
              <div class="flex flex-col h-full">
                <div class="mb-2">
                  <span class="inline-block px-2 py-0.5 bg-slate-50 text-[8px] font-black text-slate-500 uppercase tracking-wider rounded-md mb-1">{{ item.category_nombre || 'General' }}</span>
                  <h4 class="text-xs font-black text-slate-900 leading-tight line-clamp-2 group-hover:text-indigo-600 transition-colors">{{ item.product_nombre }}</h4>
                </div>
                <div class="mt-auto pt-2 flex items-center justify-between">
                  <span class="text-sm font-black text-slate-950">S/. {{ item.precio_venta.toFixed(2) }}</span>
                  <span :class="['text-[8px] font-black px-1.5 py-0.5 rounded-lg border', item.stock_actual > 5 ? 'bg-emerald-50 text-emerald-600 border-emerald-100' : 'bg-red-50 text-red-600 border-red-100']">
                    {{ item.stock_actual.toFixed(2) }} {{ item.unidad_medida || 'un.' }}
                  </span>
                </div>
              </div>
              <div v-if="getItemQtyInCart(item.product_id)" class="absolute -top-1.5 -right-1.5 w-6 h-6 bg-indigo-600 text-white rounded-full flex items-center justify-center text-[10px] font-black shadow-lg ring-2 ring-white">
                {{ getItemQtyInCart(item.product_id) }}
              </div>
            </button>
          </div>
        </div>

        <!-- Right Side: Cart Summary -->
        <div 
          :class="[
            'lg:col-span-5 xl:col-span-4 bg-white lg:rounded-[2.5rem] border border-slate-200 flex flex-col overflow-hidden transition-all duration-300 shadow-2xl lg:shadow-sm lg:sticky lg:top-24 max-h-[calc(100vh-8rem)]',
            isCartOpen ? 'fixed inset-0 z-50 rounded-none' : 'hidden lg:flex'
          ]"
        >
          <!-- Header -->
          <div class="p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-indigo-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-200">
                <i class="pi pi-shopping-cart text-white"></i>
              </div>
              <div>
                <h3 class="text-base font-black text-slate-900">Carrito</h3>
                <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest">{{ saleForm.items.length }} ítems</p>
              </div>
            </div>
            <button @click="isCartOpen = false" class="lg:hidden w-10 h-10 flex items-center justify-center bg-white border border-slate-200 rounded-2xl text-slate-400">
              <i class="pi pi-times"></i>
            </button>
          </div>

          <!-- Items Scroll Area -->
          <div class="flex-1 overflow-y-auto p-4 space-y-4 bg-slate-50/30">
            <div v-for="(item, idx) in saleForm.items" :key="idx" class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm flex flex-col gap-3 group transition-all hover:border-indigo-200">
              <div class="flex items-start gap-4">
                <!-- Mini Badge -->
                <div v-if="item.is_dimensional" class="flex flex-col items-center justify-center min-w-[3.2rem] h-12 bg-slate-900 text-white rounded-xl shadow-md">
                   <span class="text-[10px] font-black">{{ item.cantidad.toFixed(2) }}</span>
                   <span class="text-[7px] font-bold uppercase text-slate-400">{{ item.unidad_medida }}</span>
                </div>
                <div v-else class="flex flex-col items-center justify-center min-w-[3.2rem] h-12 bg-indigo-50 border border-indigo-100 text-indigo-700 rounded-xl">
                   <span class="text-[10px] font-black">{{ item.cantidad }}</span>
                   <span class="text-[7px] font-bold uppercase">{{ item.unidad_medida || 'un.' }}</span>
                </div>

                <div class="flex-1 min-w-0">
                  <h5 class="text-xs font-black text-slate-900 leading-tight truncate">{{ getProductName(item.product_id) }}</h5>
                  <p class="text-[10px] font-bold text-slate-400 mt-1">S/. {{ item.precio_unitario.toFixed(2) }} c/u</p>
                </div>

                <div class="text-right">
                  <p class="text-sm font-black text-slate-900">S/. {{ ((item.cantidad * item.precio_unitario) - item.descuento).toFixed(2) }}</p>
                  <button @click="removeItem(idx)" class="text-slate-300 hover:text-red-500 mt-1"><i class="pi pi-trash text-[10px]"></i></button>
                </div>
              </div>

              <!-- Controls -->
              <div class="flex items-center justify-between pt-2 border-t border-slate-50">
                <div v-if="!item.is_dimensional" class="flex items-center gap-1 bg-slate-100 rounded-lg p-1">
                  <button @click="item.cantidad = Math.max(1, item.cantidad - 1)" class="w-5 h-5 text-slate-500"><i class="pi pi-minus text-[8px]"></i></button>
                  <span class="text-[10px] font-black text-slate-900 min-w-[1.2rem] text-center">{{ item.cantidad }}</span>
                  <button @click="item.cantidad++" class="w-5 h-5 text-slate-500"><i class="pi pi-plus text-[8px]"></i></button>
                </div>
                <div v-else class="flex-1 px-2 py-1 bg-indigo-50 border border-indigo-100 rounded-lg text-[8px] font-mono font-bold text-indigo-600 text-center">
                   {{ getDimensionalFormula(item) }}
                </div>
                <div class="flex items-center gap-2 ml-2">
                  <span class="text-[8px] font-black text-slate-400 uppercase">Desc. S/.</span>
                  <input v-model.number="item.descuento" type="number" class="w-14 px-1 py-0.5 text-[10px] font-black bg-slate-100 border-none rounded text-right outline-none" />
                </div>
              </div>

              <!-- Dimensional Tool -->
              <div v-if="item.is_dimensional" class="grid grid-cols-12 gap-1.5 pt-1">
                <div v-if="['m', 'm2', 'm3'].includes(item.unidad_medida)" class="col-span-3">
                  <input v-model.number="item.alto" type="number" placeholder="H" class="w-full px-1 py-1 text-[10px] font-black bg-slate-50 border border-slate-200 rounded text-center"/>
                </div>
                <div v-if="['m2', 'm3'].includes(item.unidad_medida)" class="col-span-1 flex items-center justify-center">
                  <button @click="swapDimensions(item)" class="text-slate-300"><i class="pi pi-sync text-[8px]"></i></button>
                </div>
                <div v-if="['m2', 'm3'].includes(item.unidad_medida)" class="col-span-3">
                  <input v-model.number="item.ancho" type="number" placeholder="W" class="w-full px-1 py-1 text-[10px] font-black bg-slate-50 border border-slate-200 rounded text-center"/>
                </div>
                <div v-if="['m3'].includes(item.unidad_medida)" class="col-span-2">
                  <input v-model.number="item.espesor" type="number" placeholder="E" class="w-full px-1 py-1 text-[10px] font-black bg-slate-50 border border-slate-200 rounded text-center"/>
                </div>
                <div class="col-span-3">
                  <input v-model.number="item.cantidad_piezas" type="number" placeholder="N" class="w-full px-1 py-1 text-[10px] font-black bg-slate-50 border border-slate-200 rounded text-center"/>
                </div>
              </div>
            </div>

            <!-- Empty -->
            <div v-if="saleForm.items.length === 0" class="flex flex-col items-center justify-center py-12 text-slate-300">
               <i class="pi pi-shopping-bag text-3xl mb-2"></i>
               <p class="text-[10px] font-black uppercase">Vacío</p>
            </div>
          </div>

          <!-- Checkout Footer -->
          <div class="p-6 bg-slate-900 text-white space-y-4">
            <div class="space-y-1.5 border-b border-slate-800 pb-4">
              <div class="flex justify-between items-center text-slate-500 text-[10px] font-bold">
                <span>SUBTOTAL</span>
                <span>S/. {{ calculatedSubtotal.toFixed(2) }}</span>
              </div>
              <div class="flex justify-between items-center text-slate-500 text-[10px] font-bold">
                <span>IGV (18%)</span>
                <span>S/. {{ calculatedIGV.toFixed(2) }}</span>
              </div>
              <div class="flex justify-between items-center pt-1">
                <span class="text-xs font-black text-indigo-400">TOTAL A PAGAR</span>
                <span class="text-2xl font-black">S/. {{ calculatedTotal.toFixed(2) }}</span>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
               <select v-model="saleForm.tipo_documento" class="bg-slate-800 border-none rounded-xl px-3 py-2 text-[10px] font-black">
                 <option value="03">BOLETA</option>
                 <option value="01">FACTURA</option>
                 <option value="NV">NOTA VENTA</option>
               </select>
               <select v-model="saleForm.metodo_pago" class="bg-slate-800 border-none rounded-xl px-3 py-2 text-[10px] font-black">
                 <option value="EFECTIVO">EFECTIVO</option>
                 <option value="TARJETA">TARJETA</option>
                 <option value="YAPE">YAPE/PLIN</option>
               </select>
            </div>

            <div class="flex gap-2">
              <button @click="clearForm" class="flex-1 py-3 bg-slate-800 text-slate-500 rounded-xl text-[10px] font-black uppercase tracking-widest">Limpiar</button>
              <button @click="submitSale" :disabled="saleForm.items.length === 0" class="flex-[2] py-3 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-800 disabled:text-slate-600 text-white rounded-xl text-[10px] font-black uppercase tracking-widest transition-all">Pagar Ahora</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Mode 2: Sales History -->
    <div v-else class="flex-1 overflow-y-auto space-y-4 pb-20">
      <div v-for="sale in sales" :key="sale.id" class="bg-white p-5 rounded-3xl border border-slate-200 shadow-sm flex items-center gap-4 hover:border-indigo-300 transition-all">
        <div class="w-12 h-12 bg-indigo-50 text-indigo-600 rounded-2xl flex items-center justify-center"><i class="pi pi-receipt"></i></div>
        <div class="flex-1">
          <div class="flex justify-between items-start">
             <h4 class="text-sm font-black text-slate-900">#{{ sale.id.substring(0, 8).toUpperCase() }}</h4>
             <span class="text-sm font-black text-slate-950">S/. {{ sale.total.toFixed(2) }}</span>
          </div>
          <p class="text-xs text-slate-500 font-bold">{{ sale.customer?.nombre || 'Público General' }}</p>
          <p class="text-[10px] text-slate-400 mt-1 uppercase">{{ formatDate(sale.created_at) }}</p>
        </div>
        <button @click="viewDetails(sale.id)" class="w-8 h-8 flex items-center justify-center bg-slate-50 text-slate-400 rounded-xl"><i class="pi pi-chevron-right text-xs"></i></button>
      </div>
    </div>

    <!-- Mobile Floating Cart -->
    <button 
      v-if="activeSubTab === 'pos' && !isCartOpen" 
      @click="isCartOpen = true"
      class="lg:hidden fixed bottom-24 right-6 w-16 h-16 bg-indigo-600 text-white rounded-full shadow-2xl z-40 flex items-center justify-center"
    >
      <i class="pi pi-shopping-cart text-xl"></i>
      <span v-if="saleForm.items.length > 0" class="absolute -top-1 -right-1 w-6 h-6 bg-red-500 text-white rounded-full flex items-center justify-center text-[10px] font-black border-2 border-white">{{ saleForm.items.length }}</span>
    </button>

    <!-- Modals (Details & Annulment) -->
    <div v-if="showModal" class="fixed inset-0 bg-slate-900/70 backdrop-blur-md flex items-end sm:items-center justify-center p-0 sm:p-4 z-[70] animate-fade-in">
      <div @click="showModal = false" class="absolute inset-0"></div>
      <div class="relative bg-white rounded-t-[3rem] sm:rounded-[3rem] w-full max-w-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <div class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
              <span>Venta #{{ selectedSale?.sale?.id.substring(0, 8).toUpperCase() }}</span>
              <span v-if="selectedSale?.sale?.estado === 'annulled'" class="px-2 py-0.5 rounded text-[8px] bg-red-100 text-red-600 font-black uppercase tracking-wider">ANULADO</span>
            </h3>
            <p class="text-[10px] text-indigo-500 font-black uppercase tracking-widest mt-1">Resumen detallado</p>
          </div>
          <button @click="showModal = false" class="w-10 h-10 flex items-center justify-center bg-white border border-slate-200 text-slate-400 rounded-2xl"><i class="pi pi-times"></i></button>
        </div>

        <div class="flex-1 overflow-y-auto p-8 space-y-8">
          <div class="grid grid-cols-2 gap-8">
            <div class="space-y-1">
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Cliente</p>
              <p class="text-sm font-black text-slate-900">{{ selectedSale?.sale?.customer?.nombre || 'Público General' }}</p>
            </div>
            <div class="text-right space-y-1">
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Emisión</p>
              <p class="text-sm font-black text-slate-900">{{ formatDate(selectedSale?.sale?.created_at) }}</p>
            </div>
          </div>

          <div v-if="selectedSale?.has_electronic" class="p-6 bg-slate-900 rounded-[2rem] text-white">
            <div class="flex items-center justify-between mb-4">
              <div>
                <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Comprobante</p>
                <h4 class="text-lg font-black">{{ selectedSale.electronic_document.serie }}-{{ selectedSale.electronic_document.numero }}</h4>
              </div>
              <span class="px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest bg-emerald-500/20 text-emerald-400">{{ selectedSale.electronic_document.estado }}</span>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <button @click="downloadFile(selectedSale.electronic_document.document_uuid, 'pdf')" class="flex items-center justify-center gap-2 py-3 bg-white/10 rounded-xl text-[10px] font-black uppercase tracking-widest">PDF</button>
              <button @click="downloadFile(selectedSale.electronic_document.document_uuid, 'xml')" class="flex items-center justify-center gap-2 py-3 bg-white/10 rounded-xl text-[10px] font-black uppercase tracking-widest">XML</button>
            </div>
          </div>

          <div class="space-y-4">
            <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Ítems</p>
            <div v-for="item in selectedSale?.items" :key="item.id" class="flex items-center gap-4 py-3 border-b border-slate-50">
              <div class="w-10 h-10 bg-slate-50 rounded-xl flex items-center justify-center text-slate-400 font-black text-xs">{{ item.cantidad.toFixed(2) }}</div>
              <div class="flex-1">
                <h5 class="text-xs font-black text-slate-900">{{ item.product_nombre }}</h5>
                <p v-if="item.is_dimensional" class="text-[8px] text-slate-500 mt-0.5">Dim: {{ item.alto }}x{{ item.ancho }}x{{ item.espesor }} ({{ item.cantidad_piezas }})</p>
              </div>
              <p class="text-xs font-black text-slate-950">S/. {{ item.total.toFixed(2) }}</p>
            </div>
          </div>
        </div>

        <div class="p-8 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
           <div>
              <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Total</p>
              <p class="text-2xl font-black text-slate-900">S/. {{ selectedSale?.sale?.total.toFixed(2) }}</p>
           </div>
           <div class="flex gap-2">
             <button v-if="selectedSale?.sale?.estado !== 'annulled'" @click="openAnnulmentConfirm" class="px-5 py-3 bg-red-600 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest">Anular</button>
             <button @click="showModal = false" class="px-8 py-3 bg-slate-900 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest">Cerrar</button>
           </div>
        </div>
      </div>
    </div>

    <div v-if="showAnnulmentModal" class="fixed inset-0 bg-slate-900/70 backdrop-blur-md flex items-center justify-center p-4 z-[80] animate-fade-in">
      <div class="bg-white rounded-[2rem] max-w-md w-full p-6 shadow-2xl space-y-4 border border-slate-100">
        <h3 class="text-base font-black text-slate-900">Anular Comprobante</h3>
        <select v-model="annulmentReason" class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs font-bold">
          <option value="">-- Seleccionar Motivo --</option>
          <option value="ANULACION DE LA OPERACION">Anulación de la operación</option>
          <option value="ERROR EN EL RUC O NOMBRE">Error en el RUC o Nombre</option>
          <option value="DEVOLUCION TOTAL DE MERCADERIA">Devolución total</option>
        </select>
        <div class="flex justify-end gap-2 pt-2">
          <button @click="showAnnulmentModal = false" class="px-4 py-2 text-xs font-bold text-slate-500">Cancelar</button>
          <button @click="submitAnnulment" :disabled="!annulmentReason || annulling" class="px-4 py-2 text-xs font-bold text-white bg-red-600 rounded-xl">Confirmar</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, computed, watch } from 'vue'
import axios from 'axios'

const emit = defineEmits(['navigate'])

const activeSubTab = ref('pos')
const customers = ref<any[]>([])
const stocks = ref<any[]>([])
const sales = ref<any[]>([])
const productSearch = ref('')
const isCartOpen = ref(false)
const hasActiveCashSession = ref(true)
const activeSession = ref<any>(null)

// CSS Classes
const tabClass = 'flex items-center gap-2 px-6 py-3 rounded-2xl text-[10px] md:text-xs font-black uppercase tracking-widest transition-all'
const tabActive = 'bg-slate-900 text-white shadow-xl shadow-slate-200'
const tabInactive = 'text-slate-400 hover:bg-slate-50'

const saleForm = reactive({
  customer_id: '',
  tipo_documento: '03',
  metodo_pago: 'EFECTIVO',
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
  return saleForm.items.reduce((sum, item) => sum + (item.cantidad * item.precio_unitario) - (item.descuento || 0), 0)
})

const calculatedSubtotal = computed(() => {
  if (saleForm.tipo_documento === '01' || saleForm.tipo_documento === '03') {
    return calculatedTotal.value / 1.18
  }
  return calculatedTotal.value
})

const calculatedIGV = computed(() => {
  if (saleForm.tipo_documento === '01' || saleForm.tipo_documento === '03') {
    return calculatedTotal.value - calculatedSubtotal.value
  }
  return 0
})

watch(() => saleForm.customer_id, (newId) => {
  if (!newId) {
    saleForm.tipo_documento = '03'
    return
  }
  const cust = customers.value.find(c => c.id === newId)
  if (cust && (cust.tipo_documento === 'RUC' || (cust.numero_documento && cust.numero_documento.length === 11))) {
    saleForm.tipo_documento = '01'
  } else {
    saleForm.tipo_documento = '03'
  }
})

watch(() => saleForm.items, (newItems) => {
  newItems.forEach(item => {
    if (item.is_dimensional) {
      const piezas = item.cantidad_piezas || 1
      switch (item.unidad_medida) {
        case 'm':
          item.cantidad = (item.alto || 0) * piezas
          break
        case 'm2':
          item.cantidad = (item.alto || 0) * (item.ancho || 0) * piezas
          break
        case 'm3':
          item.cantidad = (item.alto || 0) * (item.ancho || 0) * (item.espesor || 0) * piezas
          break
        default:
          item.cantidad = (item.alto || 0) * (item.ancho || 0) * piezas
      }
    }
  })
}, { deep: true })



let barcodeBuffer = ''
let lastKeyTime = 0

const handleGlobalKeyPress = (e: KeyboardEvent) => {
  const target = e.target as HTMLElement
  // Ignore inputs unless it's body or generic search inputs
  if (target.tagName === 'INPUT' || target.tagName === 'SELECT' || target.tagName === 'TEXTAREA') {
    if (target.getAttribute('placeholder') === 'Buscar producto...') {
      // Allow searching normally, do not intercept
      return
    }
  }

  const currentTime = Date.now()
  if (currentTime - lastKeyTime > 50) {
    barcodeBuffer = ''
  }

  lastKeyTime = currentTime

  if (e.key !== 'Enter') {
    if (e.key.length === 1 && /[0-9a-zA-Z]/.test(e.key)) {
      barcodeBuffer += e.key
    }
  } else {
    if (barcodeBuffer.length >= 4) {
      e.preventDefault()
      
      // Look for the product with this barcode or code in stocks
      const foundStock = stocks.value.find(s => 
        (s.product_codigo && s.product_codigo === barcodeBuffer) || 
        (s.product_codigo_barras && s.product_codigo_barras === barcodeBuffer)
      )

      if (foundStock) {
        quickAddItem(foundStock)
      } else {
        // Fallback: type the code into search input
        productSearch.value = barcodeBuffer
      }
      
      barcodeBuffer = ''
    }
  }
}

onMounted(() => {
  window.addEventListener('keypress', handleGlobalKeyPress)
  checkCashSession()
  loadCustomers()
  loadStocks()
})

onUnmounted(() => {
  window.removeEventListener('keypress', handleGlobalKeyPress)
})

async function checkCashSession() {
  try {
    const res = await axios.get('/cash/active')
    if (res.data.success) {
      activeSession.value = res.data.data
      hasActiveCashSession.value = !!res.data.data
    }
  } catch (err) {
    console.error('Error checking cash session', err)
    hasActiveCashSession.value = false
  }
}

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

/**
 * Senior UX: Intercambia las dimensiones de Alto y Ancho para ítems m2/m3
 */
function swapDimensions(item: any) {
  const temp = item.alto
  item.alto = item.ancho
  item.ancho = temp
}

/**
 * Senior Feedback: Genera un string de la fórmula matemática para visualización
 */
function getDimensionalFormula(item: any): string {
  const piezasStr = item.cantidad_piezas > 1 ? ` × ${item.cantidad_piezas}` : ''
  const unit = item.unidad_medida || 'm'
  
  if (unit === 'm') {
    return `${item.alto.toFixed(2)}m${piezasStr} = ${item.cantidad.toFixed(2)}m`
  }
  if (unit === 'm2') {
    return `[${item.alto.toFixed(2)} × ${item.ancho.toFixed(2)}]${piezasStr} = ${item.cantidad.toFixed(2)}m²`
  }
  if (unit === 'm3') {
    return `[${item.alto.toFixed(2)} × ${item.ancho.toFixed(2)} × ${item.espesor.toFixed(2)}]${piezasStr} = ${item.cantidad.toFixed(2)}m³`
  }
  return `${item.cantidad.toFixed(2)} ${unit}`
}

function quickAddItem(stock: any) {
  if (stock.is_dimensional) {
    const defaultAlto = ['m', 'm2', 'm3'].includes(stock.unidad_medida) ? 1.0 : 0.0
    const defaultAncho = ['m2', 'm3'].includes(stock.unidad_medida) ? 1.0 : 0.0
    const defaultEspesor = ['m3'].includes(stock.unidad_medida) ? 0.1 : 0.0

    saleForm.items.push({
      product_id: stock.product_id,
      cantidad: defaultAlto || 1.0,
      precio_unitario: stock.precio_venta,
      descuento: 0,
      alto: defaultAlto,
      ancho: defaultAncho,
      espesor: defaultEspesor,
      cantidad_piezas: 1,
      is_dimensional: true,
      unidad_medida: stock.unidad_medida
    })
    isCartOpen.value = true
    return
  }

  const alreadyAdded = saleForm.items.find(item => item.product_id === stock.product_id && !item.is_dimensional)
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
      descuento: 0,
      is_dimensional: false,
      unidad_medida: stock.unidad_medida || 'und'
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
  saleForm.tipo_documento = '03'
  saleForm.metodo_pago = 'EFECTIVO'
  saleForm.items = []
  isCartOpen.value = false
}

async function submitSale() {
  if (saleForm.items.length === 0) return

  // Prepare payload: if customer_id is empty, send null
  const payload = {
    ...saleForm,
    customer_id: saleForm.customer_id === '' ? null : saleForm.customer_id
  }

  try {
    const res = await axios.post('/sales', payload)
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

async function downloadFile(uuid: string, type: 'pdf' | 'xml' | 'cdr') {
  try {
    // This is the endpoint to get secure URLs from the billing microservice
    // For now we can redirect directly if we know the pattern or fetch it
    const res = await axios.get(`/billing/files/${uuid}`) // We might need to implement this handler in backend
    if (res.data.files && res.data.files[type]) {
      window.open(res.data.files[type], '_blank')
    }
  } catch (err) {
    // Fallback or simple alert
    alert('Obteniendo archivo...')
    // In production, this would call the FacturaAPI redirect
  }
}

// Annulment variables & handlers
const showAnnulmentModal = ref(false)
const annulmentReason = ref('')
const annulling = ref(false)

function openAnnulmentConfirm() {
  annulmentReason.value = ''
  showAnnulmentModal.value = true
}

async function submitAnnulment() {
  if (!annulmentReason.value.trim()) return
  annulling.value = true
  try {
    const res = await axios.post(`/sales/${selectedSale.value.sale.id}/credit-note`, {
      motivo: annulmentReason.value
    })
    if (res.data.success) {
      alert('¡Venta anulada con éxito mediante Nota de Crédito!')
      showAnnulmentModal.value = false
      showModal.value = false
      loadSales() // Reload history
    }
  } catch (err: any) {
    alert('Error al anular la venta: ' + (err.response?.data?.error || err.message))
  } finally {
    annulling.value = false
  }
}

function getDocTypeName(tipo: string) {
  if (tipo === '01') return 'Factura'
  if (tipo === '03') return 'Boleta'
  if (tipo === 'NV') return 'Nota de Venta'
  return 'Ticket'
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
