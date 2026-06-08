<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
    <!-- Form -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4 h-fit">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 flex items-center gap-2">
        <i class="pi pi-plus text-purple-550"></i>
        <span>{{ editId ? 'Editar Cliente' : 'Nuevo Cliente' }}</span>
      </h3>
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-bold text-slate-600 mb-1.5">Tipo Doc.</label>
            <select
              v-model="form.tipo_documento"
              required
              class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 focus:outline-none transition-all"
            >
              <option value="DNI">DNI</option>
              <option value="RUC">RUC</option>
              <option value="CE">C.E.</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-600 mb-1.5">Nº Documento</label>
            <input
              v-model="form.numero_documento"
              type="text"
              required
              class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 focus:outline-none transition-all"
              placeholder="Ej. 70123456"
            />
          </div>
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Nombres / Razón Social</label>
          <input
            v-model="form.nombre"
            type="text"
            required
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 focus:outline-none transition-all"
            placeholder="Ej. Juan Pérez Celis"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Dirección</label>
          <input
            v-model="form.direccion"
            type="text"
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 focus:outline-none transition-all"
            placeholder="Dirección del cliente"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-600 mb-1.5">Teléfono</label>
          <input
            v-model="form.telefono"
            type="text"
            class="block w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 focus:outline-none transition-all"
            placeholder="Ej. 912345678"
          />
        </div>
        <div class="flex gap-2 pt-2">
          <button
            type="submit"
            class="flex-1 py-2.5 bg-purple-650 hover:bg-purple-550 active:bg-purple-750 text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center justify-center gap-2"
            style="background-color: rgb(147, 51, 234);"
          >
            <i :class="editId ? 'pi pi-check' : 'pi pi-plus'"></i>
            <span>{{ editId ? 'Guardar' : 'Registrar Cliente' }}</span>
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
        <i class="pi pi-users text-purple-600"></i>
        <span>Cartera de Clientes</span>
      </h3>
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Documento / Cliente</th>
              <th class="py-3 px-4">Contacto</th>
              <th class="py-3 px-4 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="cust in customers" :key="cust.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="py-3.5 px-4">
                <div class="font-bold text-slate-900">{{ cust.nombre }}</div>
                <div class="text-[10px] font-mono text-slate-400 mt-0.5">{{ cust.tipo_documento }}: {{ cust.numero_documento }}</div>
              </td>
              <td class="py-3.5 px-4 text-slate-500">
                <div v-if="cust.telefono"><i class="pi pi-phone mr-1 text-[10px]"></i>{{ cust.telefono }}</div>
                <div v-if="cust.direccion" class="text-[10px] truncate max-w-[200px] mt-0.5"><i class="pi pi-map-marker mr-1 text-[10px]"></i>{{ cust.direccion }}</div>
                <div v-if="!cust.telefono && !cust.direccion" class="text-slate-400 font-italic">Sin datos de contacto</div>
              </td>
              <td class="py-3.5 px-4 text-right space-x-2">
                <button
                  @click="editCustomer(cust)"
                  class="p-1.5 text-slate-500 hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-all"
                  title="Editar"
                >
                  <i class="pi pi-pencil"></i>
                </button>
                <button
                  @click="deleteCustomer(cust.id)"
                  class="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  title="Eliminar"
                >
                  <i class="pi pi-trash"></i>
                </button>
              </td>
            </tr>
            <tr v-if="customers.length === 0">
              <td colspan="3" class="text-center py-8 text-slate-400">No hay clientes registrados</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'

const customers = ref<any[]>([])
const editId = ref<string | null>(null)

const form = reactive({
  tipo_documento: 'DNI',
  numero_documento: '',
  nombre: '',
  direccion: '',
  telefono: ''
})

onMounted(() => {
  loadCustomers()
})

async function loadCustomers() {
  try {
    const res = await axios.get('/customers')
    if (res.data.success) {
      customers.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading customers', err)
  }
}

async function handleSubmit() {
  try {
    if (editId.value) {
      const res = await axios.put(`/customers/${editId.value}`, form)
      if (res.data.success) {
        const idx = customers.value.findIndex(c => c.id === editId.value)
        if (idx !== -1) customers.value[idx] = res.data.data
        cancelEdit()
      }
    } else {
      const res = await axios.post('/customers', form)
      if (res.data.success) {
        customers.value.push(res.data.data)
        cancelEdit()
      }
    }
  } catch (err) {
    alert('Error al registrar/actualizar el cliente')
  }
}

function editCustomer(cust: any) {
  editId.value = cust.id
  Object.assign(form, {
    tipo_documento: cust.tipo_documento,
    numero_documento: cust.numero_documento,
    nombre: cust.nombre,
    direccion: cust.direccion || '',
    telefono: cust.telefono || ''
  })
}

function cancelEdit() {
  editId.value = null
  Object.assign(form, {
    tipo_documento: 'DNI',
    numero_documento: '',
    nombre: '',
    direccion: '',
    telefono: ''
  })
}

async function deleteCustomer(id: string) {
  if (!confirm('¿Está seguro de eliminar este cliente?')) return
  try {
    const res = await axios.delete(`/customers/${id}`)
    if (res.data.success) {
      customers.value = customers.value.filter(c => c.id !== id)
    }
  } catch (err) {
    alert('Error al eliminar el cliente')
  }
}
</script>
