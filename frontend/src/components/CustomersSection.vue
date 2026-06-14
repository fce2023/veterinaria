<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
    <!-- Form -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm space-y-4 h-fit">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 flex items-center gap-2">
        <i class="pi pi-plus text-purple-600"></i>
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
            class="flex-1 py-2.5 bg-purple-600 hover:bg-purple-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center justify-center gap-2"
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
