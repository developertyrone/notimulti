<template>
  <div class="notification-history">
    <!-- Filters -->
    <div class="bg-white shadow rounded-lg p-4 mb-4">
      <h2 class="text-lg font-medium text-gray-900 mb-4">Filter Notifications</h2>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Provider Filter -->
        <div>
          <label for="provider-filter" class="block text-sm font-medium text-gray-700">Provider</label>
          <select
            id="provider-filter"
            v-model="filters.provider_id"
            @change="applyFilters"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          >
            <option value="">All Providers</option>
            <option v-for="provider in providers" :key="provider.id" :value="provider.id">
              {{ provider.id }} ({{ provider.type }})
            </option>
          </select>
        </div>

        <!-- Status Filter -->
        <div>
          <label for="status-filter" class="block text-sm font-medium text-gray-700">Status</label>
          <select
            id="status-filter"
            v-model="filters.status"
            @change="applyFilters"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          >
            <option value="">All Statuses</option>
            <option value="pending">Pending</option>
            <option value="sent">Sent</option>
            <option value="failed">Failed</option>
            <option value="retrying">Retrying</option>
          </select>
        </div>

        <!-- Date From -->
        <div>
          <label for="date-from" class="block text-sm font-medium text-gray-700">From Date</label>
          <input
            id="date-from"
            type="datetime-local"
            v-model="filters.date_from"
            @change="applyFilters"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          />
        </div>

        <!-- Date To -->
        <div>
          <label for="date-to" class="block text-sm font-medium text-gray-700">To Date</label>
          <input
            id="date-to"
            type="datetime-local"
            v-model="filters.date_to"
            @change="applyFilters"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          />
        </div>
      </div>

      <!-- Include Tests Checkbox -->
      <div class="mt-4">
        <label class="flex items-center">
          <input
            type="checkbox"
            v-model="filters.include_tests"
            @change="applyFilters"
            class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
          />
          <span class="ml-2 text-sm text-gray-700">Include test notifications</span>
        </label>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="bg-white shadow rounded-lg p-8">
      <div class="flex items-center justify-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-500"></div>
        <span class="ml-3 text-gray-600">Loading notifications...</span>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error loading notifications</h3>
          <p class="mt-1 text-sm text-red-700">{{ error }}</p>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="notifications.length === 0" class="bg-white shadow rounded-lg p-8">
      <div class="text-center">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No notifications found</h3>
        <p class="mt-1 text-sm text-gray-500">Try adjusting your filters or send some notifications first.</p>
      </div>
    </div>

    <!-- Notifications Table -->
    <div v-else class="bg-white shadow rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Recipient</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Message</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="notification in notifications" :key="notification.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
              {{ notification.provider_id }}
              <span v-if="notification.is_test" class="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                TEST
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ truncate(notification.recipient, 30) }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">
              {{ truncate(notification.message, 50) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <StatusBadge :status="notification.status" />
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(notification.created_at) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
              <button
                @click="showDetail(notification.id)"
                class="text-indigo-600 hover:text-indigo-900"
              >
                Details
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <Pagination
        :has-next="hasMore"
        :has-prev="currentCursor !== null"
        :page-size="pageSize"
        @next="loadNextPage"
        @prev="loadPrevPage"
        @update:page-size="updatePageSize"
      />
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { getNotificationHistory } from '../services/api'
import { fetchProviders } from '../services/api'
import StatusBadge from './StatusBadge.vue'
import Pagination from './Pagination.vue'

export default {
  name: 'NotificationHistory',
  components: {
    StatusBadge,
    Pagination
  },
  emits: ['show-detail'],
  setup(props, { emit }) {
    const notifications = ref([])
    const providers = ref([])
    const loading = ref(false)
    const error = ref(null)
    const hasMore = ref(false)
    const currentCursor = ref(null)
    const pageSize = ref(50)
    const cursorHistory = ref([])

    const filters = ref({
      provider_id: '',
      status: '',
      date_from: '',
      date_to: '',
      include_tests: true
    })

    const loadNotifications = async (cursor = null) => {
      loading.value = true
      error.value = null

      try {
        const apiFilters = {
          ...filters.value,
          cursor: cursor,
          page_size: pageSize.value
        }

        // Convert datetime-local to ISO8601
        if (apiFilters.date_from) {
          apiFilters.date_from = new Date(apiFilters.date_from).toISOString()
        }
        if (apiFilters.date_to) {
          apiFilters.date_to = new Date(apiFilters.date_to).toISOString()
        }

        const response = await getNotificationHistory(apiFilters)
        notifications.value = response.notifications || []
        hasMore.value = response.pagination.has_more
        currentCursor.value = cursor
      } catch (err) {
        error.value = err.message || 'Failed to load notifications'
        console.error('Error loading notifications:', err)
      } finally {
        loading.value = false
      }
    }

    const loadProviders = async () => {
      try {
        providers.value = await fetchProviders()
      } catch (err) {
        console.error('Error loading providers:', err)
      }
    }

    const applyFilters = () => {
      cursorHistory.value = []
      currentCursor.value = null
      loadNotifications()
    }

    const loadNextPage = () => {
      if (hasMore.value && notifications.value.length > 0) {
        const lastNotification = notifications.value[notifications.value.length - 1]
        if (currentCursor.value !== null) {
          cursorHistory.value.push(currentCursor.value)
        }
        loadNotifications(lastNotification.id)
      }
    }

    const loadPrevPage = () => {
      if (cursorHistory.value.length > 0) {
        const prevCursor = cursorHistory.value.pop()
        loadNotifications(prevCursor)
      } else {
        loadNotifications(null)
      }
    }

    const updatePageSize = (newSize) => {
      pageSize.value = newSize
      applyFilters()
    }

    const showDetail = (id) => {
      emit('show-detail', id)
    }

    const truncate = (text, length) => {
      if (!text) return ''
      return text.length > length ? text.substring(0, length) + '...' : text
    }

    const formatDate = (dateString) => {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString()
    }

    onMounted(() => {
      loadNotifications()
      loadProviders()
    })

    return {
      notifications,
      providers,
      loading,
      error,
      hasMore,
      currentCursor,
      pageSize,
      filters,
      applyFilters,
      loadNextPage,
      loadPrevPage,
      updatePageSize,
      showDetail,
      truncate,
      formatDate
    }
  }
}
</script>

<style scoped>
</style>
