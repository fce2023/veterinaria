<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
    <!-- Categories Panel -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4 flex items-center gap-2">
        <i class="pi pi-tag text-sky-500"></i>
        <span>Categorías de Productos</span>
      </h3>
      
      <!-- Category Form -->
      <form @submit.prevent="handleCategorySubmit" class="flex gap-3 mb-6">
        <input
          v-model="catForm.nombre"
          type="text"
          required
          class="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500 focus:outline-none transition-all"
          :placeholder="catEditId ? 'Editar nombre...' : 'Nueva categoría (Ej. Vacunas)'"
        />
        <button
          type="submit"
          class="px-5 py-2.5 text-xs font-bold text-white bg-sky-600 hover:bg-sky-500 active:bg-sky-700 rounded-xl transition-all shadow-sm flex items-center gap-2"
        >
          <i :class="catEditId ? 'pi pi-check' : 'pi pi-plus'"></i>
          <span>{{ catEditId ? 'Actualizar' : 'Agregar' }}</span>
        </button>
        <button
          v-if="catEditId"
          type="button"
          @click="cancelCatEdit"
          class="px-3 py-2.5 text-xs font-semibold text-slate-500 bg-slate-100 hover:bg-slate-200 rounded-xl transition-all"
        >
          Cancelar
        </button>
      </form>

      <!-- Categories Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Nombre</th>
              <th class="py-3 px-4 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="cat in categories" :key="cat.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="py-3.5 px-4 font-semibold text-slate-900">{{ cat.nombre }}</td>
              <td class="py-3.5 px-4 text-right space-x-2">
                <button
                  @click="editCategory(cat)"
                  class="p-1.5 text-slate-500 hover:text-sky-600 hover:bg-sky-50 rounded-lg transition-all"
                  title="Editar"
                >
                  <i class="pi pi-pencil"></i>
                </button>
                <button
                  @click="deleteCategory(cat.id)"
                  class="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  title="Eliminar"
                >
                  <i class="pi pi-trash"></i>
                </button>
              </td>
            </tr>
            <tr v-if="categories.length === 0">
              <td colspan="2" class="text-center py-8 text-slate-400">No hay categorías registradas</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Brands Panel -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col">
      <h3 class="text-base font-extrabold text-slate-800 border-b border-slate-100 pb-3 mb-4 flex items-center gap-2">
        <i class="pi pi-bookmark text-teal-500"></i>
        <span>Marcas</span>
      </h3>

      <!-- Brand Form -->
      <form @submit.prevent="handleBrandSubmit" class="flex gap-3 mb-6">
        <input
          v-model="brandForm.nombre"
          type="text"
          required
          class="flex-1 px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-teal-500/20 focus:border-teal-500 focus:outline-none transition-all"
          :placeholder="brandEditId ? 'Editar nombre...' : 'Nueva marca (Ej. Royal Canin)'"
        />
        <button
          type="submit"
          class="px-5 py-2.5 text-xs font-bold text-white bg-teal-600 hover:bg-teal-500 active:bg-teal-700 rounded-xl transition-all shadow-sm flex items-center gap-2"
        >
          <i :class="brandEditId ? 'pi pi-check' : 'pi pi-plus'"></i>
          <span>{{ brandEditId ? 'Actualizar' : 'Agregar' }}</span>
        </button>
        <button
          v-if="brandEditId"
          type="button"
          @click="cancelBrandEdit"
          class="px-3 py-2.5 text-xs font-semibold text-slate-500 bg-slate-100 hover:bg-slate-200 rounded-xl transition-all"
        >
          Cancelar
        </button>
      </form>

      <!-- Brands Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-slate-400 text-[10px] font-bold uppercase tracking-wider bg-slate-50">
              <th class="py-3 px-4">Nombre</th>
              <th class="py-3 px-4 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
            <tr v-for="brand in brands" :key="brand.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="py-3.5 px-4 font-semibold text-slate-900">{{ brand.nombre }}</td>
              <td class="py-3.5 px-4 text-right space-x-2">
                <button
                  @click="editBrand(brand)"
                  class="p-1.5 text-slate-500 hover:text-teal-600 hover:bg-teal-50 rounded-lg transition-all"
                  title="Editar"
                >
                  <i class="pi pi-pencil"></i>
                </button>
                <button
                  @click="deleteBrand(brand.id)"
                  class="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all"
                  title="Eliminar"
                >
                  <i class="pi pi-trash"></i>
                </button>
              </td>
            </tr>
            <tr v-if="brands.length === 0">
              <td colspan="2" class="text-center py-8 text-slate-400">No hay marcas registradas</td>
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

const categories = ref<any[]>([])
const brands = ref<any[]>([])

const catForm = reactive({ nombre: '' })
const catEditId = ref<string | null>(null)

const brandForm = reactive({ nombre: '' })
const brandEditId = ref<string | null>(null)

onMounted(() => {
  loadCategories()
  loadBrands()
})

async function loadCategories() {
  try {
    const res = await axios.get('/categories')
    if (res.data.success) {
      categories.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading categories', err)
  }
}

async function loadBrands() {
  try {
    const res = await axios.get('/brands')
    if (res.data.success) {
      brands.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading brands', err)
  }
}

async function handleCategorySubmit() {
  try {
    if (catEditId.value) {
      const res = await axios.put(`/categories/${catEditId.value}`, catForm)
      if (res.data.success) {
        const idx = categories.value.findIndex(c => c.id === catEditId.value)
        if (idx !== -1) categories.value[idx] = res.data.data
        cancelCatEdit()
      }
    } else {
      const res = await axios.post('/categories', catForm)
      if (res.data.success) {
        categories.value.push(res.data.data)
        catForm.nombre = ''
      }
    }
  } catch (err) {
    alert('Error al guardar categoría')
  }
}

function editCategory(cat: any) {
  catEditId.value = cat.id
  catForm.nombre = cat.nombre
}

function cancelCatEdit() {
  catEditId.value = null
  catForm.nombre = ''
}

async function deleteCategory(id: string) {
  if (!confirm('¿Está seguro de eliminar esta categoría?')) return
  try {
    const res = await axios.delete(`/categories/${id}`)
    if (res.data.success) {
      categories.value = categories.value.filter(c => c.id !== id)
    }
  } catch (err) {
    alert('Error al eliminar categoría')
  }
}

async function handleBrandSubmit() {
  try {
    if (brandEditId.value) {
      const res = await axios.put(`/brands/${brandEditId.value}`, brandForm)
      if (res.data.success) {
        const idx = brands.value.findIndex(b => b.id === brandEditId.value)
        if (idx !== -1) brands.value[idx] = res.data.data
        cancelBrandEdit()
      }
    } else {
      const res = await axios.post('/brands', brandForm)
      if (res.data.success) {
        brands.value.push(res.data.data)
        brandForm.nombre = ''
      }
    }
  } catch (err) {
    alert('Error al guardar marca')
  }
}

function editBrand(brand: any) {
  brandEditId.value = brand.id
  brandForm.nombre = brand.nombre
}

function cancelBrandEdit() {
  brandEditId.value = null
  brandForm.nombre = ''
}

async function deleteBrand(id: string) {
  if (!confirm('¿Está seguro de eliminar esta marca?')) return
  try {
    const res = await axios.delete(`/brands/${id}`)
    if (res.data.success) {
      brands.value = brands.value.filter(b => b.id !== id)
    }
  } catch (err) {
    alert('Error al eliminar marca')
  }
}
</script>
