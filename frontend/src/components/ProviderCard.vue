<template>
  <div class="bg-white shadow rounded-lg p-6 hover:shadow-md transition-shadow">
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <div class="flex items-center gap-3 mb-2">
          <h3 class="text-lg font-semibold text-gray-900">
            {{ provider.id }}
          </h3>
          <StatusBadge :status="provider.status" />
        </div>
        
        <p class="text-sm text-gray-600 mb-1">
          Type: <span class="font-medium">{{ provider.type }}</span>
        </p>
        
        <p class="text-xs text-gray-500">
          Last updated: {{ formattedTimestamp }}
        </p>
        
        <p v-if="provider.error_message" class="text-sm text-red-600 mt-2">
          {{ provider.error_message }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import StatusBadge from './StatusBadge.vue'

interface Provider {
  id: string
  type: string
  status: string
  last_updated: string
  error_message?: string
}

interface Props {
  provider: Provider
}

const props = defineProps<Props>()

const formattedTimestamp = computed(() => {
  try {
    const date = new Date(props.provider.last_updated)
    return date.toLocaleString()
  } catch {
    return props.provider.last_updated
  }
})
</script>
