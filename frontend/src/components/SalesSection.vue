<template>
  <div class="pos-container">
    
    <!-- Mode Switcher & Status -->
    <div class="pos-header">
      <div class="tab-pill">
        <button class="tab-btn" :class="{ active: activeSubTab === 'pos' }" @click="activeSubTab = 'pos'; checkCashSession(); loadStocks()">
          <i class="ti ti-device-desktop"></i><span>Venta nueva</span>
        </button>
        <button class="tab-btn" :class="{ active: activeSubTab === 'history' }" @click="activeSubTab = 'history'; loadSales()">
          <i class="ti ti-history"></i><span>Historial</span>
        </button>
      </div>
      <div v-if="activeSubTab === 'pos'" class="caja-badge" :class="{ closed: !hasActiveCashSession }">
        {{ hasActiveCashSession ? 'Caja abierta' : 'Caja cerrada' }}
      </div>
    </div>

    <!-- POS VIEW -->
    <div v-if="activeSubTab === 'pos'" class="pos-content">
      
      <!-- Left Side: Catalog -->
      <div class="pos-main">
        
        <!-- Search & Quick Stats -->
        <div class="pos-top-row">
          <div class="search-wrap">
            <i class="ti ti-search"></i>
            <input v-model="productSearch" type="text" placeholder="Buscar producto por nombre o código..." class="search-input">
          </div>
          <div class="customer-wrap">
            <i class="ti ti-user"></i>
            <select v-model="saleForm.customer_id" class="customer-select">
              <option value="">Público General</option>
              <option v-for="cust in customers" :key="cust.id" :value="cust.id">{{ cust.nombre }}</option>
            </select>
            <button class="btn-add-customer" @click="openCustomerModal" title="Nuevo Cliente">
              <i class="ti ti-user-plus"></i>
            </button>
          </div>
        </div>

        <div v-if="!hasActiveCashSession" class="warning-card">
          <div class="warning-icon"><i class="ti ti-lock"></i></div>
          <h3>Caja cerrada</h3>
          <p>Debes abrir una sesión de caja antes de realizar ventas.</p>
          <button @click="emit('navigate', 'cash')" class="btn-primary">Ir a Apertura de Caja</button>
        </div>

        <div v-else class="products-grid">
          <div v-for="item in filteredStocks" :key="item.product_id" 
               class="prod-card" @click="quickAddItem(item)"
               :class="{ 'out-of-stock': item.stock_actual <= 0 }">
            <span class="stock-tag" :class="item.stock_actual > 5 ? 'stock-ok' : 'stock-low'">
              {{ item.stock_actual.toFixed(2) }} {{ item.unidad_medida || 'un.' }}
            </span>
            <div class="prod-cat">{{ item.category_nombre || 'General' }}</div>
            <div class="prod-name">{{ item.product_nombre }}</div>
            <div class="prod-price-row">
              <span class="prod-price">S/. {{ item.precio_venta.toFixed(2) }}</span>
              <span class="prod-unit">/ {{ item.unidad_medida || 'un.' }}</span>
            </div>
            <button class="add-btn"><i class="ti ti-plus"></i>Agregar</button>
            <div v-if="getItemQtyInCart(item.product_id)" class="qty-badge">{{ getItemQtyInCart(item.product_id) }}</div>
          </div>
          
          <div v-if="filteredStocks.length === 0" class="empty-state">
            <i class="ti ti-search-off"></i>
            <p>No se encontraron productos</p>
          </div>
        </div>
      </div>

      <!-- Right Side: Cart -->
      <aside class="cart-panel" :class="{ open: isCartOpen }">
        <div class="cart-head">
          <div class="cart-icon"><i class="ti ti-shopping-bag"></i></div>
          <div>
            <div class="cart-title">Carrito</div>
            <div class="cart-subtitle">{{ saleForm.items.length }} ítems</div>
          </div>
          <button class="cart-close" @click="isCartOpen = false"><i class="ti ti-x"></i></button>
        </div>

        <div class="cart-body">
          <div v-for="(item, idx) in saleForm.items" :key="idx" class="cart-item">
            <div class="ci-info">
              <div class="ci-name">{{ getProductName(item.product_id) }}</div>
              <div class="ci-price">S/. {{ item.precio_unitario.toFixed(2) }} c/u</div>
              
              <div v-if="!item.is_dimensional" class="ci-controls">
                <button class="qty-btn" @click="item.cantidad = Math.max(1, item.cantidad - 1)">-</button>
                <span class="qty-val">{{ item.cantidad }}</span>
                <button class="qty-btn" @click="item.cantidad++">+</button>
              </div>
              <div v-else class="ci-formula">{{ getDimensionalFormula(item) }}</div>
            </div>
            
            <div class="ci-right">
              <button class="ci-remove" @click="removeItem(idx)"><i class="ti ti-trash"></i></button>
              <div class="ci-total">S/. {{ (((item.cantidad || 0) * (item.precio_unitario || 0)) - (item.descuento || 0)).toFixed(2) }}</div>
            </div>

            <!-- Dimensional Inputs -->
            <div v-if="item.is_dimensional" class="ci-dim-grid">
              <input v-if="['m', 'm2', 'm3'].includes(item.unidad_medida)" v-model.number="item.alto" type="number" step="0.01" placeholder="H" class="dim-input" @input="updateItemQuantity(item)">
              <input v-if="['m2', 'm3'].includes(item.unidad_medida)" v-model.number="item.ancho" type="number" step="0.01" placeholder="W" class="dim-input" @input="updateItemQuantity(item)">
              <input v-if="['m3'].includes(item.unidad_medida)" v-model.number="item.espesor" type="number" step="0.01" placeholder="E" class="dim-input" @input="updateItemQuantity(item)">
              <input v-model.number="item.cantidad_piezas" type="number" step="1" placeholder="N" class="dim-input" @input="updateItemQuantity(item)">
              <button v-if="['m2', 'm3'].includes(item.unidad_medida)" @click="swapDimensions(item); updateItemQuantity(item)" class="dim-swap"><i class="ti ti-arrows-exchange-2"></i></button>
            </div>
            <div class="ci-discount">
              <span>Desc. S/.</span>
              <input v-model.number="item.descuento" type="number" step="0.10" class="discount-input" @input="validateDiscount(item)">
            </div>
          </div>

          <div v-if="saleForm.items.length === 0" class="cart-empty">
            <i class="ti ti-shopping-cart-off"></i>
            <p>Tu carrito está vacío</p>
          </div>
        </div>

        <div class="cart-foot">
          <div class="totals">
            <div class="total-row"><span>Subtotal</span><span>S/. {{ calculatedSubtotal.toFixed(2) }}</span></div>
            <div class="total-row"><span>IGV (18%)</span><span>S/. {{ calculatedIGV.toFixed(2) }}</span></div>
            <div class="total-main">
              <span>Total a pagar</span>
              <span class="total-amount">S/. {{ calculatedTotal.toFixed(2) }}</span>
            </div>
          </div>

          <!-- Change Calculator -->
          <div v-if="saleForm.metodo_pago === 'EFECTIVO'" class="change-calc">
            <div class="calc-row">
              <label>Paga con S/.</label>
              <input v-model.number="pagaCon" type="number" step="0.10" class="calc-input" placeholder="0.00">
            </div>
            <div class="calc-row highlighted">
              <label>Vuelto S/.</label>
              <span class="vuelto-amount">S/. {{ vuelto.toFixed(2) }}</span>
            </div>
          </div>

          <div class="pay-options">
            <div class="pay-group">
              <label><i class="ti ti-file-text"></i> Doc.</label>
              <select v-model="saleForm.tipo_documento" class="pay-select">
                <option value="03">Boleta</option>
                <option value="01">Factura</option>
                <option value="NV">Nota Venta</option>
              </select>
            </div>
            <div class="pay-group">
              <label><i class="ti ti-credit-card"></i> Pago</label>
              <select v-model="saleForm.metodo_pago" class="pay-select">
                <option value="EFECTIVO">Efectivo</option>
                <option value="TARJETA">Tarjeta</option>
                <option value="YAPE">Yape / Plin</option>
              </select>
            </div>
          </div>

          <div class="cart-actions">
            <button class="btn-clear" @click="clearForm">Limpiar</button>
            <button class="btn-pay" :disabled="!canSubmitSale" @click="submitSale">
              <i class="ti ti-shopping-cart-check"></i> Pagar ahora
            </button>
          </div>
          <p v-if="!canSubmitSale && saleForm.items.length > 0" class="pay-warning">
            <i class="ti ti-alert-triangle"></i>
            {{ saleForm.tipo_documento === '01' ? 'Factura requiere un cliente con RUC' : 'Venta ≥ S/. 700 requiere identificar al cliente' }}
          </p>
        </div>
      </aside>
    </div>

    <!-- HISTORY VIEW -->
    <div v-else class="history-content">
      <div v-for="sale in sales" :key="sale.id" class="history-card" @click="viewDetails(sale.id)">
        <div class="h-icon"><i class="ti ti-receipt"></i></div>
        <div class="h-info">
          <div class="h-row">
            <span class="h-id">#{{ sale.id.substring(0, 8).toUpperCase() }}</span>
            <span class="h-total">S/. {{ sale.total.toFixed(2) }}</span>
          </div>
          <div class="h-customer">{{ sale.customer?.nombre || 'Público General' }}</div>
          <div class="h-date">{{ formatDate(sale.created_at) }}</div>
        </div>
        <i class="ti ti-chevron-right"></i>
      </div>
    </div>

    <!-- Mobile Cart Toggle -->
    <button v-if="activeSubTab === 'pos' && !isCartOpen" class="cart-toggle" @click="isCartOpen = true">
      <i class="ti ti-shopping-bag"></i>
      <span v-if="saleForm.items.length > 0" :key="badgeKey" class="cart-badge animate-pop">{{ saleForm.items.length }}</span>
    </button>

    <!-- Modal Detail -->
    <div v-if="showModal" class="modal-overlay" @click="showModal = false">
      <div class="modal-card" @click.stop>
        <div class="modal-head">
          <h3>Venta #{{ selectedSale?.sale?.id.substring(0, 8).toUpperCase() }}</h3>
          <button @click="showModal = false"><i class="ti ti-x"></i></button>
        </div>
        <div class="modal-body">
          <!-- Details summary -->
          <div class="detail-summary">
            <div class="detail-group">
              <label>Cliente</label>
              <p>{{ selectedSale?.sale?.customer?.nombre || 'Público General' }}</p>
            </div>
            <div class="detail-group text-right">
              <label>Fecha</label>
              <p>{{ formatDate(selectedSale?.sale?.created_at) }}</p>
            </div>
          </div>

          <!-- Electronic links if any -->
          <div v-if="selectedSale?.has_electronic" class="electronic-card">
            <div class="e-head">
              <h4>{{ selectedSale.electronic_document.serie }}-{{ selectedSale.electronic_document.numero }}</h4>
              <span class="e-status">{{ selectedSale.electronic_document.estado }}</span>
            </div>
            <div class="e-actions">
              <button @click="downloadFile(selectedSale.electronic_document.document_uuid, 'pdf')">PDF</button>
              <button @click="downloadFile(selectedSale.electronic_document.document_uuid, 'xml')">XML</button>
            </div>
          </div>

          <!-- Items list -->
          <div class="modal-items">
            <div v-for="item in selectedSale?.items" :key="item.id" class="m-item">
              <div class="m-qty">{{ item.cantidad.toFixed(2) }}</div>
              <div class="m-info">
                <p class="m-name">{{ item.product_nombre }}</p>
                <p v-if="item.is_dimensional" class="m-dim">Dim: {{ item.alto }}x{{ item.ancho }}x{{ item.espesor }}</p>
              </div>
              <div class="m-total">S/. {{ item.total.toFixed(2) }}</div>
            </div>
          </div>
        </div>
        <div class="modal-foot">
          <div class="m-grand-total">
            <label>Total</label>
            <p>S/. {{ selectedSale?.sale?.total.toFixed(2) }}</p>
          </div>
          <div class="m-actions">
            <button v-if="selectedSale?.sale?.estado !== 'annulled'" @click="openAnnulmentConfirm" class="btn-annul">Anular</button>
            <button @click="showModal = false" class="btn-close">Cerrar</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Annulment Modal -->
    <div v-if="showAnnulmentModal" class="modal-overlay" @click="showAnnulmentModal = false">
      <div class="modal-mini" @click.stop>
        <h3>Anular Comprobante</h3>
        <select v-model="annulmentReason" class="modal-select">
          <option value="">-- Seleccionar Motivo --</option>
          <option value="ANULACION DE LA OPERACION">Anulación de la operación</option>
          <option value="ERROR EN EL RUC O NOMBRE">Error en el RUC o Nombre</option>
          <option value="DEVOLUCION TOTAL DE MERCADERIA">Devolución total</option>
        </select>
        <div class="modal-actions">
          <button @click="showAnnulmentModal = false">Cancelar</button>
          <button @click="submitAnnulment" :disabled="!annulmentReason || annulling" class="btn-confirm">Confirmar</button>
        </div>
      </div>
    </div>

    <!-- Quick Customer Modal -->
    <div v-if="showCustomerModal" class="modal-overlay" @click="showCustomerModal = false">
      <div class="modal-card" style="max-width: 500px;" @click.stop>
        <div class="modal-head">
          <h3>Nuevo Cliente</h3>
          <button @click="showCustomerModal = false"><i class="ti ti-x"></i></button>
        </div>
        <div class="modal-body">
          <div class="customer-form">
            <div class="form-row">
              <div class="form-group">
                <label>Tipo Doc.</label>
                <select v-model="newCustomer.tipo_documento" class="modal-select">
                  <option value="DNI">DNI</option>
                  <option value="RUC">RUC</option>
                </select>
              </div>
              <div class="form-group" style="flex: 2;">
                <label>N° Documento</label>
                <div class="input-search-group">
                  <input v-model="newCustomer.numero_documento" type="text" maxlength="11" class="modal-input" placeholder="8 o 11 dígitos">
                  <button @click="searchIdentity" :disabled="isSearching" class="btn-search-api">
                    <i v-if="!isSearching" class="ti ti-search"></i>
                    <i v-else class="ti ti-loader animate-spin"></i>
                  </button>
                </div>
              </div>
            </div>
            <div class="form-group">
              <label>Nombre / Razón Social</label>
              <input v-model="newCustomer.nombre" type="text" class="modal-input" placeholder="Nombre completo">
            </div>
            <div class="form-group">
              <label>Dirección</label>
              <input v-model="newCustomer.direccion" type="text" class="modal-input" placeholder="Dirección opcional">
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>Teléfono</label>
                <input v-model="newCustomer.telefono" type="text" class="modal-input">
              </div>
              <div class="form-group">
                <label>Email</label>
                <input v-model="newCustomer.email" type="email" class="modal-input">
              </div>
            </div>
          </div>
        </div>
        <div class="modal-foot">
          <button @click="showCustomerModal = false" class="btn-cancel">Cancelar</button>
          <button @click="saveCustomer" :disabled="!newCustomer.nombre || !newCustomer.numero_documento || savingCustomer" class="btn-save-customer">
            {{ savingCustomer ? 'Guardando...' : 'Registrar y Seleccionar' }}
          </button>
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
const badgeKey = ref(0) // Trigger for animation
const hasActiveCashSession = ref(true)
const activeSession = ref<any>(null)

const pagaCon = ref<number | null>(null)
const vuelto = computed(() => {
  if (!pagaCon.value || pagaCon.value < calculatedTotal.value) return 0
  return pagaCon.value - calculatedTotal.value
})

const saleForm = reactive({
  customer_id: '',
  tipo_documento: '03',
  metodo_pago: 'EFECTIVO',
  items: [] as any[]
})

const showModal = ref(false)
const selectedSale = ref<any>(null)

// Quick Customer Registration
const showCustomerModal = ref(false)
const isSearching = ref(false)
const savingCustomer = ref(false)
const newCustomer = reactive({
  tipo_documento: 'DNI',
  numero_documento: '',
  nombre: '',
  direccion: '',
  telefono: '',
  email: ''
})

function openCustomerModal() {
  newCustomer.tipo_documento = 'DNI'
  newCustomer.numero_documento = ''
  newCustomer.nombre = ''
  newCustomer.direccion = ''
  newCustomer.telefono = ''
  newCustomer.email = ''
  showCustomerModal.value = true
}

async function searchIdentity() {
  if (!newCustomer.numero_documento) return
  isSearching.value = true
  try {
    const type = newCustomer.tipo_documento.toLowerCase()
    const res = await axios.get(`/public/${type}/${newCustomer.numero_documento}`)
    if (res.data.success) {
      const d = res.data.data || {}
      if (type === 'dni') {
        newCustomer.nombre = d.nombre || d.full_name || ''
      } else {
        newCustomer.nombre = d.razon_social || d.nombre || d.nombre_o_razon_social || d.razonSocial || ''
        
        let address = d.direccion || d.direccion_completa || ''
        if (!address || address.trim() === '-') {
          const parts = []
          if (d.departamento && d.departamento !== '-') parts.push(d.departamento)
          if (d.provincia && d.provincia !== '-') parts.push(d.provincia)
          if (d.distrito && d.distrito !== '-') parts.push(d.distrito)
          address = parts.join(' - ')
        }
        newCustomer.direccion = address
      }
    }
  } catch (err) {
    alert('No se pudo encontrar el documento')
  } finally {
    isSearching.value = false
  }
}

async function saveCustomer() {
  savingCustomer.value = true
  try {
    const res = await axios.post('/customers', newCustomer)
    if (res.data.success) {
      await loadCustomers()
      saleForm.customer_id = res.data.data.id
      showCustomerModal.value = false
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al guardar cliente')
  } finally {
    savingCustomer.value = false
  }
}

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
  // Golden Rule: Subtotal = Total / 1.18 (For Gravado items)
  return calculatedTotal.value / 1.18
})

const calculatedIGV = computed(() => {
  // Golden Rule: IGV = Total - Subtotal
  return calculatedTotal.value - calculatedSubtotal.value
})

// Senior Validation: Check if sale can be processed
const canSubmitSale = computed(() => {
  if (saleForm.items.length === 0) return false
  
  // Rule: Factura (01) requires customer with RUC (Tipo 6)
  if (saleForm.tipo_documento === '01') {
    const cust = customers.value.find(c => c.id === saleForm.customer_id)
    if (!cust || cust.tipo_documento !== 'RUC') return false
  }

  // Rule: Boleta (03) > 700 requires identification (not 00000000)
  if (saleForm.tipo_documento === '03' && calculatedTotal.value >= 700) {
    const cust = customers.value.find(c => c.id === saleForm.customer_id)
    if (!cust || cust.numero_documento === '00000000') return false
  }

  return true
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

function validateDiscount(item: any) {
  if (item.descuento < 0) item.descuento = 0
}

function updateItemQuantity(item: any) {
  if (!item.is_dimensional) return

  const piezas = Number(item.cantidad_piezas) || 1
  const alto = Number(item.alto) || 1
  const ancho = Number(item.ancho) || 1
  const espesor = Number(item.espesor) || 1

  switch (item.unidad_medida) {
    case 'm':
      item.cantidad = alto * piezas
      break
    case 'm2':
      item.cantidad = alto * ancho * piezas
      break
    case 'm3':
      item.cantidad = alto * ancho * espesor * piezas
      break
    default:
      item.cantidad = piezas
  }
}

function quickAddItem(stock: any) {
  if (stock.is_dimensional) {
    const isMetric = ['m', 'm2', 'm3'].includes(stock.unidad_medida)
    const defaultAlto = isMetric ? 1.0 : 1.0
    const defaultAncho = ['m2', 'm3'].includes(stock.unidad_medida) ? 1.0 : 1.0
    const defaultEspesor = ['m3'].includes(stock.unidad_medida) ? 0.1 : 1.0
    
    const newItem = reactive({
      product_id: stock.product_id,
      cantidad: 1.0,
      precio_unitario: stock.precio_venta,
      descuento: 0,
      alto: defaultAlto,
      ancho: defaultAncho,
      espesor: defaultEspesor,
      cantidad_piezas: 1,
      is_dimensional: true,
      unidad_medida: stock.unidad_medida
    })
    updateItemQuantity(newItem)
    saleForm.items.push(newItem)
    isCartOpen.value = true
    badgeKey.value++
    return
  }
  const alreadyAdded = saleForm.items.find(item => item.product_id === stock.product_id && !item.is_dimensional)
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
  badgeKey.value++
}

let barcodeBuffer = ''
let lastKeyTime = 0

const handleGlobalKeyPress = (e: KeyboardEvent) => {
  const target = e.target as HTMLElement
  if (target.tagName === 'INPUT' || target.tagName === 'SELECT' || target.tagName === 'TEXTAREA') {
    if (target.getAttribute('placeholder')?.includes('Buscar producto')) return
  }

  const currentTime = Date.now()
  if (currentTime - lastKeyTime > 50) barcodeBuffer = ''
  lastKeyTime = currentTime

  if (e.key !== 'Enter') {
    if (e.key.length === 1 && /[0-9a-zA-Z]/.test(e.key)) barcodeBuffer += e.key
  } else {
    if (barcodeBuffer.length >= 4) {
      e.preventDefault()
      const foundStock = stocks.value.find(s => 
        (s.product_codigo && s.product_codigo === barcodeBuffer) || 
        (s.product_codigo_barras && s.product_codigo_barras === barcodeBuffer)
      )
      if (foundStock) quickAddItem(foundStock)
      else productSearch.value = barcodeBuffer
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
    hasActiveCashSession.value = false
  }
}

async function loadCustomers() {
  try {
    const res = await axios.get('/customers')
    if (res.data.success) customers.value = res.data.data
  } catch (err) {}
}

async function loadStocks() {
  try {
    const res = await axios.get('/stocks')
    if (res.data.success) stocks.value = res.data.data || []
  } catch (err) {}
}

async function loadSales() {
  try {
    const res = await axios.get('/sales')
    if (res.data.success) sales.value = res.data.data || []
  } catch (err) {}
}

function swapDimensions(item: any) {
  const temp = item.alto
  item.alto = item.ancho
  item.ancho = temp
}

function getDimensionalFormula(item: any): string {
  const piezasStr = item.cantidad_piezas > 1 ? ` × ${item.cantidad_piezas}` : ''
  const unit = item.unidad_medida || 'm'
  if (unit === 'm') return `${(item.alto || 1).toFixed(2)}m${piezasStr} = ${item.cantidad.toFixed(2)}m`
  if (unit === 'm2') return `[${(item.alto || 1).toFixed(2)}x${(item.ancho || 1).toFixed(2)}]${piezasStr} = ${item.cantidad.toFixed(2)}m²`
  if (unit === 'm3') return `[${(item.alto || 1).toFixed(2)}x${(item.ancho || 1).toFixed(2)}x${(item.espesor || 1).toFixed(2)}]${piezasStr} = ${item.cantidad.toFixed(2)}m³`
  return `${item.cantidad.toFixed(2)} ${unit}`
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
  pagaCon.value = null
  isCartOpen.value = false
}

async function submitSale() {
  if (saleForm.items.length === 0) return
  const payload = { ...saleForm, customer_id: saleForm.customer_id === '' ? null : saleForm.customer_id }
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
    const res = await axios.get(`/billing/files/${uuid}`)
    if (res.data.files && res.data.files[type]) window.open(res.data.files[type], '_blank')
  } catch (err) {
    alert('Obteniendo archivo...')
  }
}

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
    const res = await axios.post(`/sales/${selectedSale.value.sale.id}/credit-note`, { motivo: annulmentReason.value })
    if (res.data.success) {
      alert('¡Venta anulada con éxito!')
      showAnnulmentModal.value = false
      showModal.value = false
      loadSales()
    }
  } catch (err: any) {
    alert('Error al anular: ' + (err.response?.data?.error || err.message))
  } finally {
    annulling.value = false
  }
}

function formatDate(dStr: string) {
  if (!dStr) return '-'
  const d = new Date(dStr)
  return d.toLocaleString('es-PE', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.pos-container { height: 100%; display: flex; flex-direction: column; gap: 20px; }

.pos-header { display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; }
.tab-pill { background: var(--surface2); padding: 4px; border-radius: 50px; display: flex; gap: 4px; }
.tab-btn {
  padding: 8px 20px; border-radius: 50px; border: none; background: transparent;
  color: var(--text2); font-size: 13px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; gap: 8px; transition: all var(--transition);
}
.tab-btn.active { background: var(--accent); color: #fff; box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3); }

.caja-badge {
  padding: 8px 16px; border-radius: 50px; background: var(--green-dim); color: var(--green);
  font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em;
  display: flex; align-items: center; gap: 8px; border: 1px solid rgba(16, 185, 129, 0.2);
}
.caja-badge::before { content: ''; width: 6px; height: 6px; border-radius: 50%; background: var(--green); box-shadow: 0 0 6px var(--green); }
.caja-badge.closed { background: var(--red-dim); color: var(--red); border-color: rgba(239, 68, 68, 0.2); }
.caja-badge.closed::before { background: var(--red); box-shadow: 0 0 6px var(--red); }

.pos-content { flex: 1; display: flex; overflow: hidden; gap: 24px; }
.pos-main { flex: 1; display: flex; flex-direction: column; min-width: 0; gap: 20px; }

.pos-top-row { display: flex; gap: 12px; flex-shrink: 0; }
.search-wrap, .customer-wrap { position: relative; flex: 1; }
.search-wrap i, .customer-wrap i { position: absolute; left: 14px; top: 50%; transform: translateY(-50%); color: var(--text3); }
.search-input, .customer-select {
  width: 100%; height: 46px; padding: 0 16px 0 42px; border-radius: var(--radius);
  border: 1px solid var(--border); background: var(--surface); color: var(--text);
  font-family: inherit; font-size: 13.5px; outline: none; transition: all var(--transition);
}
.search-input:focus, .customer-select:focus { border-color: var(--accent); box-shadow: 0 0 0 4px var(--accent-dim); }
.customer-select { appearance: none; cursor: pointer; }

.products-grid {
  flex: 1; overflow-y: auto; display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px; align-content: start; padding-bottom: 20px;
}
.prod-card {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 16px; cursor: pointer; transition: all var(--transition); position: relative;
  display: flex; flex-direction: column; gap: 8px;
}
.prod-card:hover { border-color: var(--accent); transform: translateY(-4px); box-shadow: 0 10px 20px rgba(0,0,0,0.05); }
.prod-card.out-of-stock { opacity: 0.6; cursor: not-allowed; }

.stock-tag {
  position: absolute; top: 12px; right: 12px; font-size: 9px; font-weight: 800;
  padding: 3px 8px; border-radius: 6px; text-transform: uppercase;
}
.stock-ok { background: var(--green-dim); color: var(--green); }
.stock-low { background: var(--red-dim); color: var(--red); }

.prod-cat { font-size: 9px; font-weight: 700; color: var(--text3); text-transform: uppercase; letter-spacing: 0.05em; }
.prod-name { font-size: 13px; font-weight: 700; color: var(--text); line-height: 1.4; flex: 1; min-height: 36px; }
.prod-price-row { display: flex; align-items: baseline; gap: 4px; }
.prod-price { font-size: 18px; font-weight: 800; color: var(--text); }
.prod-unit { font-size: 11px; color: var(--text2); }

.add-btn {
  width: 100%; padding: 8px; border-radius: var(--radius-sm); border: none;
  background: var(--surface2); color: var(--accent); font-weight: 700; font-size: 12px;
  cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 6px;
  transition: all var(--transition);
}
.prod-card:hover .add-btn { background: var(--accent); color: #fff; }
.qty-badge {
  position: absolute; -top: 8px; -left: 8px; width: 22px; height: 22px;
  background: var(--accent); color: #fff; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 800;
  box-shadow: 0 4px 8px rgba(79, 70, 229, 0.4); border: 2px solid var(--surface);
}

.cart-panel {
  width: 360px; background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius); display: flex; flex-direction: column; overflow: hidden;
  transition: transform var(--transition);
}
.cart-head { padding: 20px; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 12px; }
.cart-icon {
  width: 40px; height: 40px; border-radius: 12px; background: var(--accent);
  color: #fff; display: flex; align-items: center; justify-content: center; font-size: 20px;
}
.cart-title { font-size: 16px; font-weight: 800; }
.cart-subtitle { font-size: 11px; color: var(--text2); text-transform: uppercase; font-weight: 700; }
.cart-close { display: none; margin-left: auto; background: none; border: none; color: var(--text3); font-size: 20px; }

.cart-body { flex: 1; overflow-y: auto; padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.cart-item {
  background: var(--surface2); border: 1px solid var(--border); border-radius: var(--radius-sm);
  padding: 12px; display: flex; flex-direction: column; gap: 10px;
}
.ci-info { flex: 1; }
.ci-name { font-size: 13px; font-weight: 700; line-height: 1.3; }
.ci-price { font-size: 11px; color: var(--text2); margin-top: 2px; }

.ci-controls { display: flex; align-items: center; gap: 10px; margin-top: 8px; }
.qty-btn {
  width: 26px; height: 26px; border-radius: 6px; border: 1px solid var(--border);
  background: var(--surface); color: var(--text); font-weight: 800; cursor: pointer;
}
.qty-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
.qty-val { font-size: 14px; font-weight: 800; min-width: 24px; text-align: center; }

.ci-formula { font-size: 10px; font-family: monospace; color: var(--accent); font-weight: 700; margin-top: 8px; }

.ci-right { display: flex; align-items: center; justify-content: space-between; }
.ci-remove { background: none; border: none; color: var(--text3); cursor: pointer; font-size: 16px; }
.ci-remove:hover { color: var(--red); }
.ci-total { font-size: 14px; font-weight: 800; color: var(--text); }

.ci-dim-grid { display: grid; grid-template-columns: repeat(4, 1fr) auto; gap: 6px; margin-top: 4px; }
.dim-input {
  width: 100%; padding: 4px; font-size: 10px; font-weight: 800; border-radius: 4px;
  border: 1px solid var(--border); background: var(--surface); text-align: center;
}
.dim-swap { background: none; border: none; color: var(--text3); cursor: pointer; }

.ci-discount { display: flex; align-items: center; justify-content: space-between; font-size: 10px; font-weight: 700; color: var(--text3); }
.discount-input { width: 60px; padding: 2px 6px; font-size: 11px; font-weight: 800; border-radius: 4px; border: 1px solid var(--border); text-align: right; }

.cart-foot { padding: 20px; border-top: 1px solid var(--border); background: var(--surface); }
.totals { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.total-row { display: flex; justify-content: space-between; font-size: 12px; color: var(--text2); font-weight: 600; }
.total-main {
  display: flex; justify-content: space-between; align-items: center; margin-top: 8px;
  padding-top: 12px; border-top: 1px solid var(--border);
}
.total-main span:first-child { font-size: 14px; font-weight: 700; }
.total-amount { font-size: 26px; font-weight: 900; color: var(--accent); }

.change-calc {
  background: var(--surface2);
  border-radius: var(--radius-sm);
  padding: 12px;
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.calc-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.calc-row label {
  font-size: 11px;
  font-weight: 800;
  color: var(--text2);
  text-transform: uppercase;
}
.calc-input {
  width: 100px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 6px 10px;
  font-size: 14px;
  font-weight: 800;
  text-align: right;
  outline: none;
}
.calc-input:focus { border-color: var(--accent); }
.calc-row.highlighted {
  padding-top: 8px;
  border-top: 1px dashed var(--border);
}
.vuelto-amount {
  font-size: 18px;
  font-weight: 900;
  color: var(--green);
}

.pay-options { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 16px; }
.pay-group { display: flex; flex-direction: column; gap: 4px; }
.pay-group label {
  font-size: 10px;
  font-weight: 800;
  color: var(--text3);
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 4px;
}
.pay-group label i { font-size: 12px; }
.pay-select {
  width: 100%; padding: 8px; border-radius: var(--radius-sm); border: 1px solid var(--border);
  background: var(--surface2); font-size: 11px; font-weight: 700; outline: none;
}

.cart-actions { display: flex; gap: 8px; }
.btn-clear { flex: 1; padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); background: transparent; font-weight: 700; cursor: pointer; }
.btn-pay {
  flex: 2; padding: 12px; border-radius: var(--radius-sm); border: none; background: var(--accent);
  color: #fff; font-weight: 800; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}
.btn-pay:disabled { opacity: 0.5; cursor: not-allowed; box-shadow: none; }

.pay-warning {
  margin-top: 10px;
  padding: 8px 12px;
  background: var(--red-dim);
  color: var(--red);
  font-size: 11px;
  font-weight: 700;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  border: 1px solid rgba(239, 68, 68, 0.1);
}

/* History */
.history-content { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }
.history-card {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 16px; display: flex; align-items: center; gap: 16px; cursor: pointer; transition: all var(--transition);
}
.history-card:hover { border-color: var(--accent); transform: scale(1.02); }
.h-icon { width: 44px; height: 44px; border-radius: 12px; background: var(--accent-dim); color: var(--accent); display: flex; align-items: center; justify-content: center; font-size: 20px; }
.h-info { flex: 1; display: flex; flex-direction: column; gap: 2px; }
.h-row { display: flex; justify-content: space-between; align-items: baseline; }
.h-id { font-size: 13px; font-weight: 800; color: var(--text); }
.h-total { font-size: 15px; font-weight: 900; color: var(--text); }
.h-customer { font-size: 11px; font-weight: 600; color: var(--text2); }
.h-date { font-size: 10px; color: var(--text3); margin-top: 4px; text-transform: uppercase; letter-spacing: 0.05em; }

/* Mobile */
@media (max-width: 900px) {
  .pos-content { flex-direction: column; }
  .cart-panel { position: fixed; right: 0; top: 0; bottom: 0; z-index: 200; transform: translateX(100%); width: 100%; border-radius: 0; }
  .cart-panel.open { transform: translateX(0); }
  .cart-close { display: block; }
}

.cart-toggle {
  position: fixed; bottom: 30px; right: 30px; width: 64px; height: 64px; border-radius: 50%;
  background: var(--accent); color: #fff; border: none; font-size: 28px;
  box-shadow: 0 10px 25px rgba(79, 70, 229, 0.5); z-index: 100; cursor: pointer;
}
.cart-badge {
  position: absolute; top: 0; right: 0; background: var(--red); color: #fff;
  font-size: 11px; font-weight: 900; min-width: 20px; height: 20px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; border: 2px solid var(--bg);
}

/* Modals */
.modal-overlay { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.6); backdrop-filter: blur(4px); z-index: 300; display: flex; align-items: center; justify-content: center; padding: 20px; }
.modal-card { background: var(--surface); border-radius: 24px; width: 100%; max-width: 600px; display: flex; flex-direction: column; max-height: 90vh; }
.modal-head { padding: 24px; border-bottom: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; }
.modal-head h3 { font-size: 18px; font-weight: 900; }
.modal-head button { background: none; border: none; font-size: 24px; color: var(--text3); cursor: pointer; }
.modal-body { flex: 1; overflow-y: auto; padding: 24px; }

.detail-summary { display: flex; justify-content: space-between; margin-bottom: 24px; }
.detail-group label { display: block; font-size: 10px; font-weight: 800; color: var(--text3); text-transform: uppercase; margin-bottom: 4px; }
.detail-group p { font-size: 14px; font-weight: 700; color: var(--text); }

.electronic-card { background: var(--text); color: #fff; padding: 20px; border-radius: 20px; margin-bottom: 24px; }
.e-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.e-status { font-size: 10px; font-weight: 800; padding: 4px 10px; background: rgba(255,255,255,0.1); border-radius: 50px; }
.e-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.e-actions button { padding: 10px; border-radius: 12px; border: none; background: rgba(255,255,255,0.1); color: #fff; font-weight: 800; cursor: pointer; }

.modal-items { display: flex; flex-direction: column; gap: 12px; }
.m-item { display: flex; align-items: center; gap: 16px; padding: 12px; background: var(--surface2); border-radius: 16px; }
.m-qty { width: 40px; height: 40px; border-radius: 10px; background: var(--surface); display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 13px; }
.m-info { flex: 1; }
.m-name { font-size: 13px; font-weight: 700; }
.m-dim { font-size: 10px; color: var(--text2); font-weight: 600; margin-top: 2px; }
.m-total { font-size: 14px; font-weight: 800; }

.modal-foot { padding: 24px; border-top: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; }
.m-grand-total label { font-size: 10px; font-weight: 800; color: var(--text3); text-transform: uppercase; }
.m-grand-total p { font-size: 24px; font-weight: 900; color: var(--text); }
.m-actions { display: flex; gap: 10px; }
.btn-annul { padding: 12px 24px; border-radius: 16px; border: none; background: var(--red); color: #fff; font-weight: 800; cursor: pointer; }
.btn-close { padding: 12px 24px; border-radius: 16px; border: none; background: var(--text); color: #fff; font-weight: 800; cursor: pointer; }

.modal-mini { background: var(--surface); padding: 24px; border-radius: 24px; width: 100%; max-width: 400px; display: flex; flex-direction: column; gap: 16px; }
.modal-select { width: 100%; padding: 12px; border-radius: 12px; border: 1px solid var(--border); background: var(--surface2); font-weight: 700; }
.btn-confirm { padding: 12px; border-radius: 12px; border: none; background: var(--red); color: #fff; font-weight: 800; }

.warning-card { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; text-align: center; gap: 16px; background: var(--surface); border-radius: 32px; border: 1px solid var(--border); padding: 48px; }
.warning-icon { width: 80px; height: 80px; border-radius: 24px; background: var(--red-dim); color: var(--red); display: flex; align-items: center; justify-content: center; font-size: 40px; }
.btn-primary { padding: 14px 32px; border-radius: 16px; border: none; background: var(--accent); color: #fff; font-weight: 800; cursor: pointer; box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3); }

/* Quick Customer Registration Styles */
.btn-add-customer {
  position: absolute; right: 8px; top: 50%; transform: translateY(-50%);
  width: 32px; height: 32px; border-radius: 10px; border: none;
  background: var(--accent-dim); color: var(--accent); cursor: pointer;
  display: flex; align-items: center; justify-content: center; transition: all 0.2s;
}
.btn-add-customer:hover { background: var(--accent); color: #fff; }

.customer-form { display: flex; flex-direction: column; gap: 16px; }
.form-row { display: flex; gap: 12px; }
.form-group { display: flex; flex-direction: column; gap: 6px; flex: 1; }
.form-group label { font-size: 11px; font-weight: 800; color: var(--text3); text-transform: uppercase; }
.modal-input {
  width: 100%; padding: 12px; border-radius: 12px; border: 1.5px solid var(--border);
  background: var(--surface2); font-family: inherit; font-size: 13px; outline: none; transition: all 0.2s;
}
.modal-input:focus { border-color: var(--accent); background: #fff; }

.input-search-group { position: relative; display: flex; }
.btn-search-api {
  position: absolute; right: 6px; top: 6px; bottom: 6px; width: 34px;
  border-radius: 8px; border: none; background: var(--text); color: #fff;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
}
.btn-search-api:disabled { opacity: 0.5; }

.btn-cancel { padding: 12px 24px; border-radius: 16px; border: 1px solid var(--border); background: #fff; font-weight: 800; cursor: pointer; }
.btn-save-customer { flex: 1; padding: 12px 24px; border-radius: 16px; border: none; background: var(--green); color: #fff; font-weight: 800; cursor: pointer; }
.btn-save-customer:disabled { opacity: 0.5; }

@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.animate-spin { animation: spin 1s linear infinite; }

.animate-pop { animation: pop 0.3s ease-out; }
@keyframes pop {
  0% { transform: scale(1); }
  50% { transform: scale(1.4); }
  100% { transform: scale(1); }
}
</style>
