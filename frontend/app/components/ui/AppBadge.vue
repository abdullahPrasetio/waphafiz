<template>
  <span class="badge" :class="variantClass">
    <slot />
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  variant?: 'green' | 'blue' | 'amber' | 'gray' | 'purple'
}>()

const variantMap = {
  green:  'bg-green-50 text-green-700',
  blue:   'bg-blue-50 text-blue-700',
  amber:  'bg-amber-50 text-amber-900',
  gray:   'bg-gray-100 text-gray-400',
  purple: 'bg-purple-50 text-purple-800',
}
const variantClass = computed(() => variantMap[props.variant || 'gray'])
</script>
