<template>
  <div class="space-y-6">
    <div class="flex gap-4 border-b border-slate-200 pb-3">
      <button
        @click="activeSubTab = 'inventory'; loadStocks()"
        :class="['px-4 py-2 text-xs font-bold rounded-xl transition-all',
          activeSubTab === 'inventory' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-500 hover:bg-slate-100']"
      >
        <i class="pi pi-box mr-2"></i>Stock de Sucursal
      </button>
      <button
        @click="activeSubTab = 'kardex'; loadKardex()"
        :class="['px-4 py-2 text-xs font-bold rounded-xl transition-all',
          activeSubTab === 'kardex' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-500 hover:bg-slate-100']"
      >
        <i class="pi pi-history mr-2"></i>Kardex de Movimientos
      </button>
    </div>

    <!-- Sub-tab 1: Inventory Stock -->
    <div v-if="activeSubTab === 'inventory'" class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
      <div class="flex justify-between items-center mb-6">
        <h3 class="text-base font-extrabold text-slate-800">Inventario en Sucursal</h3>
        <button
          @click="loadStocks"
          class="p-2 text-xs font-semibold text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-xl transition-all flex items-center gap-1.5"
        >
          <i class="pi pi-refresh"></i>Actualizar
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Código</th>
              <th class="py-3 px-4">Producto</th>
              <th class="py-3 px-4 text-right">Precio Compra</th>
              <th class="py-3 px-4 text-right">Precio Venta</th>
              <th class="py-3 px-4 text-center">Stock Mín.</th>
              <th class="py-3 px-4 text-center">Stock Actual</th>
              <th class="py-3 px-4 text-center">Estado</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="item in stocks" :key="item.product_id" class="hover:bg-slate-50/50">
              <td class="py-3.5 px-4 font-mono text-slate-400 text-[10px]">{{ item.product_codigo || '-' }}</td>
              <td class="py-3.5 px-4 font-bold text-slate-900">{{ item.product_nombre }}</td>
              <td class="py-3.5 px-4 text-right font-mono text-slate-500">S/. {{ item.precio_compra.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-right font-mono font-semibold text-slate-900">S/. {{ item.precio_venta.toFixed(2) }}</td>
              <td class="py-3.5 px-4 text-center font-mono text-slate-500">{{ item.stock_minimo }}</td>
              <td class="py-3.5 px-4 text-center">
                <span
                  :class="['px-2.5 py-1 rounded-lg font-mono font-bold text-xs',
                    item.stock_actual <= item.stock_minimo
                      ? 'bg-rose-50 text-rose-600 border border-rose-100'
                      : 'bg-emerald-50 text-emerald-600 border border-emerald-100']"
                >
                  {{ item.stock_actual }}
                </span>
              </td>
              <td class="py-3.5 px-4 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-50 text-emerald-700 border border-emerald-200">
                  {{ item.estado }}
                </span>
              </td>
            </tr>
            <tr v-if="stocks.length === 0">
              <td colspan="7" class="text-center py-8 text-slate-400">No hay productos en inventario</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Sub-tab 2: Kardex Log -->
    <div v-else class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-6">
      <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 border-b border-slate-100 pb-4">
        <h3 class="text-base font-extrabold text-slate-800">Historial de Movimientos de Inventario</h3>
        
        <!-- Product Filter -->
        <div class="flex items-center gap-3">
          <label class="text-xs font-bold text-slate-600 whitespace-nowrap">Filtrar por Producto:</label>
          <select
            v-model="selectedProductId"
            @change="loadKardex"
            class="px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:outline-none transition-all min-w-[200px]"
          >
            <option value="">-- Todos los Productos --</option>
            <option v-for="item in stocks" :key="item.product_id" :value="item.product_id">
              {{ item.product_nombre }}
            </option>
          </select>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Fecha</th>
              <th class="py-3 px-4">Producto</th>
              <th class="py-3 px-4 text-center">Tipo</th>
              <th class="py-3 px-4">Referencia</th>
              <th class="py-3 px-4 text-center">Cantidad</th>
              <th class="py-3 px-4 text-center">Stock Ant.</th>
              <th class="py-3 px-4 text-center">Stock Nuevo</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="m in kardex" :key="m.id" class="hover:bg-slate-50/50">
              <td class="py-3 px-4 text-slate-500 font-mono text-[10px]">{{ formatDate(m.created_at) }}</td>
              <td class="py-3 px-4 font-bold text-slate-900">
                {{ m.product_nombre }}
                <span v-if="m.product_codigo" class="text-[10px] font-mono text-slate-400 block mt-0.5">Cód: {{ m.product_codigo }}</span>
              </td>
              <td class="py-3 px-4 text-center">
                <span
                  :class="['px-2 py-0.5 rounded-full text-[10px] font-bold border',
                    m.tipo_movimiento === 'INGRESO'
                      ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                      : m.tipo_movimiento === 'VENTA'
                        ? 'bg-sky-50 text-sky-700 border-sky-200'
                        : 'bg-amber-50 text-amber-700 border-amber-200']"
                >
                  {{ m.tipo_movimiento }}
                </span>
              </td>
              <td class="py-3 px-4 font-mono text-[10px] text-slate-500">{{ m.referencia || '-' }}</td>
              <td class="py-3 px-4 text-center font-bold font-mono">
                {{ m.tipo_movimiento === 'VENTA' ? '-' : '+' }}{{ m.cantidad }}
              </td>
              <td class="py-3 px-4 text-center font-mono text-slate-500">{{ m.stock_anterior }}</td>
              <td class="py-3 px-4 text-center font-mono font-bold text-slate-900">{{ m.stock_nuevo }}</td>
            </tr>
            <tr v-if="kardex.length === 0">
              <td colspan="7" class="text-center py-8 text-slate-400">No hay movimientos registrados</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'

const activeSubTab = ref('inventory')
const stocks = ref<any[]>([])
const kardex = ref<any[]>([])
const selectedProductId = ref('')

onMounted(() => {
  loadStocks()
})

async function loadStocks() {
  try {
    const res = await axios.get('/stocks')
    if (res.data.success) {
      stocks.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading stocks', err)
  }
}

async function loadKardex() {
  try {
    let url = '/kardex'
    if (selectedProductId.value) {
      url += `?product_id=${selectedProductId.value}`
    }
    const res = await axios.get(url)
    if (res.data.success) {
      kardex.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading kardex', err)
  }
}

function formatDate(dStr: string) {
  if (!dStr) return '-'
  const d = new Date(dStr)
  return d.toLocaleString('es-PE', { timeZone: 'America/Lima' })
}
</script>
