<template>
  <div class="container mx-auto px-4 py-10 space-y-8">
    <!-- Hero / actions -->
    <section class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-indigo-900 to-slate-800 text-white shadow-xl">
      <div class="absolute inset-0 opacity-30 bg-[radial-gradient(circle_at_20%_20%,rgba(255,255,255,0.18),transparent_35%),radial-gradient(circle_at_80%_0,rgba(255,255,255,0.12),transparent_30%)]"></div>
      <div class="relative px-6 py-8 sm:px-8 sm:py-10">
        <div class="flex flex-wrap items-start gap-6 justify-between">
          <div class="max-w-2xl space-y-3">
            <p class="text-sm uppercase tracking-[0.2em] text-indigo-200">Operations</p>
            <h1 class="text-3xl sm:text-4xl font-semibold">Notification Server Dashboard</h1>
            <p class="text-indigo-100 text-base sm:text-lg">
              Monitor providers, trigger history views, and jump straight to your API docs.
            </p>
            <div class="flex flex-wrap gap-3 pt-2">
              <router-link
                to="/history"
                class="inline-flex items-center gap-2 rounded-lg bg-white/10 px-4 py-2 text-sm font-medium text-white hover:bg-white/15 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-slate-900 focus:ring-white"
              >
                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                View History
              </router-link>
              <a
                :href="apiDocsUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-2 rounded-lg bg-white text-slate-900 px-4 py-2 text-sm font-semibold shadow hover:-translate-y-0.5 transition transform duration-150"
              >
                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                API Docs
              </a>
              <button
                type="button"
                @click="loadProviders"
                class="inline-flex items-center gap-2 rounded-lg border border-white/30 px-4 py-2 text-sm font-medium text-white hover:bg-white/10 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-slate-900 focus:ring-white"
              >
                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9M20 20v-5h-.581m-15.357-2a8.003 8.003 0 0115.357 2" />
                </svg>
                Retry
              </button>
            </div>
          </div>
          <div class="flex flex-col gap-3 text-sm text-indigo-100">
            <div class="inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1">
              <span class="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
              Auto-refresh every 30s
            </div>
            <div class="inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1">
              <span class="text-xs uppercase tracking-wide text-indigo-200">Last refreshed</span>
              <span class="font-semibold">{{ lastRefreshedDisplay }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="relative px-6 pb-6 sm:px-8">
        <div class="grid gap-4 sm:grid-cols-2 -mt-6">
          <div class="rounded-xl bg-white/90 backdrop-blur text-slate-900 p-4 shadow">
            <p class="text-sm text-slate-500">Providers</p>
            <p class="text-2xl font-semibold">{{ providerCount }}</p>
          </div>
          <div class="rounded-xl bg-white/90 backdrop-blur text-slate-900 p-4 shadow">
            <p class="text-sm text-slate-500">Status</p>
            <p class="text-base font-medium text-emerald-600">Live</p>
          </div>
        </div>

        <div class="mt-6 rounded-2xl bg-white/90 backdrop-blur text-slate-900 p-4 sm:p-6 shadow">
          <!-- Loading state -->
          <div v-if="loading" class="flex items-center justify-center py-12">
            <div class="text-center">
              <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mb-4"></div>
              <p class="text-gray-600">Loading providers...</p>
            </div>
          </div>

          <!-- Error state -->
          <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4">
            <div class="flex items-center gap-3">
              <svg class="h-5 w-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-red-900">Failed to load providers</h3>
                <p class="text-sm text-red-700 mt-1">{{ error }}</p>
              </div>
              <button 
                @click="loadProviders"
                class="px-3 py-1.5 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors text-sm"
              >
                Retry
              </button>
            </div>
          </div>

          <!-- Empty state -->
          <div v-else-if="providers.length === 0" class="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
            <svg class="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <h3 class="text-lg font-semibold text-gray-900 mb-2">No providers configured</h3>
            <p class="text-gray-600">Add provider configuration files to the CONFIG_DIR to get started.</p>
          </div>

          <!-- Provider grid -->
          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
            <ProviderCard 
              v-for="provider in providers" 
              :key="provider.id"
              :provider="provider"
            />
          </div>

          <!-- Last updated footer -->
          <div v-if="!loading && !error" class="mt-6 text-center text-sm text-gray-500">
            Last refreshed: {{ lastRefreshed }}
            <span class="mx-2">•</span>
            Auto-refresh every 30 seconds
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { fetchProviders, type Provider } from '../services/api'
import ProviderCard from '../components/ProviderCard.vue'

const providers = ref<Provider[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const lastRefreshed = ref<string>('')
let refreshInterval: number | null = null

const providerCount = computed(() => providers.value.length)
const lastRefreshedDisplay = computed(() => lastRefreshed.value || '—')

// URL for API documentation; override via VITE_API_DOCS_URL
const apiDocsUrl = import.meta.env.VITE_API_DOCS_URL || `${window.location.origin}/api/v1/docs`

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
