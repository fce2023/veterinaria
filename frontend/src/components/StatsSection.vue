<template>
  <div class="space-y-8 animate-fade-in">
    <!-- Top Statistics Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <!-- Sales Today Card -->
      <div class="bg-gradient-to-tr from-sky-600 to-indigo-600 p-6 rounded-2xl text-white shadow-md flex items-center justify-between border border-sky-500/20">
        <div>
          <p class="text-xs font-bold text-sky-100 uppercase tracking-wider">Ventas de Hoy</p>
          <h3 class="text-3xl font-black mt-1 font-mono">S/. {{ stats.sales_today.toFixed(2) }}</h3>
          <p class="text-[10px] text-sky-200 mt-1"><i class="pi pi-calendar mr-1"></i>Día actual</p>
        </div>
        <div class="w-12 h-12 bg-white/10 text-white rounded-xl flex items-center justify-center text-xl">
          <i class="pi pi-bolt"></i>
        </div>
      </div>

      <!-- Sales Month Card -->
      <div class="bg-gradient-to-tr from-emerald-600 to-teal-600 p-6 rounded-2xl text-white shadow-md flex items-center justify-between border border-emerald-500/20">
        <div>
          <p class="text-xs font-bold text-emerald-100 uppercase tracking-wider">Ventas del Mes</p>
          <h3 class="text-3xl font-black mt-1 font-mono">S/. {{ stats.sales_month.toFixed(2) }}</h3>
          <p class="text-[10px] text-emerald-200 mt-1"><i class="pi pi-chart-line mr-1"></i>Mes actual</p>
        </div>
        <div class="w-12 h-12 bg-white/10 text-white rounded-xl flex items-center justify-center text-xl">
          <i class="pi pi-percentage"></i>
        </div>
      </div>

      <!-- Purchases Month Card -->
      <div class="bg-gradient-to-tr from-rose-600 to-orange-600 p-6 rounded-2xl text-white shadow-md flex items-center justify-between border border-rose-500/20">
        <div>
          <p class="text-xs font-bold text-rose-100 uppercase tracking-wider">Compras del Mes</p>
          <h3 class="text-3xl font-black mt-1 font-mono">S/. {{ stats.purchases_month.toFixed(2) }}</h3>
          <p class="text-[10px] text-rose-200 mt-1"><i class="pi pi-shopping-bag mr-1"></i>Egresos mes</p>
        </div>
        <div class="w-12 h-12 bg-white/10 text-white rounded-xl flex items-center justify-center text-xl">
          <i class="pi pi-wallet"></i>
        </div>
      </div>

      <!-- Low Stock Warning Card -->
      <div :class="['p-6 rounded-2xl shadow-md flex items-center justify-between border transition-all',
        stats.low_stock_count > 0 ? 'bg-amber-50 border-amber-200 text-amber-900' : 'bg-white border-slate-200 text-slate-800']">
        <div>
          <p class="text-xs font-bold uppercase tracking-wider text-slate-500">Stock Crítico</p>
          <h3 class="text-3xl font-black mt-1 font-mono" :class="stats.low_stock_count > 0 ? 'text-amber-700' : 'text-slate-800'">
            {{ stats.low_stock_count }}
          </h3>
          <p class="text-[10px] mt-1" :class="stats.low_stock_count > 0 ? 'text-amber-600 font-semibold' : 'text-slate-400'">
            <i class="pi pi-exclamation-triangle mr-1"></i>{{ stats.low_stock_count > 0 ? '¡Productos bajo mínimo!' : 'Inventario estable' }}
          </p>
        </div>
        <div :class="['w-12 h-12 rounded-xl flex items-center justify-center text-xl',
          stats.low_stock_count > 0 ? 'bg-amber-100 text-amber-600' : 'bg-slate-100 text-slate-500']">
          <i class="pi pi-box"></i>
        </div>
      </div>
    </div>

    <!-- Main Content Area: Left (Top products), Right (Quick actions) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Top Selling Products -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm lg:col-span-2 flex flex-col justify-between">
        <div>
          <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-6">
            Productos Más Vendidos
          </h3>

          <div class="space-y-5">
            <div v-for="(prod, idx) in stats.top_products" :key="idx" class="space-y-2">
              <div class="flex justify-between items-center text-xs">
                <div>
                  <span class="font-bold text-slate-800">{{ prod.nombre }}</span>
                  <span v-if="prod.codigo" class="text-[10px] font-mono text-slate-400 ml-2">({{ prod.codigo }})</span>
                </div>
                <span class="font-mono font-bold text-slate-900">{{ prod.cantidad }} und.</span>
              </div>
              <div class="w-full bg-slate-100 h-2 rounded-full overflow-hidden">
                <div
                  class="bg-gradient-to-r from-sky-500 to-indigo-500 h-full rounded-full transition-all duration-500"
                  :style="{ width: `${calculatePercentage(prod.cantidad)}%` }"
                ></div>
              </div>
            </div>
            
            <div v-if="stats.top_products.length === 0" class="text-center py-12 text-slate-400">
              <i class="pi pi-chart-bar text-3xl block mb-2 text-slate-300"></i>
              No hay datos de ventas para este periodo.
            </div>
          </div>
        </div>

        <div class="text-[10px] text-slate-400 mt-6 text-right">
          * Basado en la cantidad total de unidades vendidas en la sucursal actual.
        </div>
      </div>

      <!-- Quick Action Shortcuts -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4 flex flex-col justify-between h-fit">
        <div>
          <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4">
            Accesos Rápidos
          </h3>
          <p class="text-xs text-slate-500 mb-6">Operaciones comunes disponibles en el sistema:</p>

          <div class="grid grid-cols-1 gap-3">
            <button
              @click="$emit('navigate', 'sales')"
              class="w-full p-3 bg-sky-50 hover:bg-sky-100 border border-sky-200 text-sky-850 rounded-xl text-left text-xs font-bold transition-all flex items-center justify-between group"
            >
              <span class="flex items-center gap-3"><i class="pi pi-desktop text-sky-600 text-sm"></i>Punto de Venta (POS)</span>
              <i class="pi pi-arrow-right text-[10px] group-hover:translate-x-1 transition-all"></i>
            </button>
            <button
              @click="$emit('navigate', 'purchases')"
              class="w-full p-3 bg-emerald-55/10 hover:bg-emerald-50 border border-emerald-200 text-emerald-850 rounded-xl text-left text-xs font-bold transition-all flex items-center justify-between group"
            >
              <span class="flex items-center gap-3"><i class="pi pi-shopping-cart text-emerald-600 text-sm"></i>Registrar Compra</span>
              <i class="pi pi-arrow-right text-[10px] group-hover:translate-x-1 transition-all"></i>
            </button>
            <button
              @click="$emit('navigate', 'kardex')"
              class="w-full p-3 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 text-indigo-850 rounded-xl text-left text-xs font-bold transition-all flex items-center justify-between group"
            >
              <span class="flex items-center gap-3"><i class="pi pi-history text-indigo-650 text-sm"></i>Ver Kardex / Stock</span>
              <i class="pi pi-arrow-right text-[10px] group-hover:translate-x-1 transition-all"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'

const stats = reactive({
  sales_today: 0,
  sales_month: 0,
  purchases_month: 0,
  low_stock_count: 0,
  top_products: [] as any[]
})

onMounted(() => {
  loadStats()
})

async function loadStats() {
  try {
    const res = await axios.get('/dashboard/stats')
    if (res.data.success) {
      Object.assign(stats, res.data.data)
      if (!stats.top_products) stats.top_products = []
    }
  } catch (err) {
    console.error('Error loading dashboard stats', err)
  }
}

function calculatePercentage(qty: number) {
  if (stats.top_products.length === 0) return 0
  const maxQty = stats.top_products[0].cantidad
  if (maxQty === 0) return 0
  return (qty / maxQty) * 100
}
</script>
