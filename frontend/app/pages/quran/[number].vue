<template>
  <div class="p-5 md:p-8">
    <NuxtLink to="/quran" class="inline-flex items-center gap-1.5 text-[13px] text-gray-500 mb-5 hover:text-gray-700">
      <IconArrowLeft :size="15" /> Kembali ke daftar surah
    </NuxtLink>

    <div v-if="store.loading" class="text-sm text-gray-400">Memuat surah...</div>

    <template v-else-if="store.currentSurah">
      <!-- Surah Header -->
      <div class="bg-white border border-gray-200 rounded-[14px] p-6 mb-4">
        <div class="flex justify-between items-start">
          <div>
            <h1 class="text-xl font-semibold">{{ store.currentSurah.name }}</h1>
            <p class="text-xs text-gray-500 mt-1">
              Surah ke-{{ store.currentSurah.number }} · {{ store.currentSurah.numberOfAyahs }} ayat ·
              {{ store.currentSurah.revelationType === 'Meccan' ? 'Makkiyah' : 'Madaniyah' }}
            </p>
          </div>
          <div class="font-arabic text-[32px] text-green-500">{{ store.currentSurah.englishName }}</div>
        </div>
      </div>

      <!-- Audio Player -->
      <QuranAudioPlayer
        v-if="currentAyah"
        :surah="store.currentSurah"
        :currentAyah="currentAyah"
        :isPlaying="isPlaying"
        @toggle="togglePlay"
        @next="nextAyah"
        @prev="prevAyah"
      />

      <!-- Ayah list -->
      <div class="flex flex-col gap-2 mt-4">
        <QuranAyahCard
          v-for="ayah in store.currentSurah.ayahs"
          :key="ayah.number"
          :ayah="ayah"
          :isPlaying="isPlaying && currentAyah?.number === ayah.number"
          @play="playAyah(ayah)"
          @markHafal="openHafalanModal(ayah)"
        />
      </div>
    </template>

    <div v-else-if="store.error" class="text-sm text-red-500">{{ store.error }}</div>

    <HafalanModal v-if="showModal" @close="showModal = false" @saved="showModal = false" />
  </div>
</template>

<script setup lang="ts">
import { IconArrowLeft } from '@tabler/icons-vue'
import type { Ayah } from '~/stores/quran'

definePageMeta({ middleware: 'auth' })

const route = useRoute()
const store = useQuranStore()

const isPlaying = ref(false)
const currentAyah = ref<Ayah | null>(null)
const showModal = ref(false)
const audio = ref<HTMLAudioElement | null>(null)

async function playAyah(ayah: Ayah) {
  if (currentAyah.value?.number === ayah.number) {
    togglePlay()
    return
  }
  currentAyah.value = ayah
  isPlaying.value = true
  if (!import.meta.client) return

  audio.value?.pause()
  try {
    const { apiFetch } = useApi()
    const res = await apiFetch<{ data: { audio: string } }>(
      `/quran/ayah/${route.params.number}/${ayah.numberInSurah}/audio`,
    )
    audio.value = new Audio(res.data.audio)
    audio.value.play()
    audio.value.onended = () => nextAyah()
  } catch (e: any) {
    isPlaying.value = false
    store.error = e?.data?.message || 'Audio tidak tersedia saat ini'
  }
}

function togglePlay() {
  isPlaying.value = !isPlaying.value
  isPlaying.value ? audio.value?.play() : audio.value?.pause()
}

function nextAyah() {
  if (!store.currentSurah || !currentAyah.value) return
  const idx = store.currentSurah.ayahs.findIndex((a) => a.number === currentAyah.value!.number)
  const next = store.currentSurah.ayahs[idx + 1]
  if (next) playAyah(next)
}

function prevAyah() {
  if (!store.currentSurah || !currentAyah.value) return
  const idx = store.currentSurah.ayahs.findIndex((a) => a.number === currentAyah.value!.number)
  const prev = store.currentSurah.ayahs[idx - 1]
  if (prev) playAyah(prev)
}

function openHafalanModal(_ayah: Ayah) {
  showModal.value = true
}

onMounted(() => store.fetchSurahDetail(Number(route.params.number)))
onUnmounted(() => { audio.value?.pause(); audio.value = null })
</script>
