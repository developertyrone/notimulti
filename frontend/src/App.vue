<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow">
      <div class="container mx-auto px-4 py-6">
        <h1 class="text-2xl font-bold text-gray-900">
          Notification Server Dashboard
        </h1>
      </div>
    </header>
    
    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
// No imports needed for basic layout
</script>
