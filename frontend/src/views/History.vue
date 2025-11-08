<template>
  <div class="history-view container mx-auto px-4 py-6">
    <!-- Page Header -->
    <div class="mb-6">
      <h1 class="text-3xl font-bold text-gray-900">Notification History</h1>
      <p class="mt-2 text-sm text-gray-600">
        View all sent notifications, delivery status, and error details
      </p>
    </div>

    <!-- Notification History Component -->
    <NotificationHistory @show-detail="showDetail" />

    <!-- Notification Detail Modal -->
    <NotificationDetail
      :notification-id="selectedNotificationId"
      :is-open="isDetailOpen"
      @close="closeDetail"
    />
  </div>
</template>

<script>
import { ref } from 'vue'
import NotificationHistory from '../components/NotificationHistory.vue'
import NotificationDetail from '../components/NotificationDetail.vue'

export default {
  name: 'History',
  components: {
    NotificationHistory,
    NotificationDetail
  },
  setup() {
    const selectedNotificationId = ref(null)
    const isDetailOpen = ref(false)

    const showDetail = (notificationId) => {
      selectedNotificationId.value = notificationId
      isDetailOpen.value = true
    }

    const closeDetail = () => {
      isDetailOpen.value = false
      // Delay clearing the ID to allow for exit animation
      setTimeout(() => {
        selectedNotificationId.value = null
      }, 300)
    }

    return {
      selectedNotificationId,
      isDetailOpen,
      showDetail,
      closeDetail
    }
  }
}
</script>

<style scoped>
.history-view {
  min-height: 100vh;
}
</style>
