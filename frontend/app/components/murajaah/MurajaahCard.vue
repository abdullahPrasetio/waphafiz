<template>
  <div
    class="bg-white border border-gray-200 rounded-[14px] px-5 py-4 mb-2 flex items-center gap-3.5 transition-opacity"
    :class="{ 'opacity-60': item.completed_at }"
  >
    <div
      class="w-10 h-10 rounded-[10px] flex items-center justify-center text-[19px] flex-shrink-0"
      :class="item.type === 'sabqi' ? 'bg-green-50 text-green-500' : 'bg-blue-50 text-blue-700'"
    >
      <IconRefresh v-if="item.type === 'sabqi'" :size="19" />
      <IconBook2 v-else :size="19" />
    </div>

    <div class="flex-1 min-w-0">
      <AppBadge :variant="item.type === 'sabqi' ? 'green' : 'blue'">
        {{ item.type === 'sabqi' ? 'Sabqi' : 'Manzil' }}
      </AppBadge>
      <div class="text-sm font-medium mt-1">{{ surahName }}</div>
      <div class="text-xs text-gray-500">
        Ayat {{ item.ayat_start }}–{{ item.ayat_end }}
        <span v-if="isPlaying" class="ml-1 text-green-500">· Ayat {{ currentAyat }}</span>
      </div>
    </div>

    <div class="flex items-center gap-2.5">
      <button
        class="w-8 h-8 rounded-full border flex items-center justify-center transition-all"
        :class="isPlaying ? 'border-green-400 text-green-500 bg-green-50' : 'border-gray-200 text-gray-400 hover:text-gray-600'"
        :disabled="isLoading"
        @click="togglePlay"
      >
        <IconLoader2 v-if="isLoading" :size="15" class="animate-spin" />
        <IconPlayerPause v-else-if="isPlaying" :size="15" />
        <IconPlayerPlay v-else :size="15" />
      </button>
      <button
        class="w-8 h-8 rounded-full border flex items-center justify-center text-[15px] transition-all"
        :class="item.completed_at ? 'bg-green-500 border-green-500 text-white' : 'border-gray-200 text-gray-400'"
        :disabled="!!item.completed_at"
        @click="$emit('complete')"
      >
        <IconCheck :size="15" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { IconRefresh, IconBook2, IconPlayerPlay, IconPlayerPause, IconCheck, IconLoader2 } from '@tabler/icons-vue'
import type { MurajaahSchedule } from '~/stores/murajaah'

const props = defineProps<{ item: MurajaahSchedule }>()
defineEmits<{ complete: [] }>()

const { apiFetch } = useApi()
const quranStore = useQuranStore()

const surahName = computed(() => {
  const s = quranStore.surahList.find((s) => s.number === props.item.surah_number)
  return s?.englishName || `Surah ${props.item.surah_number}`
})

const audio = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const isLoading = ref(false)
const currentAyat = ref(props.item.ayat_start)

function stop() {
  audio.value?.pause()
  audio.value = null
  isPlaying.value = false
  isLoading.value = false
  currentAyat.value = props.item.ayat_start
}

async function playAyat(ayat: number) {
  if (!isPlaying.value) return
  isLoading.value = true
  try {
    const res = await apiFetch<{ data: { audio: string } }>(
      `/quran/ayah/${props.item.surah_number}/${ayat}/audio`
    )
    if (!isPlaying.value) return
    isLoading.value = false
    audio.value = new Audio(res.data.audio)
    audio.value.onended = () => {
      if (!isPlaying.value) return
      const next = ayat + 1 > props.item.ayat_end ? props.item.ayat_start : ayat + 1
      currentAyat.value = next
      playAyat(next)
    }
    audio.value.play()
  } catch {
    stop()
  }
}

function togglePlay() {
  if (isPlaying.value) {
    stop()
    return
  }
  isPlaying.value = true
  currentAyat.value = props.item.ayat_start
  playAyat(props.item.ayat_start)
}

onUnmounted(() => stop())
</script>
