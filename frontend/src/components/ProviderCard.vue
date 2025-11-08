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
        
        <!-- T060: Display last test status -->
        <p v-if="provider.last_test_at" class="text-xs text-gray-500 mt-1">
          Last tested: {{ formattedLastTest }} - 
          <span :class="testStatusClass">{{ provider.last_test_status }}</span>
        </p>
        
        <p v-if="provider.error_message" class="text-sm text-red-600 mt-2">
          {{ provider.error_message }}
        </p>

        <!-- T059: Test result display -->
        <div v-if="testResult" class="mt-3 p-3 rounded-md" :class="testResultClass">
          <p class="text-sm font-medium">{{ testResult.message }}</p>
          <p v-if="testResult.error_details" class="text-xs mt-1">
            {{ formatErrorMessage(testResult.error_details) }}
          </p>
        </div>
      </div>

      <!-- T057: Test button with loading state -->
      <div class="flex-shrink-0 ml-4">
        <button
          @click="handleTestClick"
          :disabled="isTestLoading || provider.status === 'error'"
          class="px-4 py-2 text-sm font-medium rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          :class="testButtonClass"
        >
          <span v-if="isTestLoading" class="flex items-center gap-2">
            <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Testing...
          </span>
          <span v-else>Test</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import StatusBadge from './StatusBadge.vue'
import { testProvider } from '../services/api'

interface Provider {
  id: string
  type: string
  status: string
  last_updated: string
  error_message?: string
  last_test_at?: string
  last_test_status?: string
}

interface Props {
  provider: Provider
}

interface TestResult {
  result: 'success' | 'failed'
  message: string
  error_details?: string
  tested_at: string
}

const props = defineProps<Props>()

// T058: Test button state management
const isTestLoading = ref(false)
const testResult = ref<TestResult | null>(null)

const formattedTimestamp = computed(() => {
  try {
    const date = new Date(props.provider.last_updated)
    return date.toLocaleString()
  } catch {
    return props.provider.last_updated
  }
})

// T060: Format last test timestamp
const formattedLastTest = computed(() => {
  if (!props.provider.last_test_at) return ''
  try {
    const date = new Date(props.provider.last_test_at)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    
    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`
    
    const diffHours = Math.floor(diffMins / 60)
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    
    return date.toLocaleString()
  } catch {
    return props.provider.last_test_at
  }
})

// T060: Test status styling
const testStatusClass = computed(() => {
  if (props.provider.last_test_status === 'success') {
    return 'text-green-600 font-medium'
  } else if (props.provider.last_test_status === 'failed') {
    return 'text-red-600 font-medium'
  }
  return 'text-gray-600'
})

// T057, T062: Test button styling
const testButtonClass = computed(() => {
  if (isTestLoading.value) {
    return 'bg-blue-100 text-blue-700'
  }
  return 'bg-blue-600 text-white hover:bg-blue-700'
})

// T059: Test result styling
const testResultClass = computed(() => {
  if (!testResult.value) return ''
  if (testResult.value.result === 'success') {
    return 'bg-green-50 border border-green-200 text-green-800'
  }
  return 'bg-red-50 border border-red-200 text-red-800'
})

// T063: Format error messages to be user-friendly
const formatErrorMessage = (error: string): string => {
  // Remove Go stack traces and technical details
  const cleanError = error.replace(/\n.*at .*/g, '')
  
  // Extract meaningful error messages
  if (error.includes('connection refused')) {
    return 'Unable to connect to the service. Please check network connectivity and firewall settings.'
  }
  if (error.includes('timeout')) {
    return 'Request timed out. The service may be slow or unavailable.'
  }
  if (error.includes('401') || error.includes('unauthorized')) {
    return 'Authentication failed. Please verify credentials in the provider configuration.'
  }
  if (error.includes('403') || error.includes('forbidden')) {
    return 'Access denied. Please check permissions for this provider.'
  }
  if (error.includes('SMTP')) {
    const match = error.match(/smtp\.([^:]+):(\d+)/)
    if (match) {
      return `Failed to connect to SMTP server at ${match[1]}:${match[2]}. Please verify server address and port, and check firewall rules.`
    }
  }
  if (error.includes('Telegram API')) {
    if (error.includes('401')) {
      return 'Telegram bot token is invalid. Please verify the bot token in the provider configuration.'
    }
  }
  
  // Return cleaned error if no specific format matched
  return cleanError.trim() || error
}

// T058, T062: Test button click handler with loading feedback
const handleTestClick = async () => {
  // T062: Show loading state within 100ms
  isTestLoading.value = true
  testResult.value = null
  
  try {
    const result = await testProvider(props.provider.id)
    
    // T059: Display test result
    testResult.value = result
    
    // Auto-hide success message after 5 seconds
    if (result.result === 'success') {
      setTimeout(() => {
        testResult.value = null
      }, 5000)
    }
  } catch (error) {
    // T059: Display error with details
    testResult.value = {
      result: 'failed',
      message: 'Test request failed',
      error_details: error instanceof Error ? error.message : 'Unknown error occurred',
      tested_at: new Date().toISOString()
    }
  } finally {
    isTestLoading.value = false
  }
}
</script>
