<template>
  <div class="space-y-6 pb-12">
    <!-- Header Section -->
    <div class="flex flex-col gap-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-2xl font-black text-slate-900 tracking-tight">Productos</h3>
          <p class="text-xs text-slate-500 font-bold uppercase tracking-tighter">Catálogo de Inventario</p>
        </div>
        <button 
          @click="showCreateModal = true"
          class="flex items-center gap-2 px-5 py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl text-sm font-black transition-all shadow-xl shadow-indigo-200 active:scale-95"
        >
          <i class="pi pi-plus"></i>
          <span class="hidden sm:inline">Nuevo</span>
        </button>
      </div>

      <!-- Professional Search & Filters -->
      <div class="flex flex-col md:flex-row gap-3">
        <div class="flex-1 relative group">
          <i class="pi pi-search absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-indigo-500 transition-colors"></i>
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="Buscar por nombre o código..."
            class="w-full pl-12 pr-4 py-4 bg-white border border-slate-200 rounded-2xl text-sm font-medium focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all shadow-sm"
          />
        </div>
        <div class="flex gap-2">
          <select 
            v-model="filterCategory"
            class="flex-1 md:flex-none px-4 py-4 bg-white border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none shadow-sm"
          >
            <option value="">Todas las Categorías</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.nombre }}</option>
          </select>
          <button 
            @click="loadProducts"
            class="p-4 bg-white border border-slate-200 text-slate-500 hover:text-indigo-600 rounded-2xl transition-all shadow-sm active:scale-95"
          >
            <i class="pi pi-refresh" :class="{'animate-spin': loading}"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- Responsive Layout: Table for Desktop, Cards for Mobile -->
    <div class="space-y-4">
      <!-- Desktop/Tablet Table -->
      <div class="hidden md:block bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="text-slate-400 text-[10px] font-black uppercase tracking-[0.15em] bg-slate-50/50">
              <th class="py-5 px-6">Producto / Código</th>
              <th class="py-5 px-6">Categoría</th>
              <th class="py-5 px-6 text-right">Precios</th>
              <th class="py-5 px-6 text-center">Estado</th>
              <th class="py-5 px-6 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-sm">
            <tr v-for="item in filteredProducts" :key="item.id" class="hover:bg-slate-50/50 transition-colors group">
              <td class="py-5 px-6">
                <div class="flex flex-col">
                  <div class="flex items-center gap-1.5">
                    <span class="font-bold text-slate-900 group-hover:text-indigo-600 transition-colors">{{ item.nombre }}</span>
                    <span v-if="item.is_dimensional" class="px-1.5 py-0.5 text-[8px] bg-indigo-50 border border-indigo-100 text-indigo-700 font-bold uppercase rounded-md">
                      {{ item.unidad_medida }}
                    </span>
                  </div>
                  <span class="text-[10px] font-black text-slate-400 mt-0.5">{{ item.codigo || 'SIN-COD' }}</span>
                  <span v-if="item.codigo_barras" class="text-[9px] text-slate-400 font-bold flex items-center gap-1 mt-0.5">
                    <i class="pi pi-barcode text-[10px]"></i> {{ item.codigo_barras }}
                  </span>
                </div>
              </td>
              <td class="py-5 px-6">
                <span class="px-2.5 py-1 rounded-lg bg-slate-100 text-slate-600 text-[10px] font-black">
                  {{ getCategoryName(item.category_id) }}
                </span>
              </td>
              <td class="py-5 px-6 text-right">
                <div class="flex flex-col">
                  <span class="font-black text-slate-900 text-base">S/. {{ item.precio_venta.toFixed(2) }}</span>
                  <span class="text-[10px] text-slate-400 font-bold">Costo: S/. {{ item.precio_compra.toFixed(2) }}</span>
                </div>
              </td>
              <td class="py-5 px-6 text-center">
                <span 
                  :class="[
                    'inline-flex items-center px-3 py-1 rounded-full text-[10px] font-black tracking-widest',
                    item.estado === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
                  ]"
                >
                  {{ item.estado === 'active' ? 'ACTIVO' : 'INACTIVO' }}
                </span>
              </td>
              <td class="py-5 px-6 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button @click="editProduct(item)" class="w-9 h-9 flex items-center justify-center bg-blue-50 text-blue-600 hover:bg-blue-600 hover:text-white rounded-xl transition-all shadow-sm">
                    <i class="pi pi-pencil text-xs"></i>
                  </button>
                  <button @click="confirmDelete(item.id)" class="w-9 h-9 flex items-center justify-center bg-red-50 text-red-600 hover:bg-red-600 hover:text-white rounded-xl transition-all shadow-sm">
                    <i class="pi pi-trash text-xs"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Mobile Cards (Visible only on small screens) -->
      <div class="md:hidden grid grid-cols-1 gap-4">
        <div 
          v-for="item in filteredProducts" 
          :key="item.id" 
          class="bg-white p-5 rounded-3xl border border-slate-200 shadow-sm active:scale-[0.98] transition-all relative overflow-hidden"
        >
          <div class="flex justify-between items-start mb-4">
            <div class="flex flex-col">
              <span class="text-[10px] font-black text-indigo-500 uppercase tracking-widest mb-1">
                {{ getCategoryName(item.category_id) }}
              </span>
              <h4 class="font-black text-slate-900 text-lg leading-tight">{{ item.nombre }}</h4>
              <span class="text-[10px] font-bold text-slate-400 mt-1">CÓD: {{ item.codigo || 'N/A' }}</span>
              <span v-if="item.codigo_barras" class="text-[9px] text-slate-400 font-bold flex items-center gap-1 mt-0.5">
                <i class="pi pi-barcode"></i> {{ item.codigo_barras }}
              </span>
            </div>
            <span 
              :class="[
                'px-2 py-1 rounded-lg text-[9px] font-black tracking-tighter',
                item.estado === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
              ]"
            >
              {{ item.estado === 'active' ? 'ACTIVO' : 'INACTIVO' }}
            </span>
          </div>
          
          <div class="flex items-end justify-between mt-6">
            <div class="flex flex-col">
              <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest">Precio Venta</p>
              <p class="text-2xl font-black text-slate-900">S/. {{ item.precio_venta.toFixed(2) }}</p>
            </div>
            <div class="flex gap-2">
              <button @click="editProduct(item)" class="w-12 h-12 flex items-center justify-center bg-slate-100 text-slate-600 rounded-2xl active:bg-indigo-600 active:text-white transition-all">
                <i class="pi pi-pencil text-sm"></i>
              </button>
              <button @click="confirmDelete(item.id)" class="w-12 h-12 flex items-center justify-center bg-red-50 text-red-600 rounded-2xl active:bg-red-600 active:text-white transition-all">
                <i class="pi pi-trash text-sm"></i>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="filteredProducts.length === 0" class="py-20 text-center flex flex-col items-center gap-4">
        <div class="w-20 h-20 bg-slate-100 rounded-full flex items-center justify-center text-slate-300">
          <i class="pi pi-inbox text-4xl"></i>
        </div>
        <div class="max-w-xs">
          <h5 class="font-black text-slate-800 uppercase tracking-tight">Sin resultados</h5>
          <p class="text-xs text-slate-500 font-medium">{{ loading ? 'Cargando catálogo...' : 'No se encontraron productos con los filtros aplicados.' }}</p>
        </div>
      </div>
    </div>

    <!-- Modern Fullscreen Modal for Create/Edit (Mobile-friendly) -->
    <div v-if="showCreateModal" class="fixed inset-0 z-[60] flex items-end sm:items-center justify-center p-0 sm:p-4 animate-fade-in">
      <div @click="resetForm" class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"></div>
      
      <div class="relative w-full max-w-lg bg-white rounded-t-[2.5rem] sm:rounded-[2.5rem] shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <!-- Modal Header -->
        <div class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h4 class="text-xl font-black text-slate-900 tracking-tight">
              {{ editingId ? 'Editar Producto' : 'Nuevo Producto' }}
            </h4>
            <p class="text-[10px] text-indigo-500 font-black uppercase tracking-widest mt-1">Formulario de registro</p>
          </div>
          <button @click="resetForm" class="w-10 h-10 flex items-center justify-center bg-white border border-slate-200 text-slate-400 hover:text-red-500 rounded-2xl shadow-sm transition-all">
            <i class="pi pi-times text-sm"></i>
          </button>
        </div>
        
        <!-- Modal Form Content -->
        <div class="flex-1 overflow-y-auto p-8 space-y-6">
          <form @submit.prevent="handleSubmit" id="product-form" class="space-y-6">
            <div class="space-y-4">
              <h5 class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Información Básica</h5>
              
              <div class="relative group">
                <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 px-1">Nombre del Producto</label>
                <input
                  v-model="productForm.nombre"
                  type="text"
                  required
                  class="block w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all"
                  :placeholder="productPlaceholder"
                />
                <span class="text-[10px] text-slate-400 font-medium mt-1 block px-1">{{ productHelperText }}</span>
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <div class="flex items-center justify-between mb-2 px-1">
                    <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest">Código (SKU)</label>
                    <button 
                      type="button" 
                      @click="autogenerateCode" 
                      class="text-[10px] font-black text-indigo-600 hover:text-indigo-800 uppercase tracking-widest flex items-center gap-1"
                    >
                      <i class="pi pi-cog text-[9px]"></i> Auto
                    </button>
                  </div>
                  <input
                    v-model="productForm.codigo"
                    @input="validateCodeUniqueness"
                    type="text"
                    class="block w-full px-5 py-4 bg-slate-50 border rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all"
                    :class="codeError ? 'border-red-400 bg-red-50/10 text-red-900' : 'border-slate-200'"
                    placeholder="P-001"
                  />
                  <span v-if="codeError" class="text-[9px] text-red-500 font-bold mt-1 block px-1">{{ codeError }}</span>
                </div>
                <div>
                  <div class="flex items-center justify-between mb-2 px-1">
                    <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest">Categoría</label>
                    <button 
                      type="button" 
                      @click="showQuickCategoryModal = true" 
                      class="text-[10px] font-black text-indigo-600 hover:text-indigo-800 uppercase tracking-widest flex items-center gap-1"
                    >
                      <i class="pi pi-plus text-[8px]"></i> Nueva
                    </button>
                  </div>
                  <select 
                    v-model="productForm.category_id"
                    class="block w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none transition-all"
                  >
                    <option value="">General</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.nombre }}</option>
                  </select>
                </div>
              </div>

              <!-- Brand and Barcode Row -->
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <div class="flex items-center justify-between mb-2 px-1">
                    <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest">Código de Barras</label>
                    <button 
                      type="button" 
                      @click="autogenerateBarcode" 
                      class="text-[10px] font-black text-indigo-600 hover:text-indigo-800 uppercase tracking-widest flex items-center gap-1"
                    >
                      <i class="pi pi-cog text-[9px]"></i> Auto
                    </button>
                  </div>
                  <div class="relative">
                    <input
                      v-model="productForm.codigo_barras"
                      @input="validateBarcodeUniqueness"
                      type="text"
                      class="block w-full pl-5 pr-10 py-4 bg-slate-50 border rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all"
                      :class="barcodeError ? 'border-red-400 bg-red-50/10 text-red-900' : 'border-slate-200'"
                      placeholder="EAN-13, UPC o interno"
                    />
                    <div class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
                      <i class="pi pi-barcode text-lg"></i>
                    </div>
                  </div>
                  <span v-if="barcodeError" class="text-[9px] text-red-500 font-bold mt-1 block px-1">{{ barcodeError }}</span>
                  <span class="text-[8px] text-slate-400 font-bold uppercase tracking-wider mt-1 block px-1">Puedes usar pistola lectora</span>
                </div>
                <div>
                  <div class="flex items-center justify-between mb-2 px-1">
                    <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest">Marca</label>
                    <button 
                      type="button" 
                      @click="showQuickBrandModal = true" 
                      class="text-[10px] font-black text-indigo-600 hover:text-indigo-800 uppercase tracking-widest flex items-center gap-1"
                    >
                      <i class="pi pi-plus text-[8px]"></i> Nueva
                    </button>
                  </div>
                  <select 
                    v-model="productForm.brand_id"
                    class="block w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none transition-all"
                  >
                    <option value="">Sin marca</option>
                    <option v-for="brand in brands" :key="brand.id" :value="brand.id">{{ brand.nombre }}</option>
                  </select>
                </div>
              </div>

              <!-- Dimensional Settings -->
              <div class="bg-slate-50 p-6 rounded-3xl border border-slate-200/60 space-y-4">
                <div class="flex items-center gap-3">
                  <input
                    v-model="productForm.is_dimensional"
                    type="checkbox"
                    id="is_dimensional"
                    class="w-5 h-5 text-indigo-600 border-slate-300 rounded-lg focus:ring-indigo-500 cursor-pointer"
                  />
                  <label for="is_dimensional" class="text-xs font-black text-slate-700 uppercase tracking-widest cursor-pointer select-none">
                    Producto Dimensional (M, M², M³)
                  </label>
                </div>
                
                <div v-if="productForm.is_dimensional" class="animate-fade-in space-y-2">
                  <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest px-1">Unidad de Medida</label>
                  <select 
                    v-model="productForm.unidad_medida"
                    class="block w-full px-5 py-4 bg-white border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none transition-all"
                  >
                    <option value="und">Unidad (UND)</option>
                    <option value="m">Metro Lineal (M)</option>
                    <option value="m2">Metro Cuadrado (M²)</option>
                    <option value="m3">Metro Cúbico (M³)</option>
                    <option value="kg">Kilogramo (KG)</option>
                    <option value="lt">Litro (LT)</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="space-y-4 pt-4">
              <h5 class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Finanzas</h5>
              
              <div class="grid grid-cols-2 gap-4">
                <div class="bg-amber-50/50 p-4 rounded-3xl border border-amber-100/50">
                  <label class="block text-[9px] font-black text-amber-600 uppercase tracking-widest mb-2">Precio Compra</label>
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-bold text-amber-600">S/.</span>
                    <input
                      v-model.number="productForm.precio_compra"
                      type="number"
                      step="0.01"
                      class="w-full bg-transparent text-sm font-black text-amber-700 outline-none"
                    />
                  </div>
                </div>
                <div class="bg-indigo-50/50 p-4 rounded-3xl border border-indigo-100/50 shadow-inner">
                  <label class="block text-[9px] font-black text-indigo-600 uppercase tracking-widest mb-2">Precio Venta</label>
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-bold text-indigo-600">S/.</span>
                    <input
                      v-model.number="productForm.precio_venta"
                      type="number"
                      step="0.01"
                      required
                      class="w-full bg-transparent text-lg font-black text-indigo-700 outline-none"
                    />
                  </div>
                </div>
              </div>
            </div>
            
            <!-- Utility Spacer for mobile keyboard -->
            <div class="h-10 sm:hidden"></div>
          </form>
        </div>

        <!-- Modal Footer -->
        <div class="p-8 border-t border-slate-100 bg-white">
          <button
            type="submit"
            form="product-form"
            :disabled="saving"
            class="w-full py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-[1.5rem] text-sm font-black shadow-xl shadow-indigo-200 transition-all flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50"
          >
            <i v-if="saving" class="pi pi-spin pi-spinner"></i>
            <span>{{ editingId ? 'ACTUALIZAR PRODUCTO' : 'GUARDAR PRODUCTO' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Quick Add Category Modal -->
  <div v-if="showQuickCategoryModal" class="fixed inset-0 z-[60] overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 transition-opacity" aria-hidden="true" @click="showQuickCategoryModal = false">
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"></div>
      </div>
      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div class="inline-block align-bottom bg-white rounded-3xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-md sm:w-full border border-slate-100">
        <form @submit.prevent="handleQuickCategorySubmit" class="p-6 space-y-4">
          <div>
            <h3 class="text-lg font-black text-slate-900 tracking-tight">Agregar Categoría Rápida</h3>
            <p class="text-[10px] text-indigo-500 font-black uppercase tracking-widest mt-1">Crea una categoría sin cerrar el formulario</p>
          </div>
          <div class="space-y-2">
            <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest px-1">Nombre de la Categoría</label>
            <input 
              v-model="quickCategoryName"
              type="text"
              required
              class="block w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all"
              placeholder="Ej. Accesorios de Aluminio"
            />
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <button 
              type="button" 
              @click="showQuickCategoryModal = false" 
              class="px-4 py-2 text-xs font-black text-slate-500 uppercase hover:bg-slate-100 rounded-xl"
            >
              Cancelar
            </button>
            <button 
              type="submit" 
              :disabled="quickCategorySaving"
              class="px-4 py-2 text-xs font-black text-white bg-indigo-600 hover:bg-indigo-700 rounded-xl disabled:opacity-50"
            >
              <span v-if="quickCategorySaving">Guardando...</span>
              <span v-else>Guardar</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>

  <!-- Quick Add Brand Modal -->
  <div v-if="showQuickBrandModal" class="fixed inset-0 z-[60] overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 transition-opacity" aria-hidden="true" @click="showQuickBrandModal = false">
        <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"></div>
      </div>
      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div class="inline-block align-bottom bg-white rounded-3xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-md sm:w-full border border-slate-100">
        <form @submit.prevent="handleQuickBrandSubmit" class="p-6 space-y-4">
          <div>
            <h3 class="text-lg font-black text-slate-900 tracking-tight">Agregar Marca Rápida</h3>
            <p class="text-[10px] text-indigo-500 font-black uppercase tracking-widest mt-1">Crea una marca sin cerrar el formulario</p>
          </div>
          <div class="space-y-2">
            <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest px-1">Nombre de la Marca</label>
            <input 
              v-model="quickBrandName"
              type="text"
              required
              class="block w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 focus:border-indigo-500 outline-none transition-all"
              placeholder="Ej. MIROX"
            />
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <button 
              type="button" 
              @click="showQuickBrandModal = false" 
              class="px-4 py-2 text-xs font-black text-slate-500 uppercase hover:bg-slate-100 rounded-xl"
            >
              Cancelar
            </button>
            <button 
              type="submit" 
              :disabled="quickBrandSaving"
              class="px-4 py-2 text-xs font-black text-white bg-indigo-600 hover:bg-indigo-700 rounded-xl disabled:opacity-50"
            >
              <span v-if="quickBrandSaving">Guardando...</span>
              <span v-else>Guardar</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, computed } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()

// Check if the current tenant has the glass store (vidrieria) module active
const isVidrieria = computed(() => {
  return authStore.user?.modules?.includes('vidrieria') && !authStore.user?.modules?.includes('veterinaria')
})

const isBoth = computed(() => {
  return authStore.user?.modules?.includes('vidrieria') && authStore.user?.modules?.includes('veterinaria')
})

const productPlaceholder = computed(() => {
  if (isVidrieria.value) {
    return 'Ej. Vidrio Templado 6mm Incoloro'
  }
  return 'Ej. Nexgard Spectra XL'
})

const productHelperText = computed(() => {
  if (isVidrieria.value) {
    return 'Nombre descriptivo del vidrio, perfil o accesorio'
  }
  if (isBoth.value) {
    return 'Nombre del producto o artículo general'
  }
  return 'Nombre descriptivo del fármaco, vacuna o alimento'
})


const products = ref<any[]>([])
const categories = ref<any[]>([])
const brands = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const searchQuery = ref('')
const filterCategory = ref('')
const showCreateModal = ref(false)
const editingId = ref<string | null>(null)

// Quick add sub-modal variables
const showQuickCategoryModal = ref(false)
const quickCategoryName = ref('')
const quickCategorySaving = ref(false)

const showQuickBrandModal = ref(false)
const quickBrandName = ref('')
const quickBrandSaving = ref(false)

// Inline Validation Error states
const codeError = ref('')
const barcodeError = ref('')

const productForm = reactive({
  nombre: '',
  codigo: '',
  codigo_barras: '',
  category_id: '',
  brand_id: '',
  precio_compra: 0,
  precio_venta: 0,
  estado: 'active',
  is_dimensional: false,
  unidad_medida: 'und'
})

const filteredProducts = computed(() => {
  const list = products.value || []
  return list.filter(p => {
    const nameStr = p.nombre || ''
    const codeStr = p.codigo || ''
    const query = searchQuery.value.toLowerCase()
    
    const matchesSearch = nameStr.toLowerCase().includes(query) || codeStr.toLowerCase().includes(query)
    const matchesCategory = filterCategory.value === '' || p.category_id === filterCategory.value
    return matchesSearch && matchesCategory
  })
})



let barcodeBuffer = ''
let lastKeyTime = 0

const handleGlobalKeyPress = (e: KeyboardEvent) => {
  const target = e.target as HTMLElement
  // Ignore inputs unless it's the barcode input or body keypresses
  if (target.tagName === 'INPUT' || target.tagName === 'SELECT' || target.tagName === 'TEXTAREA') {
    // If the user is currently typing in the barcode field, we don't intercept
    if (target.getAttribute('placeholder') === 'EAN-13, UPC o interno') {
      return
    }
  }

  const currentTime = Date.now()
  // Barcode scanners typically send keys very quickly (< 30ms apart)
  if (currentTime - lastKeyTime > 50) {
    barcodeBuffer = ''
  }

  lastKeyTime = currentTime

  if (e.key !== 'Enter') {
    if (e.key.length === 1 && /[0-9a-zA-Z]/.test(e.key)) {
      barcodeBuffer += e.key
    }
  } else {
    // If we have a buffer (meaning a scanner scanned a barcode)
    if (barcodeBuffer.length >= 4) {
      e.preventDefault()
      
      if (showCreateModal.value) {
        productForm.codigo_barras = barcodeBuffer
        validateBarcodeUniqueness()
      } else {
        // If modal is not open, use it to search products immediately
        searchQuery.value = barcodeBuffer
      }
      
      barcodeBuffer = ''
    }
  }
}

onMounted(async () => {
  window.addEventListener('keypress', handleGlobalKeyPress)
  await Promise.all([
    loadCategories(),
    loadBrands(),
    loadProducts()
  ])
})

onUnmounted(() => {
  window.removeEventListener('keypress', handleGlobalKeyPress)
})

async function loadCategories() {
  try {
    const res = await axios.get('/categories')
    if (res.data.success) categories.value = res.data.data || []
  } catch (err) {
    console.error('Error loading categories', err)
  }
}

async function loadBrands() {
  try {
    const res = await axios.get('/brands')
    if (res.data.success) brands.value = res.data.data || []
  } catch (err) {
    console.error('Error loading brands', err)
  }
}

async function handleQuickCategorySubmit() {
  if (!quickCategoryName.value.trim()) return
  quickCategorySaving.value = true
  try {
    const res = await axios.post('/categories', { nombre: quickCategoryName.value })
    if (res.data.success) {
      categories.value.push(res.data.data)
      productForm.category_id = res.data.data.id
      quickCategoryName.value = ''
      showQuickCategoryModal.value = false
    }
  } catch (err) {
    alert('Error al agregar la categoría rápida')
  } finally {
    quickCategorySaving.value = false
  }
}

async function handleQuickBrandSubmit() {
  if (!quickBrandName.value.trim()) return
  quickBrandSaving.value = true
  try {
    const res = await axios.post('/brands', { nombre: quickBrandName.value })
    if (res.data.success) {
      brands.value.push(res.data.data)
      productForm.brand_id = res.data.data.id
      quickBrandName.value = ''
      showQuickBrandModal.value = false
    }
  } catch (err) {
    alert('Error al agregar la marca rápida')
  } finally {
    quickBrandSaving.value = false
  }
}

async function autogenerateCode() {
  try {
    const res = await axios.get('/products/next-code')
    if (res.data.success) {
      productForm.codigo = res.data.next_code
      codeError.value = ''
    }
  } catch (err) {
    console.error('Error generating next code', err)
  }
}

function autogenerateBarcode() {
  // Generate a standard internal Code128-compatible numeric barcode: date timestamp based
  const timestamp = Date.now().toString()
  productForm.codigo_barras = '750' + timestamp.substring(timestamp.length - 10)
  barcodeError.value = ''
}

async function validateCodeUniqueness() {
  if (!productForm.codigo.trim()) {
    codeError.value = ''
    return
  }
  try {
    const params: any = { codigo: productForm.codigo }
    if (editingId.value) params.exclude_id = editingId.value
    const res = await axios.get('/products/validate-uniqueness', { params })
    if (res.data.success && !res.data.valid) {
      codeError.value = res.data.message
    } else {
      codeError.value = ''
    }
  } catch (err) {
    console.error(err)
  }
}

async function validateBarcodeUniqueness() {
  if (!productForm.codigo_barras.trim()) {
    barcodeError.value = ''
    return
  }
  try {
    const params: any = { codigo_barras: productForm.codigo_barras }
    if (editingId.value) params.exclude_id = editingId.value
    const res = await axios.get('/products/validate-uniqueness', { params })
    if (res.data.success && !res.data.valid) {
      barcodeError.value = res.data.message
    } else {
      barcodeError.value = ''
    }
  } catch (err) {
    console.error(err)
  }
}

async function loadProducts() {
  loading.value = true
  try {
    const res = await axios.get('/products')
    if (res.data.success) products.value = res.data.data || []
  } catch (err) {
    console.error('Error loading products', err)
  } finally {
    loading.value = false
  }
}

function getCategoryName(id: string) {
  const cat = categories.value.find(c => c.id === id)
  return cat ? cat.nombre : 'General'
}

function resetForm() {
  editingId.value = null
  showCreateModal.value = false
  codeError.value = ''
  barcodeError.value = ''
  Object.assign(productForm, {
    nombre: '',
    codigo: '',
    codigo_barras: '',
    category_id: '',
    brand_id: '',
    precio_compra: 0,
    precio_venta: 0,
    estado: 'active',
    is_dimensional: false,
    unidad_medida: 'und'
  })
}

function editProduct(item: any) {
  editingId.value = item.id
  showCreateModal.value = true
  Object.assign(productForm, {
    nombre: item.nombre,
    codigo: item.codigo,
    codigo_barras: item.codigo_barras,
    category_id: item.category_id || '',
    brand_id: item.brand_id || '',
    precio_compra: item.precio_compra,
    precio_venta: item.precio_venta,
    estado: item.estado,
    is_dimensional: item.is_dimensional || false,
    unidad_medida: item.unidad_medida || 'und'
  })
}

async function handleSubmit() {
  // Check validation errors before submitting
  await Promise.all([
    validateCodeUniqueness(),
    validateBarcodeUniqueness()
  ])

  if (codeError.value || barcodeError.value) {
    alert('Por favor corrige los errores del formulario antes de guardar.')
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      const res = await axios.put(`/products/${editingId.value}`, productForm)
      if (res.data.success) {
        const index = products.value.findIndex(p => p.id === editingId.value)
        if (index !== -1) products.value[index] = res.data.data
        resetForm()
      }
    } else {
      const res = await axios.post('/products', productForm)
      if (res.data.success) {
        products.value.unshift(res.data.data)
        resetForm()
      }
    }
  } catch (err) {
    alert('Error al guardar el producto')
  } finally {
    saving.value = false
  }
}

async function confirmDelete(id: string) {
  if (confirm('¿Estás seguro de eliminar este producto?')) {
    try {
      const res = await axios.delete(`/products/${id}`)
      if (res.data.success) {
        products.value = products.value.filter(p => p.id !== id)
      }
    } catch (err) {
      alert('Error al eliminar el producto')
    }
  }
}
</script>
