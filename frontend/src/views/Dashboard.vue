<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">Provider Dashboard</h1>
    
    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="text-center">
        <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mb-4"></div>
        <p class="text-gray-600">Loading providers...</p>
      </div>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-6 mb-6">
      <div class="flex items-center gap-3">
        <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div class="flex-1">
          <h3 class="text-lg font-semibold text-red-900">Failed to load providers</h3>
          <p class="text-sm text-red-700 mt-1">{{ error }}</p>
        </div>
        <button 
          @click="loadProviders"
          class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
        >
          Retry
        </button>
      </div>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="providers.length === 0" class="bg-gray-50 border border-gray-200 rounded-lg p-12 text-center">
      <svg class="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">No providers configured</h3>
      <p class="text-gray-600">Add provider configuration files to the CONFIG_DIR to get started.</p>
    </div>
    
    <!-- Provider grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <ProviderCard 
        v-for="provider in providers" 
        :key="provider.id"
        :provider="provider"
      />
    </div>
    
    <!-- Last updated footer -->
    <div v-if="!loading && !error" class="mt-8 text-center text-sm text-gray-500">
      Last refreshed: {{ lastRefreshed }}
      <span class="mx-2">•</span>
      Auto-refresh every 30 seconds
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { fetchProviders, type Provider } from '../services/api'
import ProviderCard from '../components/ProviderCard.vue'

const providers = ref<Provider[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const lastRefreshed = ref<string>('')
let refreshInterval: number | null = null

async function loadProviders() {
  try {
    loading.value = true
    error.value = null
    providers.value = await fetchProviders()
    lastRefreshed.value = new Date().toLocaleTimeString()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error occurred'
    console.error('Failed to load providers:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  // Initial load
  loadProviders()
  
  // Set up auto-refresh every 30 seconds
  refreshInterval = window.setInterval(() => {
    loadProviders()
  }, 30000)
})

onUnmounted(() => {
  // Clean up interval on component unmount
  if (refreshInterval !== null) {
    clearInterval(refreshInterval)
  }
})
</script>
