<template>
  <div class="flex flex-col h-full space-y-6">
    <!-- Header Summary -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div class="bg-white p-6 rounded-[2.5rem] border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 bg-emerald-50 text-emerald-600 rounded-2xl flex items-center justify-center">
          <i class="pi pi-wallet text-xl"></i>
        </div>
        <div>
          <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Saldo en Caja</p>
          <p class="text-xl font-black text-slate-900">S/. {{ currentBalance.toFixed(2) }}</p>
        </div>
      </div>
      
      <div class="bg-white p-6 rounded-[2.5rem] border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 bg-indigo-50 text-indigo-600 rounded-2xl flex items-center justify-center">
          <i class="pi pi-shopping-cart text-xl"></i>
        </div>
        <div>
          <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Ventas del Turno</p>
          <p class="text-xl font-black text-slate-900">S/. {{ (activeSession?.total_ventas_efe || 0).toFixed(2) }}</p>
        </div>
      </div>

      <div class="bg-white p-6 rounded-[2.5rem] border border-slate-200 shadow-sm flex items-center gap-4">
        <div :class="['w-12 h-12 rounded-2xl flex items-center justify-center', activeSession ? 'bg-emerald-50 text-emerald-600' : 'bg-red-50 text-red-600']">
          <i :class="['pi', activeSession ? 'pi-lock-open' : 'pi-lock']"></i>
        </div>
        <div>
          <p class="text-[10px] text-slate-400 font-black uppercase tracking-widest">Estado</p>
          <p class="text-sm font-black text-slate-900 uppercase">{{ activeSession ? 'Caja Abierta' : 'Caja Cerrada' }}</p>
        </div>
      </div>
    </div>

    <!-- Actions Bar -->
    <div class="flex flex-wrap gap-4">
      <button 
        v-if="!activeSession"
        @click="showOpenModal = true"
        class="px-8 py-3 bg-indigo-600 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-indigo-200 hover:bg-indigo-500 transition-all flex items-center gap-2"
      >
        <i class="pi pi-plus-circle"></i> Apertura de Caja
      </button>

      <template v-else>
        <button 
          @click="showCloseModal = true"
          class="px-8 py-3 bg-slate-900 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-slate-200 hover:bg-slate-800 transition-all flex items-center gap-2"
        >
          <i class="pi pi-lock"></i> Cerrar Caja (Arqueo)
        </button>

        <button 
          @click="openMovementModal('INGRESO')"
          class="px-8 py-3 bg-white border border-slate-200 text-emerald-600 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-emerald-50 transition-all flex items-center gap-2"
        >
          <i class="pi pi-arrow-up"></i> Ingreso de Efectivo
        </button>

        <button 
          @click="openMovementModal('EGRESO')"
          class="px-8 py-3 bg-white border border-slate-200 text-red-600 rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-red-50 transition-all flex items-center gap-2"
        >
          <i class="pi pi-arrow-down"></i> Salida / Gasto
        </button>
      </template>
    </div>

    <!-- History Table -->
    <div class="bg-white rounded-[2.5rem] border border-slate-200 shadow-sm overflow-hidden flex-1 flex flex-col">
      <div class="p-6 border-b border-slate-100 flex items-center justify-between">
        <h3 class="text-sm font-black text-slate-900 uppercase tracking-widest">Historial de Sesiones</h3>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50">
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Apertura</th>
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Cierre</th>
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Inicial</th>
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Ventas Efe.</th>
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Final Real</th>
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Diferencia</th>
              <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">Estado</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="sess in sessions" :key="sess.id" class="hover:bg-slate-50 transition-colors">
              <td class="p-4">
                <p class="text-xs font-bold text-slate-900">{{ formatDate(sess.opened_at) }}</p>
              </td>
              <td class="p-4">
                <p class="text-xs font-bold text-slate-500">{{ sess.closed_at ? formatDate(sess.closed_at) : '-' }}</p>
              </td>
              <td class="p-4 text-xs font-bold text-slate-900">S/. {{ sess.saldo_inicial.toFixed(2) }}</td>
              <td class="p-4 text-xs font-bold text-emerald-600">S/. {{ sess.total_ventas_efe.toFixed(2) }}</td>
              <td class="p-4 text-xs font-bold text-slate-900">S/. {{ sess.saldo_final_real?.toFixed(2) || '-' }}</td>
              <td class="p-4">
                <span v-if="sess.estado === 'CLOSED'" :class="['text-xs font-black', getDiff(sess) >= 0 ? 'text-emerald-600' : 'text-red-600']">
                  S/. {{ getDiff(sess).toFixed(2) }}
                </span>
                <span v-else class="text-xs text-slate-300">-</span>
              </td>
              <td class="p-4">
                <span :class="['text-[9px] font-black px-2 py-1 rounded-lg uppercase tracking-widest', sess.estado === 'OPEN' ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-400']">
                  {{ sess.estado }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="sessions.length === 0" class="p-12 text-center text-slate-300">
          <i class="pi pi-inbox text-4xl mb-2"></i>
          <p class="text-xs font-bold uppercase">No hay registros de caja</p>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <!-- Open Modal -->
    <div v-if="showOpenModal" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 z-[60] animate-fade-in">
      <div class="bg-white rounded-[2.5rem] w-full max-w-md p-8 shadow-2xl">
        <h3 class="text-xl font-black text-slate-900 mb-2">Apertura de Caja</h3>
        <p class="text-xs text-slate-500 font-bold mb-6">Ingresa el monto inicial con el que inicias el turno.</p>
        
        <div class="space-y-4">
          <div>
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Monto Inicial (S/.)</label>
            <input 
              v-model.number="openForm.saldo_inicial"
              type="number" 
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-lg font-black outline-none focus:ring-4 focus:ring-indigo-500/10 transition-all"
            />
          </div>
          <div class="flex gap-3 pt-4">
            <button @click="showOpenModal = false" class="flex-1 py-3 text-[10px] font-black uppercase text-slate-400 hover:text-slate-600">Cancelar</button>
            <button @click="handleOpen" class="flex-[2] py-4 bg-indigo-600 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-indigo-200">Confirmar Apertura</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Close Modal -->
    <div v-if="showCloseModal" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 z-[60] animate-fade-in">
      <div class="bg-white rounded-[2.5rem] w-full max-w-md p-8 shadow-2xl">
        <h3 class="text-xl font-black text-slate-900 mb-2">Cierre de Caja</h3>
        <p class="text-xs text-slate-500 font-bold mb-6">Realiza el arqueo contando el efectivo físico en caja.</p>
        
        <div class="space-y-4">
          <div class="bg-slate-50 p-4 rounded-2xl space-y-2 mb-4">
            <div class="flex justify-between text-xs font-bold text-slate-600">
              <span>Monto Esperado:</span>
              <span>S/. {{ currentBalance.toFixed(2) }}</span>
            </div>
            <div class="text-[9px] text-slate-400">
              (Inicial + Ventas Efe + Ingresos - Egresos)
            </div>
          </div>

          <div>
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Monto Físico Contado (S/.)</label>
            <input 
              v-model.number="closeForm.saldo_final_real"
              type="number" 
              class="w-full px-4 py-3 bg-white border border-slate-200 rounded-2xl text-lg font-black outline-none focus:ring-4 focus:ring-indigo-500/10 transition-all"
            />
          </div>

          <div>
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Notas / Observaciones</label>
            <textarea 
              v-model="closeForm.notas"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold outline-none focus:ring-4 focus:ring-indigo-500/10 transition-all"
              rows="3"
            ></textarea>
          </div>

          <div class="flex gap-3 pt-4">
            <button @click="showCloseModal = false" class="flex-1 py-3 text-[10px] font-black uppercase text-slate-400 hover:text-slate-600">Cancelar</button>
            <button @click="handleClose" class="flex-[2] py-4 bg-slate-900 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-slate-200">Cerrar Turno</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Movement Modal -->
    <div v-if="showMovementModal" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 z-[60] animate-fade-in">
      <div class="bg-white rounded-[2.5rem] w-full max-w-md p-8 shadow-2xl">
        <h3 class="text-xl font-black text-slate-900 mb-2">{{ movementForm.tipo === 'INGRESO' ? 'Nuevo Ingreso' : 'Nueva Salida / Gasto' }}</h3>
        
        <div class="space-y-4 mt-6">
          <div>
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Monto (S/.)</label>
            <input 
              v-model.number="movementForm.monto"
              type="number" 
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-lg font-black outline-none focus:ring-4 focus:ring-indigo-500/10 transition-all"
            />
          </div>

          <div>
            <label class="block text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Motivo / Descripción</label>
            <input 
              v-model="movementForm.motivo"
              type="text" 
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold outline-none focus:ring-4 focus:ring-indigo-500/10 transition-all"
            />
          </div>

          <div class="flex gap-3 pt-4">
            <button @click="showMovementModal = false" class="flex-1 py-3 text-[10px] font-black uppercase text-slate-400 hover:text-slate-600">Cancelar</button>
            <button @click="handleMovement" class="flex-[2] py-4 bg-indigo-600 text-white rounded-2xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-indigo-200">Registrar</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import axios from 'axios'

const activeSession = ref<any>(null)
const sessions = ref<any[]>([])
const showOpenModal = ref(false)
const showCloseModal = ref(false)
const showMovementModal = ref(false)

const openForm = reactive({ saldo_inicial: 0 })
const closeForm = reactive({ saldo_final_real: 0, notas: '' })
const movementForm = reactive({ tipo: 'INGRESO', monto: 0, motivo: '' })

const currentBalance = computed(() => {
  if (!activeSession.value) return 0
  const s = activeSession.value
  return s.saldo_inicial + s.total_ventas_efe + s.total_ingresos - s.total_egresos
})

onMounted(async () => {
  await loadActiveSession()
  await loadHistory()
})

async function loadActiveSession() {
  try {
    const res = await axios.get('/cash/active')
    if (res.data.success) activeSession.value = res.data.data
  } catch (err) {
    console.error('Error loading active session', err)
  }
}

async function loadHistory() {
  try {
    const res = await axios.get('/cash/sessions')
    if (res.data.success) sessions.value = res.data.data
  } catch (err) {
    console.error('Error loading history', err)
  }
}

async function handleOpen() {
  try {
    const res = await axios.post('/cash/open', openForm)
    if (res.data.success) {
      activeSession.value = res.data.data
      showOpenModal.value = false
      await loadHistory()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al abrir caja')
  }
}

async function handleClose() {
  try {
    const res = await axios.post('/cash/close', closeForm)
    if (res.data.success) {
      activeSession.value = null
      showCloseModal.value = false
      await loadHistory()
      alert('Caja cerrada con éxito')
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al cerrar caja')
  }
}

function openMovementModal(tipo: string) {
  movementForm.tipo = tipo
  movementForm.monto = 0
  movementForm.motivo = ''
  showMovementModal.value = true
}

async function handleMovement() {
  try {
    const res = await axios.post('/cash/movements', movementForm)
    if (res.data.success) {
      showMovementModal.value = false
      await loadActiveSession()
      await loadHistory()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al registrar movimiento')
  }
}

function getDiff(sess: any) {
  const expected = sess.saldo_inicial + sess.total_ventas_efe + sess.total_ingresos - sess.total_egresos
  return sess.saldo_final_real - expected
}

function formatDate(dStr: string) {
  if (!dStr) return '-'
  const d = new Date(dStr)
  return d.toLocaleString('es-PE', { 
    day: '2-digit', 
    month: 'short', 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}
</script>
