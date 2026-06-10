<template>
  <div v-if="branches.length > 1" class="relative">
    <button 
      @click="isOpen = !isOpen"
      class="flex items-center space-x-2 bg-white/10 hover:bg-white/20 px-3 py-1.5 rounded-lg transition-colors border border-white/20"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
      <span class="text-sm font-semibold truncate max-w-[150px]">{{ currentBranchName }}</span>
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 transition-transform duration-200" :class="{ 'rotate-180': isOpen }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- Dropdown -->
    <div v-if="isOpen" 
      class="absolute right-0 mt-2 w-56 bg-white rounded-xl shadow-xl border border-gray-100 py-2 z-50 transform origin-top-right transition-all animate-in fade-in zoom-in duration-200"
    >
      <div class="px-4 py-2 text-xs font-bold text-gray-400 uppercase tracking-wider">Cambiar de Sede</div>
      <button
        v-for="branch in branches"
        :key="branch.id"
        @click="selectBranch(branch.id)"
        class="w-full text-left px-4 py-2.5 text-sm hover:bg-indigo-50 flex items-center justify-between group transition-colors"
        :class="branch.id === authStore.user?.branch_id ? 'text-indigo-700 bg-indigo-50/50 font-bold' : 'text-gray-700'"
      >
        <span>{{ branch.nombre }}</span>
        <svg v-if="branch.id === authStore.user?.branch_id" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </button>
    </div>

    <!-- Overlay to close -->
    <div v-if="isOpen" @click="isOpen = false" class="fixed inset-0 z-40"></div>
  </div>
  <div v-else-if="branches.length === 1" class="flex items-center space-x-2 text-white/80 text-sm px-3 py-1.5">
    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
    </svg>
    <span class="font-medium">{{ branches[0].nombre }}</span>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const branches = ref<any[]>([])
const isOpen = ref(false)

const currentBranchName = computed(() => {
  const branch = branches.value.find(b => b.id === authStore.user?.branch_id)
  return branch ? branch.nombre : 'Seleccionar Sede'
})

const fetchBranches = async () => {
  try {
    const res = await axios.get('/branches')
    if (res.data.success) {
      branches.value = res.data.data || []
    }
  } catch (err) {
    console.error('Error fetching branches', err)
  }
}

const selectBranch = async (branchId: string) => {
  if (branchId === authStore.user?.branch_id) {
    isOpen.value = false
    return
  }
  
  const success = await authStore.switchBranch(branchId)
  if (success) {
    isOpen.value = false
  }
}

onMounted(fetchBranches)
</script>
