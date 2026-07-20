<template>
  <div class="p-5 md:p-8">
    <!-- Header -->
    <div class="mb-6 flex items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold">Assalamu'alaikum, {{ authStore.user?.name?.split(' ')[0] }}</h1>
        <p class="text-[13px] text-gray-500 mt-0.5">{{ today }}</p>
        <div v-if="streak > 0" class="inline-flex items-center gap-1.5 bg-amber-50 text-amber-900 text-xs font-medium px-3 py-1 rounded-full mt-2">
          <IconFlame :size="13" /> {{ streak }} hari streak
        </div>
      </div>
      <button
        class="flex items-center gap-1.5 text-[12.5px] font-medium text-green-600 bg-white border border-gray-200 hover:border-green-300 rounded-[10px] px-3 py-2 flex-shrink-0"
        @click="shareModalOpen = true">
        <IconShare :size="15" /> <span class="hidden md:inline">Link Pantau</span>
      </button>
    </div>

    <ShareManagerModal :open="shareModalOpen" :member-name="authStore.user?.name" @close="shareModalOpen = false" />

    <!-- Reminder muraja'ah -->
    <div v-if="!murajaahStore.allDone && murajaahStore.totalCount > 0" class="bg-green-50 border border-green-200 rounded-[10px] px-4 py-3 flex items-center gap-2.5 mb-6 text-[13px] text-green-900">
      <IconBell class="text-green-500 flex-shrink-0" :size="17" />
      Kamu belum menyelesaikan muraja'ah hari ini. Ada {{ murajaahStore.totalCount - murajaahStore.completedCount }} sesi yang menunggu.
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <IconBook2 class="text-green-500 mb-2" :size="19" />
        <div class="text-xs text-gray-500 mb-1">Total Hafalan</div>
        <div class="text-[22px] font-semibold">{{ hafalanStore.summary?.totalJuz || 0 }} Juz</div>
        <div class="text-[11px] text-gray-400 mt-0.5">{{ hafalanStore.summary?.totalAyat || 0 }} ayat</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <IconRefresh class="text-green-500 mb-2" :size="19" />
        <div class="text-xs text-gray-500 mb-1">Muraja'ah Hari Ini</div>
        <div class="text-[22px] font-semibold">{{ murajaahStore.completedCount }} / {{ murajaahStore.totalCount }}</div>
        <div class="text-[11px] text-gray-400 mt-0.5">{{ murajaahStore.totalCount - murajaahStore.completedCount }} sesi tersisa</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <IconUsers class="text-green-500 mb-2" :size="19" />
        <div class="text-xs text-gray-500 mb-1">Anggota Aktif</div>
        <div class="text-[22px] font-semibold">-</div>
        <div class="text-[11px] text-gray-400 mt-0.5">hari ini</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <IconCalendarCheck class="text-green-500 mb-2" :size="19" />
        <div class="text-xs text-gray-500 mb-1">Hafalan Bulan Ini</div>
        <div class="text-[22px] font-semibold">-</div>
        <div class="text-[11px] text-gray-400 mt-0.5">ayat</div>
      </div>
    </div>

    <!-- Grid muraja'ah + progress -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium">Muraja'ah Hari Ini</span>
          <NuxtLink to="/murajaah" class="text-xs text-green-500">Lihat semua →</NuxtLink>
        </div>
        <div v-if="murajaahStore.loading" class="text-xs text-gray-400">Memuat...</div>
        <div v-else-if="murajaahStore.today.length === 0" class="text-xs text-gray-400">Belum ada jadwal muraja'ah.</div>
        <div v-else>
          <div v-for="item in murajaahStore.today.slice(0, 3)" :key="item.id" class="flex items-center gap-2.5 py-2 border-b border-gray-50 last:border-0">
            <AppBadge :variant="item.type === 'sabqi' ? 'green' : 'blue'">{{ item.type === 'sabqi' ? 'Sabqi' : 'Manzil' }}</AppBadge>
            <div class="flex-1">
              <div class="text-[13px] font-medium">{{ surahMeta(item.surah_number)?.englishName || `Surah ${item.surah_number}` }}</div>
              <div class="text-[11px] text-gray-500">Ayat {{ item.ayat_start }}–{{ item.ayat_end }}</div>
            </div>
            <div class="w-6 h-6 rounded-full border flex items-center justify-center text-[13px] flex-shrink-0"
              :class="item.completed_at ? 'bg-green-500 border-green-500 text-white' : 'border-gray-200 text-gray-400'">
              <IconCheck :size="13" />
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium">Progress Hafalan Saya</span>
          <NuxtLink to="/hafalan" class="text-xs text-green-500">Detail →</NuxtLink>
        </div>
        <div v-if="hafalanStore.loading" class="text-xs text-gray-400">Memuat...</div>
        <div v-else-if="hafalanStore.list.length === 0" class="text-xs text-gray-400">
          Belum ada hafalan tercatat. <NuxtLink to="/hafalan" class="text-green-500">Mulai catat →</NuxtLink>
        </div>
        <div v-else>
          <div v-for="item in hafalanBySurah" :key="item.surahNumber" class="mb-3">
            <div class="flex justify-between text-xs mb-1">
              <span class="font-medium">{{ surahMeta(item.surahNumber)?.englishName || `Surah ${item.surahNumber}` }}</span>
              <span class="text-gray-500">{{ Math.min(100, Math.round(item.recordedAyat / (surahMeta(item.surahNumber)?.numberOfAyahs || 1) * 100)) }}%</span>
            </div>
            <AppProgressBar :percent="Math.min(100, item.recordedAyat / (surahMeta(item.surahNumber)?.numberOfAyahs || 1) * 100)" :height="5" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  IconFlame, IconBell, IconBook2, IconRefresh,
  IconUsers, IconCalendarCheck, IconCheck, IconShare,
} from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

const authStore = useAuthStore()
const hafalanStore = useHafalanStore()
const murajaahStore = useMurajaahStore()
const quranStore = useQuranStore()

function surahMeta(surahNumber: number) {
  return quranStore.surahList.find((s) => s.number === surahNumber)
}

// Group hafalan by surah, sum recorded ayat per surah
const hafalanBySurah = computed(() => {
  const map = new Map<number, { surahNumber: number; recordedAyat: number }>()
  for (const h of hafalanStore.list) {
    const count = h.ayat_end - h.ayat_start + 1
    const existing = map.get(h.surah_number)
    if (existing) {
      existing.recordedAyat += count
    } else {
      map.set(h.surah_number, { surahNumber: h.surah_number, recordedAyat: count })
    }
  }
  return [...map.values()].slice(0, 4)
})

const streak = ref(0)
const shareModalOpen = ref(false)
const today = new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })

onMounted(async () => {
  await Promise.all([
    hafalanStore.fetchList(),
    hafalanStore.fetchSummary(),
    murajaahStore.fetchToday(),
    quranStore.fetchSurahList(),
  ])
})
</script>
