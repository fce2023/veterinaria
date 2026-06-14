<template>
  <div class="space-y-6">
    <!-- Tab navigation -->
    <div class="flex gap-4 border-b border-slate-200 pb-3">
      <button
        @click="activeSubTab = 'create'"
        :class="['px-4 py-2 text-xs font-bold rounded-xl transition-all',
          activeSubTab === 'create' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-500 hover:bg-slate-100']"
      >
        <i class="pi pi-plus mr-2"></i>Registrar Compra
      </button>
      <button
        @click="activeSubTab = 'history'; loadPurchases()"
        :class="['px-4 py-2 text-xs font-bold rounded-xl transition-all',
          activeSubTab === 'history' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-500 hover:bg-slate-100']"
      >
        <i class="pi pi-list mr-2"></i>Historial de Compras
      </button>
    </div>

    <!-- Mode 1: Create Purchase -->
    <div v-if="activeSubTab === 'create'" class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left side: Add items & select supplier -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-6 h-fit">
        <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 flex items-center gap-2">
          <i class="pi pi-shopping-cart text-emerald-500"></i>
          <span>Detalles de la Compra</span>
        </h3>

        <!-- Supplier Select -->
        <div>
          <div class="flex items-center justify-between mb-1.5">
            <label class="block text-xs font-bold text-slate-600">Proveedor</label>
            <button 
              type="button" 
              @click="showQuickSupplierModal = true" 
              class="text-[10px] font-black text-emerald-600 hover:text-emerald-800 uppercase tracking-widest flex items-center gap-1"
            >
              <i class="pi pi-plus text-[8px]"></i> Nuevo
            </button>
          </div>
          <select
            v-model="purchaseForm.supplier_id"
            required
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 focus:outline-none transition-all"
          >
            <option value="">-- Seleccionar Proveedor --</option>
            <option v-for="sup in suppliers" :key="sup.id" :value="sup.id">
              {{ sup.razon_social }} (RUC: {{ sup.ruc }})
            </option>
          </select>
        </div>

        <div class="border-t border-slate-100 pt-4 space-y-4">
          <h4 class="text-xs font-bold text-slate-700 uppercase tracking-wider">Agregar Ítem</h4>
          
          <!-- Product Autocomplete Search -->
          <div class="relative">
            <label class="block text-xs font-semibold text-slate-500 mb-1">Producto / SKU / Barras</label>
            <div class="relative">
              <input 
                ref="searchInputRef"
                v-model="productSearchQuery"
                @focus="showProductDropdown = true"
                type="text"
                placeholder="Escribe para buscar..."
                class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 focus:outline-none transition-all"
              />
              <button 
                v-if="productSearchQuery" 
                type="button" 
                @click="productSearchQuery = ''; itemInput.product_id = ''" 
                class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
              >
                <i class="pi pi-times text-[10px]"></i>
              </button>
            </div>
            
            <!-- Floating list -->
            <div 
              v-if="showProductDropdown && filteredProductsList.length > 0" 
              class="absolute z-50 left-0 right-0 mt-1 max-h-48 overflow-y-auto bg-white border border-slate-200 rounded-xl shadow-xl divide-y divide-slate-100"
            >
              <button
                v-for="prod in filteredProductsList"
                :key="prod.id"
                type="button"
                @click="selectProduct(prod)"
                class="w-full text-left px-3 py-2 text-xs hover:bg-slate-50 flex flex-col transition-colors"
              >
                <span class="font-bold text-slate-900">{{ prod.nombre }}</span>
                <span class="text-[10px] text-slate-400 mt-0.5">Código: {{ prod.codigo || '-' }} | Stock: S/. {{ prod.precio_venta }}</span>
              </button>
            </div>
          </div>

          <!-- Quantity and Unit Cost -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold text-slate-500 mb-1">Cantidad</label>
              <input
                v-model.number="itemInput.cantidad"
                type="number"
                step="0.01"
                min="0.01"
                class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
              />
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-500 mb-1">Costo Unitario (S/.)</label>
              <input
                v-model.number="itemInput.costo_unitario"
                type="number"
                step="0.01"
                min="0.01"
                class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
              />
            </div>
          </div>

          <button
            type="button"
            @click="addItem"
            class="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-xs font-bold transition-all flex items-center justify-center gap-2 shadow-sm"
          >
            <i class="pi pi-plus-circle"></i>
            <span>Agregar al Detalle</span>
          </button>
        </div>
      </div>

      <!-- Right side: Table of items and actions -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm lg:col-span-2 flex flex-col justify-between">
        <div>
          <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4">
            Detalle de la Compra
          </h3>

          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
                  <th class="py-3 px-4">Producto</th>
                  <th class="py-3 px-4 text-center">Cantidad</th>
                  <th class="py-3 px-4 text-right">Costo Unit.</th>
                  <th class="py-3 px-4 text-right">Total</th>
                  <th class="py-3 px-4 text-center">Acciones</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
                <tr v-for="(item, idx) in purchaseForm.items" :key="idx" class="hover:bg-slate-50/50">
                  <td class="py-3 px-4 font-semibold text-slate-900">{{ getProductName(item.product_id) }}</td>
                  <td class="py-2 px-2 text-center">
                    <input 
                      v-model.number="item.cantidad"
                      type="number"
                      step="0.01"
                      min="0.01"
                      class="w-16 px-2 py-1 border border-slate-200 rounded-lg text-center font-mono text-xs focus:ring-2 focus:ring-emerald-500/20 outline-none"
                    />
                  </td>
                  <td class="py-2 px-2 text-right">
                    <div class="flex items-center justify-end gap-1">
                      <span class="text-[10px] text-slate-400">S/.</span>
                      <input 
                        v-model.number="item.costo_unitario"
                        type="number"
                        step="0.01"
                        min="0.01"
                        class="w-20 px-2 py-1 border border-slate-200 rounded-lg text-right font-mono text-xs focus:ring-2 focus:ring-emerald-500/20 outline-none"
                      />
                    </div>
                  </td>
                  <td class="py-3 px-4 text-right font-semibold font-mono text-slate-900">
                    S/. {{ (item.cantidad * item.costo_unitario).toFixed(2) }}
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
                <tr v-if="purchaseForm.items.length === 0">
                  <td colspan="5" class="text-center py-12 text-slate-400">
                    <i class="pi pi-inbox text-3xl block mb-2 text-slate-300"></i>
                    Aún no has agregado productos al detalle.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="border-t border-slate-100 pt-6 mt-6 space-y-4">
          <div class="flex justify-end gap-12 text-xs text-slate-600">
            <div class="space-y-1.5 text-right">
              <div>Subtotal (sin IGV):</div>
              <div>IGV (18%):</div>
              <div class="text-sm font-black text-slate-950">Total General:</div>
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
              Vaciar Todo
            </button>
            <button
              @click="submitPurchase"
              :disabled="purchaseForm.items.length === 0 || !purchaseForm.supplier_id"
              class="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center gap-2"
            >
              <i class="pi pi-check-circle"></i>
              <span>Confirmar y Guardar Compra</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Mode 2: Purchases History -->
    <div v-else class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4">
        Lista de Compras Realizadas
      </h3>

      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Código / Fecha</th>
              <th class="py-3 px-4">Proveedor</th>
              <th class="py-3 px-4 text-right">Subtotal</th>
              <th class="py-3 px-4 text-right">IGV</th>
              <th class="py-3 px-4 text-right">Total</th>
              <th class="py-3 px-4 text-center">Estado</th>
              <th class="py-3 px-4 text-center">Detalle</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="p in purchases" :key="p.id" class="hover:bg-slate-50/50">
              <td class="py-3.5 px-4">
                <div class="font-mono font-bold text-slate-900">COMPRA-{{ p.id.substring(0, 8) }}</div>
                <div class="text-[10px] text-slate-400 mt-0.5">{{ formatDate(p.fecha) }}</div>
              </td>
              <td class="py-3.5 px-4 font-semibold text-slate-900">{{ p.supplier?.razon_social || '-' }}</td>
              <td class="py-3.5 px-4 text-right font-mono text-slate-500">S/. {{ p.subtotal.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-right font-mono text-slate-500">S/. {{ p.igv.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-right font-mono font-bold text-slate-900">S/. {{ p.total.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-50 text-emerald-700 border border-emerald-200">
                  {{ p.estado }}
                </span>
              </td>
              <td class="py-3.5 px-4 text-center">
                <button
                  @click="viewDetails(p.id)"
                  class="text-sky-600 hover:text-sky-800 p-1 hover:bg-sky-50 rounded-lg transition-all"
                >
                  <i class="pi pi-eye"></i>
                </button>
              </td>
            </tr>
            <tr v-if="purchases.length === 0">
              <td colspan="7" class="text-center py-8 text-slate-400">No hay compras registradas</td>
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
            Detalles de Compra: <span class="font-mono text-sky-600">COMPRA-{{ selectedPurchase?.purchase?.id.substring(0, 8) }}</span>
          </h3>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-600">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <div class="grid grid-cols-2 gap-4 text-xs text-slate-600 mb-6 bg-slate-50 p-4 rounded-xl">
          <div>
            <div><strong>Proveedor:</strong> {{ selectedPurchase?.purchase?.supplier?.razon_social }}</div>
            <div><strong>RUC:</strong> {{ selectedPurchase?.purchase?.supplier?.ruc }}</div>
          </div>
          <div class="text-right">
            <div><strong>Fecha:</strong> {{ formatDate(selectedPurchase?.purchase?.fecha) }}</div>
            <div><strong>Total General:</strong> S/. {{ selectedPurchase?.purchase?.total.toFixed(2) }}</div>
          </div>
        </div>

        <div class="overflow-y-auto flex-1 mb-4">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
                <th class="py-2 px-3">Código</th>
                <th class="py-2 px-3">Producto</th>
                <th class="py-2 px-3 text-center">Cantidad</th>
                <th class="py-2 px-3 text-right">Costo Unit.</th>
                <th class="py-2 px-3 text-right">Subtotal</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
              <tr v-for="item in selectedPurchase?.items" :key="item.id">
                <td class="py-2.5 px-3 font-mono text-[10px] text-slate-400">{{ item.product_codigo || '-' }}</td>
                <td class="py-2.5 px-3 font-semibold text-slate-950">{{ item.product_nombre }}</td>
                <td class="py-2.5 px-3 text-center font-mono">{{ item.cantidad }}</td>
                <td class="py-2.5 px-3 text-right font-mono">S/. {{ item.costo_unitario.toFixed(2) }}</td>
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

    <!-- Quick Add Supplier Modal -->
    <div v-if="showQuickSupplierModal" class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100">
        <form @submit.prevent="handleQuickSupplierSubmit" class="space-y-4">
          <div>
            <h3 class="text-base font-extrabold text-slate-900">Agregar Proveedor Rápido</h3>
            <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-1">Crea un proveedor sin cerrar el formulario de compras</p>
          </div>
          <div class="space-y-3">
            <div class="flex items-center gap-2 mb-1 px-1">
              <input 
                id="quickGenericSupplier"
                v-model="quickSupplierIsGeneric"
                type="checkbox"
                class="rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
              />
              <label for="quickGenericSupplier" class="text-[10px] font-bold text-slate-600 uppercase tracking-wider cursor-pointer">
                Sin RUC / Proveedor Genérico
              </label>
            </div>
            <div>
              <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest px-1 mb-1">RUC (11 dígitos)</label>
              <div class="relative">
                <input 
                  v-model="quickSupplierRUC"
                  type="text"
                  required
                  maxlength="11"
                  :disabled="quickSupplierIsGeneric"
                  class="block w-full pl-4 pr-10 py-3 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 outline-none transition-all disabled:opacity-60"
                  placeholder="Ej. 20123456789"
                />
                <div class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
                  <i v-if="quickSupplierLoading" class="pi pi-spin pi-spinner text-emerald-600"></i>
                  <i v-else class="pi pi-search text-xs"></i>
                </div>
              </div>
              <span v-if="!quickSupplierIsGeneric" class="text-[8px] text-slate-400 font-bold uppercase tracking-wider mt-1 block px-0.5">Se consultará SUNAT al completar 11 dígitos</span>
            </div>
            <div>
              <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest px-1 mb-1">Razón Social</label>
              <input 
                v-model="quickSupplierName"
                type="text"
                required
                class="block w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 outline-none transition-all"
                placeholder="Ej. Distribuidora Veterinaria S.A.C."
              />
            </div>
            <div>
              <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest px-1 mb-1">Dirección</label>
              <input 
                v-model="quickSupplierAddress"
                type="text"
                class="block w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 outline-none transition-all"
                placeholder="Ej. Av. Industrial 500, Lima"
              />
            </div>
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <button 
              type="button" 
              @click="showQuickSupplierModal = false" 
              class="px-4 py-2 text-xs font-bold text-slate-500 hover:bg-slate-100 rounded-xl transition-all"
            >
              Cancelar
            </button>
            <button 
              type="submit" 
              :disabled="quickSupplierSaving"
              class="px-4 py-2 text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-700 rounded-xl disabled:opacity-50 transition-all"
            >
              <span v-if="quickSupplierSaving">Guardando...</span>
              <span v-else>Guardar</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, computed, watch } from 'vue'
import axios from 'axios'

const activeSubTab = ref('create')
const suppliers = ref<any[]>([])
const products = ref<any[]>([])
const purchases = ref<any[]>([])
const stocks = ref<any[]>([])

// Zero friction search filters
const productSearchQuery = ref('')
const showProductDropdown = ref(false)

// Quick add Supplier variables
const showQuickSupplierModal = ref(false)
const quickSupplierName = ref('')
const quickSupplierRUC = ref('')
const quickSupplierAddress = ref('')
const quickSupplierSaving = ref(false)
const quickSupplierLoading = ref(false)
const quickSupplierIsGeneric = ref(false)

watch(() => quickSupplierRUC.value, async (newVal) => {
  if (quickSupplierIsGeneric.value) return
  const cleanRuc = newVal.trim()
  if (cleanRuc.length === 11) {
    quickSupplierLoading.value = true
    try {
      const res = await axios.get(`/public/ruc/${cleanRuc}`)
      if (res.data.success && res.data.data) {
        const d = res.data.data
        quickSupplierName.value = d.razon_social || d.nombre_o_razon_social || d.razonSocial || d.nombre || ''
        
        let address = d.direccion || d.direccion_completa || ''
        if (!address || address.trim() === '-') {
          const parts = []
          if (d.departamento && d.departamento !== '-') parts.push(d.departamento)
          if (d.provincia && d.provincia !== '-') parts.push(d.provincia)
          if (d.distrito && d.distrito !== '-') parts.push(d.distrito)
          address = parts.join(' - ')
        }
        quickSupplierAddress.value = address
      }
    } catch (err) {
      console.error('Error querying RUC', err)
    } finally {
      quickSupplierLoading.value = false
    }
  }
})

watch(() => quickSupplierIsGeneric.value, (newVal) => {
  if (newVal) {
    quickSupplierRUC.value = '00000000000'
    quickSupplierName.value = 'PROVEEDOR VARIOS / GENÉRICO'
    quickSupplierAddress.value = 'LIMA'
  } else {
    quickSupplierRUC.value = ''
    quickSupplierName.value = ''
    quickSupplierAddress.value = ''
  }
})

const purchaseForm = reactive({
  supplier_id: '',
  items: [] as any[]
})

const itemInput = reactive({
  product_id: '',
  cantidad: 1,
  costo_unitario: 1.00
})

const showModal = ref(false)
const selectedPurchase = ref<any>(null)

// Focus ref handlers
const searchInputRef = ref<HTMLInputElement | null>(null)

const filteredProductsList = computed(() => {
  const q = productSearchQuery.value.toLowerCase().trim()
  if (!q) return products.value
  return products.value.filter(p => 
    p.nombre.toLowerCase().includes(q) || 
    (p.codigo && p.codigo.toLowerCase().includes(q)) ||
    (p.codigo_barras && p.codigo_barras.toLowerCase().includes(q))
  )
})

const calculatedTotal = computed(() => {
  return purchaseForm.items.reduce((sum, item) => sum + (item.cantidad * item.costo_unitario), 0)
})

const calculatedSubtotal = computed(() => {
  return calculatedTotal.value / 1.18
})

const calculatedIGV = computed(() => {
  return calculatedTotal.value - calculatedSubtotal.value
})



let barcodeBuffer = ''
let lastKeyTime = 0

const handleGlobalKeyPress = (e: KeyboardEvent) => {
  const target = e.target as HTMLElement
  if (target.tagName === 'INPUT' || target.tagName === 'SELECT' || target.tagName === 'TEXTAREA') {
    return
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
      
      // Look for the product with this barcode or code in products list
      const foundProduct = products.value.find(p => 
        (p.codigo && p.codigo === barcodeBuffer) || 
        (p.codigo_barras && p.codigo_barras === barcodeBuffer)
      )

      if (foundProduct) {
        // Auto add to detail items or autofill input
        const existsIdx = purchaseForm.items.findIndex(item => item.product_id === foundProduct.id)
        if (existsIdx !== -1) {
          purchaseForm.items[existsIdx].cantidad += 1
        } else {
          purchaseForm.items.push({
            product_id: foundProduct.id,
            cantidad: 1,
            costo_unitario: foundProduct.precio_compra || 1.00
          })
        }
      } else {
        alert(`Producto con código o barras "${barcodeBuffer}" no encontrado en el catálogo.`)
      }
      
      barcodeBuffer = ''
    }
  }
}

onMounted(() => {
  window.addEventListener('keypress', handleGlobalKeyPress)
  loadSuppliers()
  loadProducts()
})

onUnmounted(() => {
  window.removeEventListener('keypress', handleGlobalKeyPress)
})

async function loadProducts() {
  try {
    const res = await axios.get('/products')
    if (res.data.success) products.value = res.data.data || []
  } catch (err) {
    console.error('Error loading products', err)
  }
}

async function handleQuickSupplierSubmit() {
  if (!quickSupplierName.value.trim() || !quickSupplierRUC.value.trim()) return
  quickSupplierSaving.value = true
  try {
    const res = await axios.post('/suppliers', {
      razon_social: quickSupplierName.value,
      ruc: quickSupplierRUC.value,
      direccion: quickSupplierAddress.value
    })
    if (res.data.success) {
      suppliers.value.push(res.data.data)
      purchaseForm.supplier_id = res.data.data.id
      quickSupplierName.value = ''
      quickSupplierRUC.value = ''
      quickSupplierAddress.value = ''
      quickSupplierIsGeneric.value = false
      showQuickSupplierModal.value = false
    }
  } catch (err) {
    alert('Error al agregar el proveedor rápido')
  } finally {
    quickSupplierSaving.value = false
  }
}

function selectProduct(prod: any) {
  itemInput.product_id = prod.id
  itemInput.costo_unitario = prod.precio_compra || 1.00
  productSearchQuery.value = prod.nombre
  showProductDropdown.value = false
}

async function loadSuppliers() {
  try {
    const res = await axios.get('/suppliers')
    if (res.data.success) suppliers.value = res.data.data || []
  } catch (err) {
    console.error('Error loading suppliers', err)
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

async function loadPurchases() {
  try {
    const res = await axios.get('/purchases')
    if (res.data.success) purchases.value = res.data.data || []
  } catch (err) {
    console.error('Error loading purchases', err)
  }
}

function getProductName(pId: string) {
  const p = products.value.find(prod => prod.id === pId)
  return p ? p.nombre : 'Producto desconocido'
}

function addItem() {
  if (!itemInput.product_id || itemInput.cantidad <= 0 || itemInput.costo_unitario <= 0) {
    alert('Ingrese un producto, cantidad y costo unitario válidos.')
    return
  }
  
  // Check if item already in details
  const existsIdx = purchaseForm.items.findIndex(item => item.product_id === itemInput.product_id)
  if (existsIdx !== -1) {
    purchaseForm.items[existsIdx].cantidad += itemInput.cantidad
    // Update buy cost to latest input
    purchaseForm.items[existsIdx].costo_unitario = itemInput.costo_unitario
  } else {
    purchaseForm.items.push({
      product_id: itemInput.product_id,
      cantidad: itemInput.cantidad,
      costo_unitario: itemInput.costo_unitario
    })
  }

  // Reset item inputs and autofocus search
  itemInput.product_id = ''
  itemInput.cantidad = 1
  itemInput.costo_unitario = 1.00
  productSearchQuery.value = ''
  
  if (searchInputRef.value) {
    searchInputRef.value.focus()
  }
}

function removeItem(idx: number) {
  purchaseForm.items.splice(idx, 1)
}

function clearForm() {
  purchaseForm.supplier_id = ''
  purchaseForm.items = []
}

async function submitPurchase() {
  if (!purchaseForm.supplier_id || purchaseForm.items.length === 0) return
  
  try {
    const res = await axios.post('/purchases', purchaseForm)
    if (res.data.success) {
      alert('¡Compra registrada con éxito!')
      clearForm()
      activeSubTab.value = 'history'
      loadPurchases()
    }
  } catch (err) {
    alert('Error al registrar la compra')
  }
}

async function viewDetails(id: string) {
  try {
    const res = await axios.get(`/purchases/${id}`)
    if (res.data.success) {
      selectedPurchase.value = res.data.data
      showModal.value = true
    }
  } catch (err) {
    alert('Error al cargar detalles de la compra')
  }
}

function formatDate(dStr: string) {
  if (!dStr) return '-'
  const d = new Date(dStr)
  return d.toLocaleString('es-PE', { timeZone: 'America/Lima' })
}
</script>
