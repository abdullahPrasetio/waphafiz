<template>
  <div class="bg-gray-100 rounded-full overflow-hidden" :style="{ height: `${height}px` }">
    <div
      class="rounded-full transition-all duration-500"
      :class="colorClass"
      :style="{ width: `${Math.min(percent, 100)}%`, height: `${height}px` }"
    />
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  percent: number
  height?: number
  color?: 'green' | 'amber'
}>()

const colorClass = computed(() =>
  props.color === 'amber' ? 'bg-[#EF9F27]' : 'bg-green-500'
)
</script>
