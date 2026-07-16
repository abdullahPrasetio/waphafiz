<template>
  <div
    class="rounded-full flex items-center justify-center font-medium flex-shrink-0"
    :class="[colorClass, sizeClass]"
  >
    {{ initials }}
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  name: string
  size?: 'xs' | 'sm' | 'md' | 'lg'
  color?: 'green' | 'purple' | 'amber' | 'blue'
}>()

const initials = computed(() =>
  props.name
    .split(' ')
    .slice(0, 2)
    .map((w) => w[0])
    .join('')
    .toUpperCase()
)

const colorMap = {
  green:  'bg-green-50 text-green-600',
  purple: 'bg-purple-50 text-purple-800',
  amber:  'bg-amber-50 text-amber-900',
  blue:   'bg-blue-50 text-blue-700',
}
const colorClass = computed(() => colorMap[props.color || 'green'])

const sizeMap = {
  xs: 'w-6 h-6 text-[10px]',
  sm: 'w-8 h-8 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-12 h-12 text-base',
}
const sizeClass = computed(() => sizeMap[props.size || 'md'])
</script>
