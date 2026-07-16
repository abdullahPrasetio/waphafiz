<template>
  <Teleport to="body">
    <Transition name="toast">
      <div
        v-if="visible"
        class="fixed bottom-24 md:bottom-6 right-4 z-[200] px-4 py-3 rounded-lg shadow-lg text-sm font-medium flex items-center gap-2 max-w-xs"
        :class="typeClass"
      >
        <IconCheck v-if="type === 'success'" :size="16" />
        <IconX v-else :size="16" />
        {{ message }}
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { IconCheck, IconX } from '@tabler/icons-vue'

const props = defineProps<{
  message: string
  type?: 'success' | 'error'
  visible: boolean
}>()

const typeClass = computed(() =>
  props.type === 'error'
    ? 'bg-red-50 text-red-700 border border-red-200'
    : 'bg-green-50 text-green-700 border border-green-200'
)
</script>

<style scoped>
.toast-enter-active, .toast-leave-active { transition: all 0.2s ease; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateY(8px); }
</style>
