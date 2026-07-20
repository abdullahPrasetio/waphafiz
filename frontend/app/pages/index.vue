<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
navigateTo('/dashboard')
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-emerald-500"></div>
  </div>
</template>
