<template>
  <div class="bg-white border border-gray-200 rounded-[14px] px-5 py-4 flex items-center gap-3.5 flex-wrap">
    <button
      class="w-9 h-9 rounded-full bg-green-500 flex items-center justify-center text-white flex-shrink-0 border-none cursor-pointer hover:bg-green-700 transition-colors"
      @click="$emit('toggle')"
    >
      <IconPlayerPause v-if="isPlaying" :size="17" />
      <IconPlayerPlay v-else :size="17" />
    </button>

    <div class="flex-shrink-0">
      <div class="text-[13px] font-medium">{{ surah.name }} · Ayat {{ currentAyah.number }}</div>
      <div class="text-[11px] text-gray-500">Mishary Rashid Alafasy</div>
    </div>

    <div class="flex-1 flex items-center gap-2 order-3 w-full md:w-auto">
      <span class="text-[11px] text-gray-400">{{ formatTime(currentTime) }}</span>
      <div class="flex-1 h-1 bg-gray-100 rounded-full overflow-hidden cursor-pointer" @click="seek">
        <div class="h-1 bg-green-500 rounded-full transition-all" :style="{ width: `${progress}%` }" />
      </div>
      <span class="text-[11px] text-gray-400">{{ formatTime(duration) }}</span>
    </div>

    <div class="flex gap-1.5 flex-shrink-0">
      <button class="p-1.5 text-gray-400 hover:text-gray-600 border-none bg-none" @click="$emit('prev')">
        <IconPlayerSkipBack :size="16" />
      </button>
      <button class="p-1.5 text-gray-400 hover:text-gray-600 border-none bg-none" @click="$emit('next')">
        <IconPlayerSkipForward :size="16" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { IconPlayerPlay, IconPlayerPause, IconPlayerSkipBack, IconPlayerSkipForward } from '@tabler/icons-vue'
import type { Surah, Ayah } from '~/stores/quran'

defineProps<{ surah: Surah; currentAyah: Ayah; isPlaying: boolean }>()
defineEmits<{ toggle: []; next: []; prev: [] }>()

const currentTime = ref(0)
const duration = ref(0)
const progress = computed(() => duration.value ? (currentTime.value / duration.value) * 100 : 0)

function formatTime(s: number) {
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${m}:${sec.toString().padStart(2, '0')}`
}

function seek(e: MouseEvent) {
  // controlled by parent via audio element
}
</script>
