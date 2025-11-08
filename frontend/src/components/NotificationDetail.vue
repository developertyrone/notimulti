<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <!-- Background overlay -->
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" @click="close"></div>

      <!-- Modal panel -->
      <div class="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-3xl sm:w-full sm:p-6">
        <div class="sm:flex sm:items-start">
          <div class="w-full mt-3 text-center sm:mt-0 sm:text-left">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                Notification Details
              </h3>
              <button
                @click="close"
                class="text-gray-400 hover:text-gray-500 focus:outline-none"
              >
                <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Loading State -->
            <div v-if="loading" class="py-8">
              <div class="flex items-center justify-center">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500"></div>
                <span class="ml-3 text-gray-600">Loading details...</span>
              </div>
            </div>

            <!-- Error State -->
            <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4">
              <p class="text-sm text-red-700">{{ error }}</p>
            </div>

            <!-- Notification Details -->
            <div v-else-if="notification" class="mt-4">
              <!-- Status and Type Badges -->
              <div class="flex items-center space-x-2 mb-4">
                <StatusBadge :status="notification.status" />
                <span v-if="notification.is_test" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  TEST NOTIFICATION
                </span>
              </div>

              <!-- Details Grid -->
              <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                <!-- ID -->
                <div class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Notification ID</dt>
                  <dd class="mt-1 text-sm text-gray-900 font-mono">{{ notification.id }}</dd>
                </div>

                <!-- Provider -->
                <div class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Provider</dt>
                  <dd class="mt-1 text-sm text-gray-900">
                    {{ notification.provider_id }} 
                    <span class="text-gray-500">({{ notification.provider_type }})</span>
                  </dd>
                </div>

                <!-- Recipient -->
                <div class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Recipient</dt>
                  <dd class="mt-1 text-sm text-gray-900 font-mono">{{ notification.recipient }}</dd>
                </div>

                <!-- Priority -->
                <div class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Priority</dt>
                  <dd class="mt-1 text-sm text-gray-900 capitalize">{{ notification.priority }}</dd>
                </div>

                <!-- Attempts -->
                <div class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Delivery Attempts</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ notification.attempts }}</dd>
                </div>

                <!-- Created At -->
                <div class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Created At</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ formatDate(notification.created_at) }}</dd>
                </div>

                <!-- Delivered At -->
                <div v-if="notification.delivered_at" class="sm:col-span-1">
                  <dt class="text-sm font-medium text-gray-500">Delivered At</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ formatDate(notification.delivered_at) }}</dd>
                </div>

                <!-- Subject (if present) -->
                <div v-if="notification.subject" class="sm:col-span-2">
                  <dt class="text-sm font-medium text-gray-500">Subject</dt>
                  <dd class="mt-1 text-sm text-gray-900">{{ notification.subject }}</dd>
                </div>

                <!-- Message -->
                <div class="sm:col-span-2">
                  <dt class="text-sm font-medium text-gray-500">Message</dt>
                  <dd class="mt-1 text-sm text-gray-900 whitespace-pre-wrap bg-gray-50 p-3 rounded border border-gray-200">{{ notification.message }}</dd>
                </div>

                <!-- Error Message (if failed) -->
                <div v-if="notification.error_message" class="sm:col-span-2">
                  <dt class="text-sm font-medium text-red-600">Error Details</dt>
                  <dd class="mt-1 text-sm text-red-700 bg-red-50 p-3 rounded border border-red-200 font-mono text-xs">{{ notification.error_message }}</dd>
                </div>

                <!-- Metadata (if present) -->
                <div v-if="notification.metadata && hasMetadata" class="sm:col-span-2">
                  <dt class="text-sm font-medium text-gray-500">Metadata</dt>
                  <dd class="mt-1 text-sm text-gray-900">
                    <div class="bg-gray-50 p-3 rounded border border-gray-200">
                      <pre class="text-xs font-mono">{{ formatMetadata(notification.metadata) }}</pre>
                    </div>
                  </dd>
                </div>
              </dl>
            </div>
          </div>
        </div>

        <!-- Modal Actions -->
        <div class="mt-5 sm:mt-6">
          <button
            @click="close"
            type="button"
            class="inline-flex justify-center w-full rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:text-sm"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch, computed } from 'vue'
import { getNotificationDetail } from '../services/api'
import StatusBadge from './StatusBadge.vue'

export default {
  name: 'NotificationDetail',
  components: {
    StatusBadge
  },
  props: {
    notificationId: {
      type: Number,
      default: null
    },
    isOpen: {
      type: Boolean,
      default: false
    }
  },
  emits: ['close'],
  setup(props, { emit }) {
    const notification = ref(null)
    const loading = ref(false)
    const error = ref(null)

    const hasMetadata = computed(() => {
      if (!notification.value?.metadata) return false
      try {
        const parsed = typeof notification.value.metadata === 'string' 
          ? JSON.parse(notification.value.metadata) 
          : notification.value.metadata
        return parsed && Object.keys(parsed).length > 0
      } catch {
        return false
      }
    })

    const loadNotification = async (id) => {
      if (!id) return

      loading.value = true
      error.value = null
      notification.value = null

      try {
        notification.value = await getNotificationDetail(id)
      } catch (err) {
        error.value = err.message || 'Failed to load notification details'
        console.error('Error loading notification detail:', err)
      } finally {
        loading.value = false
      }
    }

    const close = () => {
      emit('close')
    }

    const formatDate = (dateString) => {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString()
    }

    const formatMetadata = (metadata) => {
      try {
        const parsed = typeof metadata === 'string' ? JSON.parse(metadata) : metadata
        return JSON.stringify(parsed, null, 2)
      } catch {
        return metadata
      }
    }

    watch(() => props.notificationId, (newId) => {
      if (newId && props.isOpen) {
        loadNotification(newId)
      }
    }, { immediate: true })

    watch(() => props.isOpen, (isOpen) => {
      if (isOpen && props.notificationId) {
        loadNotification(props.notificationId)
      }
    })

    return {
      notification,
      loading,
      error,
      hasMetadata,
      close,
      formatDate,
      formatMetadata
    }
  }
}
</script>

<style scoped>
</style>
