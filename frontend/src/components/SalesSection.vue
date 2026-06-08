<template>
  <div class="space-y-6">
    <!-- Tab navigation -->
    <div class="flex gap-4 border-b border-slate-200 pb-3">
      <button
        @click="activeSubTab = 'pos'; loadStocks()"
        :class="['px-4 py-2 text-xs font-bold rounded-xl transition-all',
          activeSubTab === 'pos' ? 'bg-sky-600 text-white shadow-sm' : 'text-slate-500 hover:bg-slate-100']"
      >
        <i class="pi pi-desktop mr-2"></i>Punto de Venta (POS)
      </button>
      <button
        @click="activeSubTab = 'history'; loadSales()"
        :class="['px-4 py-2 text-xs font-bold rounded-xl transition-all',
          activeSubTab === 'history' ? 'bg-sky-600 text-white shadow-sm' : 'text-slate-500 hover:bg-slate-100']"
      >
        <i class="pi pi-list mr-2"></i>Historial de Ventas
      </button>
    </div>

    <!-- Mode 1: POS Screen -->
    <div v-if="activeSubTab === 'pos'" class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left side: Customer & Cart Add -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-6 h-fit">
        <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 flex items-center gap-2">
          <i class="pi pi-shopping-bag text-sky-500"></i>
          <span>Detalles de Transacción</span>
        </h3>

        <!-- Customer Select -->
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Cliente</label>
          <select
            v-model="saleForm.customer_id"
            required
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500 focus:outline-none transition-all"
          >
            <option value="">-- Seleccionar Cliente --</option>
            <option v-for="cust in customers" :key="cust.id" :value="cust.id">
              {{ cust.nombre }} ({{ cust.tipo_documento }}: {{ cust.numero_documento }})
            </option>
          </select>
          <div v-if="customers.length === 0" class="text-[10px] text-red-500 mt-1">
            * Registra clientes primero en la sección de Clientes.
          </div>
        </div>

        <div class="border-t border-slate-100 pt-4 space-y-4">
          <h4 class="text-xs font-bold text-slate-700 uppercase tracking-wider">Agregar Producto al Carrito</h4>
          
          <!-- Product (Stock) Select -->
          <div>
            <label class="block text-xs font-semibold text-slate-500 mb-1">Producto</label>
            <select
              v-model="itemInput.product_id"
              @change="onProductSelect"
              class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500 focus:outline-none transition-all"
            >
              <option value="">-- Seleccionar Producto --</option>
              <option v-for="item in stocks" :key="item.product_id" :value="item.product_id" :disabled="item.stock_actual <= 0">
                {{ item.product_nombre }} [Stock: {{ item.stock_actual }}] - S/. {{ item.precio_venta.toFixed(2) }}
              </option>
            </select>
          </div>

          <!-- Price (display/override) & Quantity -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold text-slate-500 mb-1">Precio Unitario</label>
              <div class="relative">
                <span class="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-400 text-xs">S/.</span>
                <input
                  v-model.number="itemInput.precio_unitario"
                  type="number"
                  step="0.01"
                  min="0.01"
                  class="block w-full pl-8 pr-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500"
                />
              </div>
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-500 mb-1">Cantidad</label>
              <input
                v-model.number="itemInput.cantidad"
                type="number"
                step="0.01"
                min="0.01"
                class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500"
              />
            </div>
          </div>

          <!-- Discount -->
          <div>
            <label class="block text-xs font-semibold text-slate-500 mb-1">Descuento Total en Item</label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-3 flex items-center text-slate-400 text-xs">S/.</span>
              <input
                v-model.number="itemInput.descuento"
                type="number"
                step="0.01"
                min="0"
                class="block w-full pl-8 pr-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500"
                placeholder="0.00"
              />
            </div>
          </div>

          <button
            type="button"
            @click="addItem"
            class="w-full py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-xs font-bold transition-all flex items-center justify-center gap-2"
          >
            <i class="pi pi-plus-circle"></i>
            <span>Agregar al Carrito</span>
          </button>
        </div>
      </div>

      <!-- Right side: Cart list and payment actions -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm lg:col-span-2 flex flex-col justify-between">
        <div>
          <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4">
            Carrito de Ventas
          </h3>

          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
                  <th class="py-3 px-4">Producto</th>
                  <th class="py-3 px-4 text-center">Cantidad</th>
                  <th class="py-3 px-4 text-right">P. Unit.</th>
                  <th class="py-3 px-4 text-right">Desc.</th>
                  <th class="py-3 px-4 text-right">Total</th>
                  <th class="py-3 px-4 text-center">Eliminar</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
                <tr v-for="(item, idx) in saleForm.items" :key="idx" class="hover:bg-slate-50/50">
                  <td class="py-3 px-4 font-semibold text-slate-900">{{ getProductName(item.product_id) }}</td>
                  <td class="py-3 px-4 text-center font-mono">{{ item.cantidad }}</td>
                  <td class="py-3 px-4 text-right font-mono">S/. {{ item.precio_unitario.toFixed(2) }}</td>
                  <td class="py-3 px-4 text-right font-mono text-red-500">-S/. {{ item.descuento.toFixed(2) }}</td>
                  <td class="py-3 px-4 text-right font-semibold font-mono text-slate-900">
                    S/. {{ ((item.cantidad * item.precio_unitario) - item.descuento).toFixed(2) }}
                  </td>
                  <td class="py-3 px-4 text-center">
                    <button
                      @click="removeItem(idx)"
                      class="text-red-500 hover:text-red-700 p-1 hover:bg-red-50 rounded-lg transition-all"
                    >
                      <i class="pi pi-trash"></i>
                    </button>
                  </td>
                </tr>
                <tr v-if="saleForm.items.length === 0">
                  <td colspan="6" class="text-center py-12 text-slate-400">
                    <i class="pi pi-inbox text-3xl block mb-2 text-slate-300"></i>
                    El carrito está vacío. Agrega productos.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="border-t border-slate-100 pt-6 mt-6 space-y-4">
          <div class="flex justify-end gap-12 text-xs text-slate-600">
            <div class="space-y-1.5 text-right">
              <div>Subtotal afecto (sin IGV):</div>
              <div>IGV (18%):</div>
              <div class="text-sm font-black text-slate-950">Total a Pagar:</div>
            </div>
            <div class="space-y-1.5 text-right font-mono font-bold text-slate-900">
              <div>S/. {{ calculatedSubtotal.toFixed(2) }}</div>
              <div>S/. {{ calculatedIGV.toFixed(2) }}</div>
              <div class="text-sm font-black text-slate-950">S/. {{ calculatedTotal.toFixed(2) }}</div>
            </div>
          </div>

          <div class="flex justify-end gap-3">
            <button
              @click="clearForm"
              class="px-5 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl text-xs font-bold transition-all"
            >
              Cancelar Venta
            </button>
            <button
              @click="submitSale"
              :disabled="saleForm.items.length === 0 || !saleForm.customer_id"
              class="px-6 py-2.5 bg-sky-600 hover:bg-sky-500 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center gap-2"
            >
              <i class="pi pi-check-circle"></i>
              <span>Procesar Pago (Completar)</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Mode 2: Sales History -->
    <div v-else class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4">
        Lista de Ventas Realizadas
      </h3>

      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Código / Fecha</th>
              <th class="py-3 px-4">Cliente</th>
              <th class="py-3 px-4 text-right">Subtotal</th>
              <th class="py-3 px-4 text-right">IGV</th>
              <th class="py-3 px-4 text-right">Total</th>
              <th class="py-3 px-4 text-center">Estado</th>
              <th class="py-3 px-4 text-center">Detalle</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="sale in sales" :key="sale.id" class="hover:bg-slate-50/50">
              <td class="py-3.5 px-4">
                <div class="font-mono font-bold text-slate-900">VENTA-{{ sale.id.substring(0, 8) }}</div>
                <div class="text-[10px] text-slate-400 mt-0.5">{{ formatDate(sale.created_at) }}</div>
              </td>
              <td class="py-3.5 px-4 font-semibold text-slate-900">{{ sale.customer?.nombre || '-' }}</td>
              <td class="py-3.5 px-4 text-right font-mono text-slate-500">S/. {{ sale.subtotal.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-right font-mono text-slate-500">S/. {{ sale.igv.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-right font-mono font-bold text-slate-900">S/. {{ sale.total.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-50 text-emerald-700 border border-emerald-200">
                  {{ sale.estado }}
                </span>
              </td>
              <td class="py-3.5 px-4 text-center">
                <button
                  @click="viewDetails(sale.id)"
                  class="text-sky-600 hover:text-sky-800 p-1 hover:bg-sky-50 rounded-lg transition-all"
                >
                  <i class="pi pi-eye"></i>
                </button>
              </td>
            </tr>
            <tr v-if="sales.length === 0">
              <td colspan="7" class="text-center py-8 text-slate-400">No hay ventas registradas</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Details Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 z-50">
      <div class="bg-white rounded-2xl max-w-2xl w-full p-6 shadow-2xl border border-slate-100 flex flex-col max-h-[85vh]">
        <div class="flex justify-between items-center border-b border-slate-100 pb-3.5 mb-4">
          <h3 class="text-base font-extrabold text-slate-900">
            Detalles de Venta: <span class="font-mono text-sky-600">VENTA-{{ selectedSale?.sale?.id.substring(0, 8) }}</span>
          </h3>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-600">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <div class="grid grid-cols-2 gap-4 text-xs text-slate-600 mb-6 bg-slate-50 p-4 rounded-xl">
          <div>
            <div><strong>Cliente:</strong> {{ selectedSale?.sale?.customer?.nombre }}</div>
            <div><strong>Doc:</strong> {{ selectedSale?.sale?.customer?.tipo_documento }} {{ selectedSale?.sale?.customer?.numero_documento }}</div>
          </div>
          <div class="text-right">
            <div><strong>Fecha:</strong> {{ formatDate(selectedSale?.sale?.created_at) }}</div>
            <div><strong>Total General:</strong> S/. {{ selectedSale?.sale?.total.toFixed(2) }}</div>
          </div>
        </div>

        <div class="overflow-y-auto flex-1 mb-4">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
                <th class="py-2 px-3">Código</th>
                <th class="py-2 px-3">Producto</th>
                <th class="py-2 px-3 text-center">Cantidad</th>
                <th class="py-2 px-3 text-right">P. Unit.</th>
                <th class="py-2 px-3 text-right">Desc.</th>
                <th class="py-2 px-3 text-right">Total</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
              <tr v-for="item in selectedSale?.items" :key="item.id">
                <td class="py-2.5 px-3 font-mono text-[10px] text-slate-400">{{ item.product_codigo || '-' }}</td>
                <td class="py-2.5 px-3 font-semibold text-slate-950">{{ item.product_nombre }}</td>
                <td class="py-2.5 px-3 text-center font-mono">{{ item.cantidad }}</td>
                <td class="py-2.5 px-3 text-right font-mono">S/. {{ item.precio_unitario.toFixed(2) }}</td>
                <td class="py-2.5 px-3 text-right font-mono text-red-500">-S/. {{ item.descuento.toFixed(2) }}</td>
                <td class="py-2.5 px-3 text-right font-mono font-bold">S/. {{ item.total.toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex justify-end pt-3 border-t border-slate-100">
          <button
            @click="showModal = false"
            class="px-5 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-xs font-bold transition-all"
          >
            Cerrar Detalle
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

const saleForm = reactive({
  customer_id: '',
  items: [] as any[]
})

const itemInput = reactive({
  product_id: '',
  cantidad: 1,
  precio_unitario: 0.00,
  descuento: 0
})

const showModal = ref(false)
const selectedSale = ref<any>(null)

const calculatedTotal = computed(() => {
  return saleForm.items.reduce((sum, item) => sum + (item.cantidad * item.precio_unitario) - item.descuento, 0)
})

const calculatedSubtotal = computed(() => {
  return calculatedTotal.value / 1.18
})

const calculatedIGV = computed(() => {
  return calculatedTotal.value - calculatedSubtotal.value
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
    if (res.data.success) stocks.value = res.data.data
  } catch (err) {
    console.error('Error loading stocks', err)
  }
}

async function loadSales() {
  try {
    const res = await axios.get('/sales')
    if (res.data.success) sales.value = res.data.data
  } catch (err) {
    console.error('Error loading sales', err)
  }
}

function onProductSelect() {
  const stockObj = stocks.value.find(s => s.product_id === itemInput.product_id)
  if (stockObj) {
    itemInput.precio_unitario = stockObj.precio_venta
    itemInput.cantidad = 1
    itemInput.descuento = 0
  }
}

function getProductName(pId: string) {
  const stockObj = stocks.value.find(s => s.product_id === pId)
  return stockObj ? stockObj.product_nombre : 'Producto desconocido'
}

function addItem() {
  if (!itemInput.product_id || itemInput.cantidad <= 0 || itemInput.precio_unitario <= 0) {
    alert('Ingrese un producto, cantidad y precio válidos.')
    return
  }

  const stockObj = stocks.value.find(s => s.product_id === itemInput.product_id)
  if (!stockObj) return

  // Check frontend stock
  const alreadyAdded = saleForm.items.find(item => item.product_id === itemInput.product_id)
  const currentAddedQty = alreadyAdded ? alreadyAdded.cantidad : 0
  const neededQty = currentAddedQty + itemInput.cantidad

  if (neededQty > stockObj.stock_actual) {
    alert(`No hay suficiente stock. Disponible: ${stockObj.stock_actual}`)
    return
  }

  if (alreadyAdded) {
    alreadyAdded.cantidad += itemInput.cantidad
    alreadyAdded.descuento += itemInput.descuento
    alreadyAdded.precio_unitario = itemInput.precio_unitario // Update to latest overridden price
  } else {
    saleForm.items.push({
      product_id: itemInput.product_id,
      cantidad: itemInput.cantidad,
      precio_unitario: itemInput.precio_unitario,
      descuento: itemInput.descuento
    })
  }

  // Reset item inputs
  itemInput.product_id = ''
  itemInput.cantidad = 1
  itemInput.precio_unitario = 0
  itemInput.descuento = 0
}

function removeItem(idx: number) {
  saleForm.items.splice(idx, 1)
}

function clearForm() {
  saleForm.customer_id = ''
  saleForm.items = []
}

async function submitSale() {
  if (!saleForm.customer_id || saleForm.items.length === 0) return

  try {
    const res = await axios.post('/sales', saleForm)
    if (res.data.success) {
      alert('¡Venta completada con éxito!')
      clearForm()
      await loadStocks() // Reload stocks
      activeSubTab.value = 'history'
      loadSales()
    }
  } catch (err: any) {
    const errMsg = err.response?.data?.error || 'Error al procesar la venta'
    alert(errMsg)
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
    alert('Error al cargar detalles de la venta')
  }
}

function formatDate(dStr: string) {
  if (!dStr) return '-'
  const d = new Date(dStr)
  return d.toLocaleString('es-PE', { timeZone: 'America/Lima' })
}
</script>
