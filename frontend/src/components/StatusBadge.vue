<template>
  <span 
    :class="badgeClasses"
    :aria-label="`Status: ${status}`"
    role="status"
    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
  >
    {{ statusText }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  status: string
}

const props = defineProps<Props>()

const badgeClasses = computed(() => {
  const normalizedStatus = props.status.toLowerCase()
  
  switch (normalizedStatus) {
    case 'active':
      return 'bg-green-100 text-green-800'
    case 'error':
      return 'bg-red-100 text-red-800'
    case 'disabled':
      return 'bg-gray-100 text-gray-800'
    case 'initializing':
      return 'bg-yellow-100 text-yellow-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
})

const statusText = computed(() => {
  return props.status.charAt(0).toUpperCase() + props.status.slice(1)
})
</script>
