<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
    <!-- Form -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4 h-fit">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 flex items-center gap-2">
        <i class="pi pi-plus text-indigo-500"></i>
        <span>{{ editId ? 'Editar Proveedor' : 'Nuevo Proveedor' }}</span>
      </h3>
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div class="flex items-center gap-2 mb-1 px-1">
          <input 
            id="genericSupplier"
            v-model="supplierIsGeneric"
            type="checkbox"
            class="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
          />
          <label for="genericSupplier" class="text-[10px] font-bold text-slate-600 uppercase tracking-wider cursor-pointer">
            Sin RUC / Proveedor Genérico
          </label>
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">RUC (11 dígitos)</label>
          <div class="relative">
            <input
              v-model="form.ruc"
              type="text"
              required
              maxlength="11"
              :disabled="supplierIsGeneric"
              class="block w-full pl-3 pr-10 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:outline-none transition-all disabled:opacity-60"
              placeholder="Ej. 20987654321"
            />
            <div class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
              <i v-if="queryLoading" class="pi pi-spin pi-spinner text-indigo-600"></i>
              <i v-else class="pi pi-search text-xs"></i>
            </div>
          </div>
          <span v-if="!supplierIsGeneric" class="text-[9px] text-slate-400 font-bold uppercase mt-1 block px-0.5">Se consultará SUNAT al completar 11 dígitos</span>
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Razón Social</label>
          <input
            v-model="form.razon_social"
            type="text"
            required
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:outline-none transition-all"
            placeholder="Ej. Distribuidora Vet S.A."
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Dirección</label>
          <input
            v-model="form.direccion"
            type="text"
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:outline-none transition-all"
            placeholder="Av. Industrial 500, Lima"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Teléfono</label>
          <input
            v-model="form.telefono"
            type="text"
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 focus:outline-none transition-all"
            placeholder="Ej. 987654321"
          />
        </div>
        <div class="flex gap-2 pt-2">
          <button
            type="submit"
            class="flex-1 py-2.5 bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center justify-center gap-2"
          >
            <i :class="editId ? 'pi pi-check' : 'pi pi-plus'"></i>
            <span>{{ editId ? 'Guardar Cambios' : 'Registrar Proveedor' }}</span>
          </button>
          <button
            v-if="editId"
            type="button"
            @click="cancelEdit"
            class="px-4 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl text-xs font-bold transition-all"
          >
            Cancelar
          </button>
        </div>
      </form>
    </div>

    <!-- List -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm lg:col-span-2 flex flex-col">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4 flex items-center gap-2">
        <i class="pi pi-truck text-indigo-500"></i>
        <span>Directorio de Proveedores</span>
      </h3>
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">RUC / Razón Social</th>
              <th class="py-3 px-4">Contacto</th>
              <th class="py-3 px-4 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="supplier in suppliers" :key="supplier.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="py-3.5 px-4">
                <div class="font-bold text-slate-900">{{ supplier.razon_social }}</div>
                <div class="text-[10px] font-mono text-slate-400 mt-0.5">RUC: {{ supplier.ruc }}</div>
              </td>
              <td class="py-3.5 px-4 text-slate-500">
                <div v-if="supplier.telefono"><i class="pi pi-phone mr-1 text-[10px]"></i>{{ supplier.telefono }}</div>
                <div v-if="supplier.direccion" class="text-[10px] truncate max-w-[200px] mt-0.5"><i class="pi pi-map-marker mr-1 text-[10px]"></i>{{ supplier.direccion }}</div>
                <div v-if="!supplier.telefono && !supplier.direccion" class="text-slate-400 font-italic">Sin contacto</div>
              </td>
              <td class="py-3.5 px-4 text-right space-x-2">
                <button
                  @click="editSupplier(supplier)"
                  class="p-1.5 text-slate-500 hover:text-indigo-600 hover:bg-indigo-55/10 rounded-lg transition-all"
                  title="Editar"
                >
                  <i class="pi pi-pencil"></i>
                </button>
                <button
                  @click="deleteSupplier(supplier.id)"
                  class="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  title="Eliminar"
                >
                  <i class="pi pi-trash"></i>
                </button>
              </td>
            </tr>
            <tr v-if="suppliers.length === 0">
              <td colspan="3" class="text-center py-8 text-slate-400">No hay proveedores registrados</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import axios from 'axios'

const suppliers = ref<any[]>([])
const editId = ref<string | null>(null)
const queryLoading = ref(false)
const supplierIsGeneric = ref(false)

const form = reactive({
  ruc: '',
  razon_social: '',
  direccion: '',
  telefono: ''
})

watch(() => form.ruc, async (newVal) => {
  if (supplierIsGeneric.value) return
  const cleanRuc = newVal.trim()
  if (cleanRuc.length === 11 && !editId.value) {
    queryLoading.value = true
    try {
      const res = await axios.get(`/public/ruc/${cleanRuc}`)
      if (res.data.success && res.data.data) {
        const d = res.data.data
        form.razon_social = d.razon_social || d.nombre_o_razon_social || d.razonSocial || d.nombre || ''
        
        let address = d.direccion || d.direccion_completa || ''
        if (!address || address.trim() === '-') {
          const parts = []
          if (d.departamento && d.departamento !== '-') parts.push(d.departamento)
          if (d.provincia && d.provincia !== '-') parts.push(d.provincia)
          if (d.distrito && d.distrito !== '-') parts.push(d.distrito)
          address = parts.join(' - ')
        }
        form.direccion = address
      }
    } catch (err) {
      console.error('Error querying RUC', err)
    } finally {
      queryLoading.value = false
    }
  }
})

watch(() => supplierIsGeneric.value, (newVal) => {
  if (newVal) {
    form.ruc = '00000000000'
    form.razon_social = 'PROVEEDOR VARIOS / GENÉRICO'
    form.direccion = 'LIMA'
  } else {
    form.ruc = ''
    form.razon_social = ''
    form.direccion = ''
  }
})

onMounted(() => {
  loadSuppliers()
})

async function loadSuppliers() {
  try {
    const res = await axios.get('/suppliers')
    if (res.data.success) {
      suppliers.value = res.data.data || []
    }
  } catch (err) {
    console.error('Error loading suppliers', err)
  }
}

async function handleSubmit() {
  try {
    if (editId.value) {
      const res = await axios.put(`/suppliers/${editId.value}`, form)
      if (res.data.success) {
        const idx = suppliers.value.findIndex(s => s.id === editId.value)
        if (idx !== -1) suppliers.value[idx] = res.data.data
        cancelEdit()
      }
    } else {
      const res = await axios.post('/suppliers', form)
      if (res.data.success) {
        suppliers.value.push(res.data.data)
        cancelEdit()
      }
    }
  } catch (err) {
    alert('Error al registrar/actualizar el proveedor')
  }
}

function editSupplier(supplier: any) {
  editId.value = supplier.id
  Object.assign(form, {
    ruc: supplier.ruc,
    razon_social: supplier.razon_social,
    direccion: supplier.direccion || '',
    telefono: supplier.telefono || ''
  })
}

function cancelEdit() {
  editId.value = null
  supplierIsGeneric.value = false
  Object.assign(form, {
    ruc: '',
    razon_social: '',
    direccion: '',
    telefono: ''
  })
}

async function deleteSupplier(id: string) {
  if (!confirm('¿Está seguro de eliminar este proveedor?')) return
  try {
    const res = await axios.delete(`/suppliers/${id}`)
    if (res.data.success) {
      suppliers.value = suppliers.value.filter(s => s.id !== id)
    }
  } catch (err) {
    alert('Error al eliminar el proveedor')
  }
}
</script>
