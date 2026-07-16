<template>
  <div
    class="bg-white border rounded-[14px] px-5 py-4 transition-colors"
    :class="isPlaying ? 'border-green-400' : 'border-gray-200'"
  >
    <div class="flex items-center justify-between mb-2.5">
      <div class="w-7 h-7 rounded-full bg-green-50 flex items-center justify-center text-xs font-medium text-green-700">
        {{ ayah.number }}
      </div>
      <div class="flex items-center gap-2.5">
        <button
          class="text-gray-400 hover:text-green-500 transition-colors bg-none border-none"
          :class="{ 'text-green-500': isPlaying }"
          @click="$emit('play')"
        >
          <IconPlayerPause v-if="isPlaying" :size="16" />
          <IconPlayerPlay v-else :size="16" />
        </button>
        <button class="text-gray-400 hover:text-green-500 transition-colors bg-none border-none text-xs font-medium" @click="$emit('markHafal')">
          + Hafal
        </button>
      </div>
    </div>

    <div class="font-arabic text-right text-2xl leading-loose text-gray-900 mb-2">{{ ayah.text }}</div>
    <div class="text-[13px] text-gray-500 leading-relaxed">{{ ayah.translation }}</div>
  </div>
</template>

<script setup lang="ts">
import { IconPlayerPlay, IconPlayerPause } from '@tabler/icons-vue'
import type { Ayah } from '~/stores/quran'

defineProps<{ ayah: Ayah; isPlaying: boolean }>()
defineEmits<{ play: []; markHafal: [] }>()
</script>
