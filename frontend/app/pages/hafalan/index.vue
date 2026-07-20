<template>
  <div class="p-5 md:p-8">
    <div class="mb-5">
      <h1 class="text-xl font-semibold">Hafalan Saya</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">Catat dan pantau progress hafalan per surah</p>
    </div>

    <!-- Toolbar -->
    <div class="flex flex-col md:flex-row items-stretch md:items-center gap-2.5 mb-5">
      <div class="relative flex-1">
        <IconSearch class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="15" />
        <input v-model="search" type="text" class="w-full pl-8 pr-3 py-2 text-[13px] border border-gray-200 rounded-[10px] bg-white" placeholder="Cari surah..." />
      </div>
      <div class="flex gap-2">
        <button v-for="f in filters" :key="f.value"
          class="px-3.5 py-2 border rounded-[10px] text-[12.5px] transition-all"
          :class="activeFilter === f.value ? 'border-green-500 bg-green-50 text-green-700 font-medium' : 'border-gray-200 text-gray-500 bg-white'"
          @click="activeFilter = f.value"
        >{{ f.label }}</button>
        <button class="btn-primary flex items-center gap-1.5 text-sm" @click="showModal = true">
          <IconPlus :size="15" /> Tambah
        </button>
      </div>
    </div>

    <!-- List -->
    <div v-if="hafalanStore.loading" class="text-sm text-gray-400">Memuat...</div>
    <div v-else-if="filtered.length === 0" class="text-sm text-gray-400 py-8 text-center">
      Tidak ada hafalan ditemukan.
    </div>
    <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
      <div
        v-for="h in filtered" :key="h.id"
        class="bg-white border rounded-[14px] p-3 transition-colors"
        :class="{
          'border-green-400': h.status === 'hafal',
          'border-yellow-300': h.status === 'sedang',
          'border-gray-200': h.status === 'belum',
        }"
      >
        <!-- Header: nomor + badge + link ke surah -->
        <div class="flex justify-between items-start mb-1.5">
          <span class="text-[11px] text-gray-400">{{ h.surah_number }}</span>
          <div class="flex items-center gap-1.5">
            <AppBadge :variant="statusVariant(h.status)">{{ statusLabel(h.status) }}</AppBadge>
            <NuxtLink :to="`/quran/${h.surah_number}`" class="text-gray-300 hover:text-green-500 transition-colors" title="Buka surah">
              <IconExternalLink :size="12" />
            </NuxtLink>
          </div>
        </div>

        <div class="text-sm font-medium mb-1">{{ surahMeta(h.surah_number)?.englishName || '-' }}</div>
        <div class="font-arabic text-right text-[15px] text-gray-500 mb-2">{{ surahMeta(h.surah_number)?.name }}</div>
        <div v-if="h.status !== 'belum'" class="text-[12px] font-medium text-gray-700 mb-1">
          Ayat {{ h.ayat_start }}–{{ h.ayat_end }}
        </div>
        <AppProgressBar
          :percent="h.status === 'belum' ? 0 : (h.ayat_end - h.ayat_start + 1) / (surahMeta(h.surah_number)?.numberOfAyahs || 1) * 100"
          :height="3"
          :color="h.status === 'sedang' ? 'amber' : 'green'"
        />

        <!-- Baris bawah: ayat count + play -->
        <div class="flex items-center justify-between mt-1.5">
          <div class="text-[11px] text-gray-400">
            <template v-if="h.status === 'belum'">{{ surahMeta(h.surah_number)?.numberOfAyahs || '-' }} ayat</template>
            <template v-else>{{ h.ayat_end - h.ayat_start + 1 }} / {{ surahMeta(h.surah_number)?.numberOfAyahs || '-' }} ayat</template>
          </div>
          <button
            v-if="h.status !== 'belum'"
            type="button"
            class="flex items-center gap-1 text-[11px] font-medium rounded-full px-2 py-1"
            :class="playingId === h.id ? 'text-green-600 bg-green-50' : 'text-gray-400 hover:text-green-600'"
            :disabled="loadingId === h.id"
            @click="toggleRepeat(h)"
          >
            <IconLoader2 v-if="loadingId === h.id" :size="13" class="animate-spin" />
            <IconPlayerPause v-else-if="playingId === h.id" :size="13" />
            <IconRepeat v-else :size="13" />
            {{ loadingId === h.id ? '' : playingId === h.id ? `Ayat ${currentAyat}` : 'Dengar' }}
          </button>
        </div>

        <!-- Aksi edit & hapus (hanya record yang sudah ada) -->
        <div v-if="h.status !== 'belum'" class="flex items-center gap-1 mt-2 pt-2 border-t border-gray-100">
          <button
            v-for="s in STATUS_OPTIONS"
            :key="s"
            type="button"
            class="flex-1 text-[10px] py-0.5 rounded-[6px] border transition-colors"
            :class="h.status === s
              ? s === 'hafal' ? 'bg-green-50 border-green-300 text-green-700 font-medium'
              : s === 'sedang' ? 'bg-amber-50 border-amber-300 text-amber-700 font-medium'
              : 'bg-gray-100 border-gray-300 text-gray-600 font-medium'
              : 'border-gray-100 text-gray-400 hover:border-gray-200'"
            :disabled="updatingId === h.id"
            @click="onUpdateStatus(h, s)"
          >
            <IconLoader2 v-if="updatingId === h.id && h.status !== s" :size="10" class="animate-spin mx-auto" />
            <template v-else>{{ s === 'hafal' ? 'Hafal' : s === 'sedang' ? 'Sedang' : 'Belum' }}</template>
          </button>
          <button
            type="button"
            class="w-6 h-6 flex items-center justify-center text-red-300 hover:text-red-500 hover:bg-red-50 rounded-[6px] transition-colors ml-0.5"
            :disabled="deletingId === h.id"
            @click="deleteHafalan(h)"
          >
            <IconLoader2 v-if="deletingId === h.id" :size="11" class="animate-spin" />
            <IconTrash v-else :size="11" />
          </button>
        </div>
      </div>
    </div>

    <!-- Error audio toast -->
    <Transition name="fade">
      <div v-if="audioError" class="fixed bottom-20 left-1/2 -translate-x-1/2 bg-red-500 text-white text-[13px] px-4 py-2 rounded-[10px] shadow-lg z-50">
        {{ audioError }}
      </div>
    </Transition>

    <!-- Modal Tambah -->
    <HafalanModal v-if="showModal" @close="showModal = false" @saved="onSaved" />
  </div>
</template>

<script setup lang="ts">
import { IconSearch, IconPlus, IconRepeat, IconPlayerPause, IconLoader2, IconTrash, IconExternalLink } from '@tabler/icons-vue'

const STATUS_OPTIONS = ['hafal', 'sedang', 'belum'] as const
type HafalanStatus = typeof STATUS_OPTIONS[number]
import type { HafalanProgress } from '~/stores/hafalan'
import { SURAH_NAMES_ID } from '~/utils/surahNames'

definePageMeta({ middleware: 'auth' })

const hafalanStore = useHafalanStore()
const quranStore = useQuranStore()
const { apiFetch } = useApi()
const showModal = ref(false)
const search = ref('')
const activeFilter = ref('semua')

const audio = ref<HTMLAudioElement | null>(null)
const playingId = ref<string | null>(null)
const currentAyat = ref<number | null>(null)
const loadingId = ref<string | null>(null)
const audioError = ref<string | null>(null)
const updatingId = ref<string | null>(null)
const deletingId = ref<string | null>(null)

async function onUpdateStatus(h: HafalanProgress, newStatus: string) {
  if (newStatus === h.status) return
  updatingId.value = h.id
  try {
    await apiFetch(`/hafalan/${h.id}`, { method: 'PATCH', body: { status: newStatus } })
    await hafalanStore.fetchList()
  } catch {
    audioError.value = 'Gagal memperbarui status'
    setTimeout(() => { audioError.value = null }, 3000)
  } finally {
    updatingId.value = null
  }
}

async function deleteHafalan(h: HafalanProgress) {
  const ok = await useConfirm().confirm({
    title: `Hapus hafalan Surah ${h.surah_number}?`,
    message: `Ayat ${h.ayat_start}–${h.ayat_end} akan dihapus permanen.`,
    confirmText: 'Hapus',
    variant: 'danger',
  })
  if (!ok) return
  if (playingId.value === h.id) stopAudio()
  deletingId.value = h.id
  try {
    await apiFetch(`/hafalan/${h.id}`, { method: 'DELETE' })
    await hafalanStore.fetchList()
  } catch {
    audioError.value = 'Gagal menghapus hafalan'
    setTimeout(() => { audioError.value = null }, 3000)
  } finally {
    deletingId.value = null
  }
}

function stopAudio() {
  audio.value?.pause()
  audio.value = null
  playingId.value = null
  currentAyat.value = null
  loadingId.value = null
}

async function playCurrent(h: HafalanProgress) {
  if (playingId.value !== h.id || currentAyat.value === null) return
  loadingId.value = h.id
  try {
    const res = await apiFetch<{ data: { audio: string } }>(
      `/quran/ayah/${h.surah_number}/${currentAyat.value}/audio`,
    )
    if (playingId.value !== h.id) return
    loadingId.value = null
    audio.value = new Audio(res.data.audio)
    audio.value.onended = () => {
      if (playingId.value !== h.id || currentAyat.value === null) return
      currentAyat.value = currentAyat.value + 1 > h.ayat_end ? h.ayat_start : currentAyat.value + 1
      playCurrent(h)
    }
    audio.value.play()
  } catch {
    stopAudio()
    audioError.value = 'Audio tidak tersedia saat ini'
    setTimeout(() => { audioError.value = null }, 3000)
  }
}

function toggleRepeat(h: HafalanProgress) {
  if (playingId.value === h.id) {
    stopAudio()
    return
  }
  stopAudio()
  playingId.value = h.id
  currentAyat.value = h.ayat_start
  playCurrent(h)
}

onUnmounted(() => stopAudio())

const filters = [
  { label: 'Semua', value: 'semua' },
  { label: 'Hafal', value: 'hafal' },
  { label: 'Sedang', value: 'sedang' },
  { label: 'Belum', value: 'belum' },
]

function surahMeta(surahNumber: number) {
  return quranStore.surahList.find((s) => s.number === surahNumber)
}

// Surah numbers yang sudah punya record hafalan
const recordedSurahNumbers = computed(() => new Set(hafalanStore.list.map((h) => h.surah_number)))

// Untuk filter "belum": surah dari 114 list yang belum ada recordnya
const belumItems = computed(() =>
  quranStore.surahList
    .filter((s) => !recordedSurahNumbers.value.has(s.number))
    .map((s) => ({
      id: `belum-${s.number}`,
      surah_number: s.number,
      ayat_start: 1,
      ayat_end: s.numberOfAyahs,
      status: 'belum' as const,
      noted_at: '',
    }))
)

const filtered = computed(() => {
  let list: typeof hafalanStore.list
  if (activeFilter.value === 'belum') {
    list = belumItems.value
  } else if (activeFilter.value === 'semua') {
    list = [...hafalanStore.list, ...belumItems.value]
  } else {
    list = hafalanStore.list.filter((h) => h.status === activeFilter.value)
  }
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter((h) => {
      const meta = surahMeta(h.surah_number)
      return (
        meta?.englishName.toLowerCase().includes(q) ||
        (SURAH_NAMES_ID[h.surah_number] || '').toLowerCase().includes(q) ||
        String(h.surah_number).includes(q)
      )
    })
  }
  return list
})

function statusVariant(status: string) {
  return { hafal: 'green', sedang: 'amber', belum: 'gray' }[status] as any
}
function statusLabel(status: string) {
  return { hafal: 'Hafal', sedang: 'Sedang', belum: 'Belum' }[status]
}

function onSaved() {
  showModal.value = false
  hafalanStore.fetchList()
}

onMounted(() => {
  hafalanStore.fetchList()
  quranStore.fetchSurahList()
})
</script>

<style>
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
