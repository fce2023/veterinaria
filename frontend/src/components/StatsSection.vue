<template>
  <div class="stats-container animate-fade-in">
    <!-- Top Statistics Strip -->
    <div class="stats-strip">
      <div class="stat-card">
        <div class="stat-icon green"><i class="ti ti-trending-up"></i></div>
        <div class="stat-info">
          <div class="stat-label">Ventas Hoy</div>
          <div class="stat-value">S/. {{ stats.sales_today.toFixed(2) }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon accent"><i class="ti ti-chart-dots"></i></div>
        <div class="stat-info">
          <div class="stat-label">Ventas Mes</div>
          <div class="stat-value">S/. {{ stats.sales_month.toFixed(2) }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon red"><i class="ti ti-shopping-cart-x"></i></div>
        <div class="stat-info">
          <div class="stat-label">Compras Mes</div>
          <div class="stat-value">S/. {{ stats.purchases_month.toFixed(2) }}</div>
        </div>
      </div>
      <div class="stat-card" :class="{ warning: stats.low_stock_count > 0 }">
        <div class="stat-icon amber"><i class="ti ti-alert-triangle"></i></div>
        <div class="stat-info">
          <div class="stat-label">Stock Crítico</div>
          <div class="stat-value">{{ stats.low_stock_count }} prod.</div>
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="stats-grid">
      <!-- Top Selling Products -->
      <div class="glass-card main-stats">
        <div class="card-head">
          <h3 class="card-title">Productos más vendidos</h3>
          <p class="card-subtitle">Basado en unidades totales</p>
        </div>

        <div class="top-list">
          <div v-for="(prod, idx) in stats.top_products" :key="idx" class="top-item">
            <div class="top-info">
              <span class="top-name">{{ prod.nombre }}</span>
              <span class="top-qty">{{ prod.cantidad }} und.</span>
            </div>
            <div class="progress-bg">
              <div class="progress-bar" :style="{ width: `${calculatePercentage(prod.cantidad)}%` }"></div>
            </div>
          </div>
          
          <div v-if="stats.top_products.length === 0" class="empty-list">
            <i class="ti ti-chart-bar"></i>
            <p>No hay datos disponibles</p>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="glass-card side-stats">
        <div class="card-head">
          <h3 class="card-title">Accesos Rápidos</h3>
        </div>
        <div class="actions-list">
          <button @click="$emit('navigate', 'sales')" class="action-btn">
            <div class="action-icon"><i class="ti ti-device-desktop"></i></div>
            <span>Punto de Venta (POS)</span>
          </button>
          <button @click="$emit('navigate', 'purchases')" class="action-btn">
            <div class="action-icon"><i class="ti ti-shopping-cart"></i></div>
            <span>Registrar Compra</span>
          </button>
          <button @click="$emit('navigate', 'kardex')" class="action-btn">
            <div class="action-icon"><i class="ti ti-arrows-exchange"></i></div>
            <span>Movimientos / Stock</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
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

<style scoped>
.stats-container { display: flex; flex-direction: column; gap: 24px; }

.stats-strip { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; }
.stat-card {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 20px; display: flex; align-items: center; gap: 16px; transition: all var(--transition);
}
.stat-card:hover { border-color: var(--accent); transform: translateY(-2px); }
.stat-card.warning { border-color: var(--amber); background: var(--amber-dim); }

.stat-icon {
  width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 20px;
}
.stat-icon.green { background: var(--green-dim); color: var(--green); }
.stat-icon.accent { background: var(--accent-dim); color: var(--accent); }
.stat-icon.red { background: var(--red-dim); color: var(--red); }
.stat-icon.amber { background: var(--amber-dim); color: var(--amber); }

.stat-label { font-size: 11px; font-weight: 700; color: var(--text3); text-transform: uppercase; letter-spacing: 0.05em; }
.stat-value { font-size: 18px; font-weight: 900; color: var(--text); margin-top: 2px; }

.stats-grid { display: grid; grid-template-columns: 2fr 1fr; gap: 24px; }
@media (max-width: 900px) { .stats-grid { grid-template-columns: 1fr; } }

.glass-card { background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
.card-head { padding: 24px; border-bottom: 1px solid var(--border); }
.card-title { font-size: 16px; font-weight: 800; color: var(--text); }
.card-subtitle { font-size: 11px; color: var(--text3); margin-top: 4px; }

.top-list { padding: 24px; display: flex; flex-direction: column; gap: 20px; }
.top-item { display: flex; flex-direction: column; gap: 8px; }
.top-info { display: flex; justify-content: space-between; font-size: 13px; font-weight: 700; }
.top-qty { color: var(--accent); }
.progress-bg { width: 100%; height: 6px; background: var(--surface2); border-radius: 10px; overflow: hidden; }
.progress-bar { height: 100%; background: var(--accent); border-radius: 10px; transition: width 0.8s ease-out; }

.actions-list { padding: 24px; display: flex; flex-direction: column; gap: 12px; }
.action-btn {
  width: 100%; padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border);
  background: var(--surface2); display: flex; align-items: center; gap: 12px;
  cursor: pointer; transition: all var(--transition); font-family: inherit; font-size: 13px; font-weight: 600; color: var(--text);
}
.action-btn:hover { border-color: var(--accent); background: var(--surface); color: var(--accent); }
.action-icon { font-size: 18px; color: var(--text3); }
.action-btn:hover .action-icon { color: var(--accent); }

.empty-list { text-align: center; padding: 40px 0; color: var(--text3); }
.empty-list i { font-size: 32px; margin-bottom: 10px; opacity: 0.5; }
</style>
