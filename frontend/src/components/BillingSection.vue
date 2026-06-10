<template>
  <div class="max-w-4xl mx-auto space-y-6">
    <div class="bg-white p-6 rounded-xl border border-slate-200 shadow-sm">
      <div class="flex items-center gap-4 mb-6">
        <div class="p-3 bg-indigo-50 rounded-lg text-indigo-600">
          <i class="pi pi-file-export text-2xl"></i>
        </div>
        <div>
          <h3 class="text-lg font-extrabold text-slate-800">Configuración de Facturación Electrónica</h3>
          <p class="text-xs text-slate-500 font-medium">Configura tu conexión con FacturaAPI para emitir comprobantes válidos ante SUNAT.</p>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <i class="pi pi-spin pi-spinner text-3xl text-indigo-500"></i>
      </div>

      <form v-else @submit.prevent="saveConfig" class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <h4 class="text-xs font-bold text-slate-400 uppercase tracking-wider">Credenciales de API</h4>
            
            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">URL del API</label>
              <input 
                v-model="config.api_url"
                type="url" 
                placeholder="https://api.factura.api/v1"
                class="w-full px-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:ring-indigo-500 outline-none"
                required
              />
              <p class="text-[10px] text-slate-400 mt-1">Endpoint base de tu microservicio de facturación.</p>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">API Key / Token</label>
              <div class="relative">
                <input 
                  :type="showKey ? 'text' : 'password'"
                  v-model="config.api_key"
                  placeholder="Tu clave secreta..."
                  class="w-full pl-4 pr-10 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:ring-indigo-500 outline-none font-mono"
                  required
                />
                <button 
                  type="button"
                  @click="showKey = !showKey"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-indigo-600"
                >
                  <i :class="showKey ? 'pi pi-eye-slash' : 'pi pi-eye'"></i>
                </button>
              </div>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">Tenant UUID</label>
              <input 
                v-model="config.tenant_uuid"
                type="text" 
                placeholder="00000000-0000-0000-0000-000000000000"
                class="w-full px-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:ring-indigo-500 outline-none font-mono"
                required
              />
              <p class="text-[10px] text-slate-400 mt-1">Identificador único de tu empresa en FacturaAPI.</p>
            </div>
          </div>

          <div class="space-y-4">
            <h4 class="text-xs font-bold text-slate-400 uppercase tracking-wider">Entorno y Estado</h4>

            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">Modo de Operación</label>
              <div class="grid grid-cols-2 gap-2">
                <button 
                  type="button"
                  @click="config.modo = 'dev'"
                  :class="[
                    'py-2 px-4 rounded-lg text-xs font-bold border transition-all',
                    config.modo === 'dev' ? 'bg-amber-50 border-amber-200 text-amber-700' : 'bg-white border-slate-200 text-slate-500 hover:bg-slate-50'
                  ]"
                >
                  <i class="pi pi-test mr-2"></i> Desarrollo / Pruebas
                </button>
                <button 
                  type="button"
                  @click="config.modo = 'prod'"
                  :class="[
                    'py-2 px-4 rounded-lg text-xs font-bold border transition-all',
                    config.modo === 'prod' ? 'bg-emerald-50 border-emerald-200 text-emerald-700 shadow-sm' : 'bg-white border-slate-200 text-slate-500 hover:bg-slate-50'
                  ]"
                >
                  <i class="pi pi-check-circle mr-2"></i> Producción
                </button>
              </div>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">Estado del Servicio</label>
              <select 
                v-model="config.estado"
                class="w-full px-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:ring-indigo-500 outline-none"
              >
                <option value="active">Activo</option>
                <option value="inactive">Inactivo (No emitir)</option>
              </select>
            </div>

            <div class="p-4 bg-slate-50 rounded-xl border border-slate-200">
              <div class="flex items-start gap-3">
                <i class="pi pi-info-circle text-indigo-500 mt-0.5"></i>
                <div class="text-[11px] text-slate-600 leading-relaxed">
                  <strong>Importante:</strong> Asegúrate de que tu empresa tenga el certificado digital cargado en FacturaAPI antes de pasar a modo producción.
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="pt-6 border-t border-slate-100 flex items-center justify-between">
          <div class="flex items-center gap-2 text-xs font-bold" :class="config.estado === 'active' ? 'text-emerald-600' : 'text-slate-400'">
            <span class="w-2 h-2 rounded-full" :class="config.estado === 'active' ? 'bg-emerald-500 animate-pulse' : 'bg-slate-300'"></span>
            {{ config.estado === 'active' ? 'Servicio Configurado' : 'Servicio Deshabilitado' }}
          </div>
          
          <button 
            type="submit"
            :disabled="saving"
            class="px-8 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-bold shadow-lg shadow-indigo-200 transition-all flex items-center gap-2 disabled:opacity-50"
          >
            <i v-if="saving" class="pi pi-spin pi-spinner"></i>
            <span>{{ saving ? 'Guardando...' : 'Guardar Configuración' }}</span>
          </button>
        </div>
      </form>
    </div>

    <!-- Help Card -->
    <div class="bg-indigo-900 text-indigo-100 p-6 rounded-xl border border-indigo-800 shadow-sm flex flex-col md:flex-row items-center gap-6">
      <div class="flex-1">
        <h4 class="font-bold mb-1">¿Necesitas ayuda con la integración?</h4>
        <p class="text-xs text-indigo-300">Consulta la documentación de FacturaAPI para obtener tus credenciales y configurar los certificados digitales.</p>
      </div>
      <a href="#" class="px-4 py-2 bg-indigo-800 hover:bg-indigo-700 rounded-lg text-xs font-bold transition-all border border-indigo-700">
        Ver Documentación <i class="pi pi-external-link ml-1"></i>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'

const loading = ref(true)
const saving = ref(false)
const showKey = ref(false)

const config = reactive({
  api_url: '',
  api_key: '',
  tenant_uuid: '',
  modo: 'dev',
  estado: 'active'
})

onMounted(async () => {
  await loadConfig()
})

async function loadConfig() {
  loading.value = true
  try {
    const res = await axios.get('/billing/config')
    if (res.data.success && res.data.data) {
      Object.assign(config, res.data.data)
    }
  } catch (err) {
    console.error('Error loading billing config', err)
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    const res = await axios.post('/billing/config', config)
    if (res.data.success) {
      alert('Configuración guardada correctamente')
    }
  } catch (err) {
    alert('Error al guardar la configuración')
  } finally {
    saving.value = false
  }
}
</script>
