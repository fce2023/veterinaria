<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
    <!-- Form -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4 h-fit">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 flex items-center gap-2">
        <i class="ti ti-user-plus text-indigo-600"></i>
        <span>{{ editId ? 'Editar Cliente' : 'Nuevo Cliente' }}</span>
      </h3>
      <form @submit.prevent="handleSubmit" class="customer-form space-y-4">
        <div class="form-row">
          <div class="form-group">
            <label class="field-label">Tipo Doc.</label>
            <select
              v-model="form.tipo_documento"
              required
              class="modal-select"
            >
              <option value="DNI">DNI</option>
              <option value="RUC">RUC</option>
              <option value="CE">C.E.</option>
            </select>
          </div>
          <div class="form-group" style="flex: 2;">
            <label class="field-label">N° Documento</label>
            <div class="input-search-group">
              <input
                v-model="form.numero_documento"
                type="text"
                required
                maxlength="11"
                class="modal-input"
                placeholder="8 o 11 dígitos"
                @keyup.enter="searchDocument"
              />
              <button 
                type="button" 
                @click="searchDocument"
                class="btn-search-api"
                :disabled="searchingApi"
                title="Consultar en SUNAT/RENIEC"
              >
                <i v-if="searchingApi" class="ti ti-loader animate-spin"></i>
                <i v-else class="ti ti-search"></i>
              </button>
            </div>
          </div>
        </div>

        <div class="form-group">
          <label class="field-label">Nombre / Razón Social</label>
          <input
            v-model="form.nombre"
            type="text"
            required
            class="modal-input"
            placeholder="Nombre completo"
          />
        </div>

        <div class="form-group">
          <label class="field-label">Dirección</label>
          <input
            v-model="form.direccion"
            type="text"
            class="modal-input"
            placeholder="Dirección opcional"
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="field-label">Teléfono</label>
            <input
              v-model="form.telefono"
              type="text"
              class="modal-input"
              placeholder="Ej. 912345678"
            />
          </div>
          <div class="form-group">
            <label class="field-label">Email</label>
            <input
              v-model="form.email"
              type="email"
              class="modal-input"
              placeholder="correo@ejemplo.com"
            />
          </div>
        </div>

        <div class="flex gap-2 pt-2">
          <button
            type="submit"
            class="flex-1 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center justify-center gap-2"
          >
            <i :class="editId ? 'ti ti-check' : 'ti ti-plus'"></i>
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
                  v-if="authStore.user?.modules?.includes('veterinaria')"
                  @click="openPets(cust)"
                  class="p-1.5 text-slate-500 hover:text-emerald-600 hover:bg-emerald-50 rounded-lg transition-all"
                  title="Mascotas"
                >
                  <i class="pi pi-github"></i>
                </button>
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

  <!-- Pets Management Modal -->
  <div v-if="selectedCustomer" class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
    <div class="bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
      <div class="p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50">
        <div>
          <h3 class="text-lg font-extrabold text-slate-800 flex items-center gap-2">
            <i class="pi pi-github text-emerald-600"></i>
            <span>Mascotas de: {{ selectedCustomer.nombre }}</span>
          </h3>
          <p class="text-xs text-slate-500 font-medium">Gestiona las mascotas vinculadas a este propietario.</p>
        </div>
        <button @click="selectedCustomer = null" class="p-2 hover:bg-slate-200 rounded-full transition-all">
          <i class="pi pi-times text-slate-400"></i>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
        <!-- New Pet Form -->
        <div class="bg-slate-50 p-5 rounded-xl border border-slate-200 space-y-4 h-fit">
          <h4 class="text-xs font-bold text-slate-800 border-b border-slate-200 pb-2 uppercase tracking-wider">
            {{ petEditId ? 'Editar Mascota' : 'Nueva Mascota' }}
          </h4>
          <form @submit.prevent="handlePetSubmit" class="space-y-3">
            <div>
              <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Nombre</label>
              <input v-model="petForm.nombre" type="text" required class="block w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:ring-emerald-500 focus:border-emerald-500 outline-none" />
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Especie</label>
                <select v-model="petForm.especie" class="block w-full px-3 py-2 border border-slate-300 rounded-lg text-xs outline-none">
                  <option value="Perro">Perro</option>
                  <option value="Gato">Gato</option>
                  <option value="Otro">Otro</option>
                </select>
              </div>
              <div>
                <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Sexo</label>
                <select v-model="petForm.sexo" class="block w-full px-3 py-2 border border-slate-300 rounded-lg text-xs outline-none">
                  <option value="Macho">Macho</option>
                  <option value="Hembra">Hembra</option>
                </select>
              </div>
            </div>
            <div>
              <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Raza</label>
              <input v-model="petForm.raza" type="text" class="block w-full px-3 py-2 border border-slate-300 rounded-lg text-xs outline-none" placeholder="Ej. Labrador" />
            </div>
            <button type="submit" class="w-full py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold shadow-sm transition-all">
              {{ petEditId ? 'Actualizar' : 'Agregar Mascota' }}
            </button>
            <button v-if="petEditId" type="button" @click="resetPetForm" class="w-full py-1 text-[10px] text-slate-400 font-bold hover:text-red-500">Cancelar edición</button>
          </form>
        </div>

        <!-- Pets List -->
        <div class="md:col-span-2 space-y-4">
          <div v-if="petLoading" class="flex justify-center py-12">
            <i class="pi pi-spin pi-spinner text-3xl text-emerald-500"></i>
          </div>
          <div v-else-if="pets.length === 0" class="bg-white border-2 border-dashed border-slate-200 rounded-xl p-12 text-center text-slate-400">
            <i class="pi pi-inbox text-4xl mb-2"></i>
            <p class="text-sm">No hay mascotas registradas para este cliente.</p>
          </div>
          <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div v-for="pet in pets" :key="pet.id" class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm hover:border-emerald-200 transition-all group relative">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center font-bold">
                  {{ pet.nombre.charAt(0) }}
                </div>
                <div>
                  <h5 class="font-bold text-slate-900 text-sm">{{ pet.nombre }}</h5>
                  <p class="text-[10px] text-slate-500 font-medium">{{ pet.especie }} • {{ pet.raza || 'Sin raza' }}</p>
                </div>
              </div>
              <div class="mt-3 pt-3 border-t border-slate-50 flex items-center justify-between">
                <span class="text-[9px] font-bold px-1.5 py-0.5 rounded bg-slate-100 text-slate-500 uppercase">{{ pet.sexo }}</span>
                <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-all">
                  <button @click="editPet(pet)" class="p-1 text-blue-500 hover:bg-blue-50 rounded"><i class="pi pi-pencil text-xs"></i></button>
                  <button @click="deletePet(pet.id)" class="p-1 text-red-500 hover:bg-red-50 rounded"><i class="pi pi-trash text-xs"></i></button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()

const customers = ref<any[]>([])
const editId = ref<string | null>(null)
const searchingApi = ref(false)

// Pets management state
const selectedCustomer = ref<any>(null)
const pets = ref<any[]>([])
const petLoading = ref(false)
const petEditId = ref<string | null>(null)
const petForm = reactive({
  nombre: '',
  especie: 'Perro',
  raza: '',
  sexo: 'Macho',
  customer_id: ''
})

const form = reactive({
  tipo_documento: 'DNI',
  numero_documento: '',
  nombre: '',
  direccion: '',
  telefono: '',
  email: ''
})

onMounted(() => {
  loadCustomers()
})

async function searchDocument() {
  const doc = form.numero_documento.trim()
  const type = form.tipo_documento
  
  if (type === 'DNI' && doc.length !== 8) return
  if (type === 'RUC' && doc.length !== 11) return
  if (type === 'CE') return

  searchingApi.value = true
  try {
    const endpoint = type === 'DNI' ? `/public/dni/${doc}` : `/public/ruc/${doc}`
    const res = await axios.get(endpoint)
    if (res.data.success) {
      const info = res.data.data || {}
      form.nombre = info.nombre || info.razon_social || info.nombre_o_razon_social || ''
      
      let address = info.direccion || info.direccion_completa || ''
      if (!address || address.trim() === '-') {
        const parts = []
        if (info.departamento && info.departamento !== '-') parts.push(info.departamento)
        if (info.provincia && info.provincia !== '-') parts.push(info.provincia)
        if (info.distrito && info.distrito !== '-') parts.push(info.distrito)
        address = parts.join(' - ')
      }
      form.direccion = address
    }
  } catch (err: any) {
    console.error('Error querying document', err)
  } finally {
    searchingApi.value = false
  }
}

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
    telefono: cust.telefono || '',
    email: cust.email || ''
  })
}

function cancelEdit() {
  editId.value = null
  Object.assign(form, {
    tipo_documento: 'DNI',
    numero_documento: '',
    nombre: '',
    direccion: '',
    telefono: '',
    email: ''
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

// Pet Methods
async function openPets(cust: any) {
  selectedCustomer.value = cust
  petForm.customer_id = cust.id
  await loadPets(cust.id)
}

async function loadPets(customerId: string) {
  petLoading.value = true
  try {
    const res = await axios.get(`/pets?customer_id=${customerId}`)
    if (res.data.success) {
      pets.value = res.data.data || []
    }
  } catch (err) {
    console.error('Error loading pets', err)
  } finally {
    petLoading.value = false
  }
}

async function handlePetSubmit() {
  try {
    if (petEditId.value) {
      const res = await axios.put(`/pets/${petEditId.value}`, petForm)
      if (res.data.success) {
        const idx = pets.value.findIndex(p => p.id === petEditId.value)
        if (idx !== -1) pets.value[idx] = res.data.data
        resetPetForm()
      }
    } else {
      const res = await axios.post('/pets', petForm)
      if (res.data.success) {
        pets.value.push(res.data.data)
        resetPetForm()
      }
    }
  } catch (err) {
    alert('Error al guardar la mascota')
  }
}

function editPet(pet: any) {
  petEditId.value = pet.id
  Object.assign(petForm, {
    nombre: pet.nombre,
    especie: pet.especie,
    raza: pet.raza,
    sexo: pet.sexo,
    customer_id: pet.customer_id
  })
}

function resetPetForm() {
  petEditId.value = null
  Object.assign(petForm, {
    nombre: '',
    especie: 'Perro',
    raza: '',
    sexo: 'Macho'
  })
}

async function deletePet(id: string) {
  if (!confirm('¿Seguro que desea eliminar esta mascota?')) return
  try {
    const res = await axios.delete(`/pets/${id}`)
    if (res.data.success) {
      pets.value = pets.value.filter(p => p.id !== id)
    }
  } catch (err) {
    alert('Error al eliminar la mascota')
  }
}
</script>

<style scoped>
.customer-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 11px;
  font-weight: 800;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.modal-input, .modal-select {
  width: 100%;
  height: 40px;
  padding: 0 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  font-size: 13px;
  color: #1e293b;
  outline: none;
  transition: all 0.2s;
}

.modal-input:focus, .modal-select:focus {
  border-color: #6366f1;
  background: white;
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
}

.input-search-group {
  display: flex;
  gap: 8px;
}

.btn-search-api {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #6366f1;
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.2);
}

.btn-search-api:hover:not(:disabled) {
  background: #4f46e5;
  transform: translateY(-1px);
}

.btn-search-api:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
