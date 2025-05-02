<!-- src/components/MyNewComponent.vue -->
<script setup lang="ts">
// Component logic here
import { ref, onMounted } from 'vue'
import { getCount, updateCount } from '@/services/api'

const count = ref<number>(0)
const loading = ref<boolean>(false)
const error = ref<string>('')

const fetchCount = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await getCount()
    count.value = response.data
  } catch (err) {
    error.value = 'Failed to fetch count'
    console.error(err)
  } finally {
    loading.value = false
  }
}

const incrementCount = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await updateCount()
    count.value++
  } catch (err) {
    error.value = "Failed to update count"
    console.error(err)
  } finally {
    loading.value = false
  }
}

onMounted(fetchCount)
</script>

<template>
  <div class="counter-container p-6 max-w-md mx-auto bg-white rounded-xl shadow-md">
    <h2 class="text-2xl font-bold mb-4">Counter</h2>

    <div v-if="loading" class="flex justify-center my-4">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
    </div>

    <div v-else-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      {{ error }}
    </div>

    <div v-else class="text-center">
      <div class="text-6xl font-bold mb-6">{{ count }}</div>

      <button
        @click="incrementCount"
        class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
        :disabled="loading"
      >
        Increment
      </button>
    </div>
  </div>
</template>


<style scoped>
/* Component styles here */
</style>

