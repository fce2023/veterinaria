<template>
  <div v-if="branches.length > 1" class="branch-selector">
    <button 
      @click="isOpen = !isOpen"
      class="selector-btn"
    >
      <i class="ti ti-map-pin"></i>
      <span class="btn-text">{{ currentBranchName }}</span>
      <i class="ti ti-chevron-down" :class="{ 'rotate': isOpen }"></i>
    </button>

    <!-- Dropdown -->
    <div v-if="isOpen" class="selector-dropdown">
      <div class="dropdown-label">Cambiar de Sede</div>
      <button
        v-for="branch in branches"
        :key="branch.id"
        @click="selectBranch(branch.id)"
        class="dropdown-item"
        :class="{ active: branch.id === authStore.user?.branch_id }"
      >
        <span>{{ branch.nombre }}</span>
        <i v-if="branch.id === authStore.user?.branch_id" class="ti ti-check"></i>
      </button>
    </div>

    <!-- Overlay to close -->
    <div v-if="isOpen" @click="isOpen = false" class="selector-overlay"></div>
  </div>
  <div v-else-if="branches.length === 1" class="branch-badge">
    <i class="ti ti-map-pin"></i>
    <span>{{ branches[0].nombre }}</span>
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

<style scoped>
.branch-selector { position: relative; }
.selector-btn {
  display: flex; align-items: center; gap: 8px; padding: 8px 12px;
  background: var(--surface2); border: 1px solid var(--border);
  border-radius: var(--radius-sm); color: var(--text);
  font-family: inherit; font-size: 13px; font-weight: 600;
  cursor: pointer; transition: all var(--transition);
}
.selector-btn:hover { border-color: var(--accent); background: var(--surface); }
.btn-text { max-width: 150px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ti-chevron-down { font-size: 14px; transition: transform var(--transition); }
.ti-chevron-down.rotate { transform: rotate(180deg); }

.selector-dropdown {
  position: absolute; right: 0; top: calc(100% + 8px);
  width: 220px; background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius); shadow: 0 10px 25px rgba(0,0,0,0.1);
  padding: 8px; z-index: 200;
}
.dropdown-label { font-size: 10px; font-weight: 800; color: var(--text3); text-transform: uppercase; padding: 8px 12px; }
.dropdown-item {
  width: 100%; display: flex; align-items: center; justify-content: space-between;
  padding: 10px 12px; border-radius: var(--radius-sm); border: none;
  background: transparent; color: var(--text2); font-family: inherit;
  font-size: 13px; font-weight: 500; cursor: pointer; transition: all var(--transition);
}
.dropdown-item:hover { background: var(--surface2); color: var(--text); }
.dropdown-item.active { background: var(--accent-dim); color: var(--accent); font-weight: 700; }

.selector-overlay { position: fixed; inset: 0; z-index: 150; }

.branch-badge {
  display: flex; align-items: center; gap: 6px; padding: 6px 12px;
  background: var(--accent-dim); color: var(--accent);
  border-radius: 50px; font-size: 12px; font-weight: 700;
}
</style>
