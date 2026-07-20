<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header publik -->
    <header class="bg-white border-b border-gray-200 px-5 py-3.5 flex items-center gap-2.5">
      <div class="w-8 h-8 rounded-[10px] bg-green-500 flex items-center justify-center text-white">
        <IconBook2 :size="17" />
      </div>
      <div>
        <div class="text-[13px] font-semibold leading-tight">WAPHafiz</div>
        <div class="text-[11px] text-gray-400 leading-tight">Pantauan progress hafalan</div>
      </div>
    </header>

    <main class="max-w-2xl mx-auto p-5 md:p-8">
      <!-- Loading -->
      <div v-if="pending" class="py-16 text-center text-sm text-gray-400">Memuat...</div>

      <!-- Error / tidak valid -->
      <div v-else-if="error || !dash" class="py-16 text-center">
        <IconLinkOff class="mx-auto text-gray-300 mb-3" :size="40" />
        <div class="text-[15px] font-medium text-gray-700 mb-1">Link tidak ditemukan atau sudah tidak berlaku</div>
        <p class="text-[13px] text-gray-400">Minta link pantau baru kepada keluarga yang membagikannya.</p>
      </div>

      <!-- Dashboard read-only -->
      <template v-else>
        <div class="mb-5">
          <h1 class="text-xl font-semibold">Progress Hafalan {{ dash.member_name }}</h1>
          <p class="text-[12.5px] text-gray-500 mt-0.5">
            Dibagikan untuk: {{ dash.label }} · diperbarui {{ formatDateTime(dash.generated_at) }}
          </p>
        </div>

        <!-- Stats -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-5">
          <div class="bg-white border border-gray-200 rounded-[14px] p-4">
            <div class="text-xs text-gray-500 mb-1">Ayat dihafal</div>
            <div class="text-[22px] font-semibold">{{ dash.stats.total_hafal_ayat }}</div>
          </div>
          <div class="bg-white border border-gray-200 rounded-[14px] p-4">
            <div class="text-xs text-gray-500 mb-1">Surah selesai</div>
            <div class="text-[22px] font-semibold">{{ dash.stats.total_hafal_surah }}</div>
          </div>
          <div class="bg-white border border-gray-200 rounded-[14px] p-4">
            <div class="text-xs text-gray-500 mb-1">Persentase</div>
            <div class="text-[22px] font-semibold">{{ dash.stats.percent_hafal.toFixed(0) }}%</div>
            <div class="text-[11px] text-gray-400 mt-0.5">dari yang dicatat</div>
          </div>
          <div class="bg-white border border-gray-200 rounded-[14px] p-4">
            <div class="text-xs text-gray-500 mb-1">Streak muraja'ah</div>
            <div class="text-[22px] font-semibold">{{ dash.stats.streak }} <span class="text-[13px] font-normal text-gray-400">hari</span></div>
          </div>
        </div>

        <!-- Muraja'ah hari ini -->
        <div class="bg-white border border-gray-200 rounded-[14px] p-4 mb-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-medium">Muraja'ah Hari Ini</span>
            <span class="text-[13px] font-semibold" :class="murajaahAllDone ? 'text-green-500' : 'text-amber-600'">
              {{ dash.murajaah_today.done }} / {{ dash.murajaah_today.total }}
            </span>
          </div>
          <AppProgressBar :percent="murajaahPercent" :height="6" />
          <p class="text-[11.5px] text-gray-400 mt-2">
            {{ dash.murajaah_today.total === 0 ? 'Belum ada jadwal muraja\'ah hari ini.'
              : murajaahAllDone ? 'Alhamdulillah, seluruh muraja\'ah hari ini selesai.'
              : `${dash.murajaah_today.total - dash.murajaah_today.done} sesi belum diselesaikan.` }}
          </p>
        </div>

        <!-- Progress per surah -->
        <div class="bg-white border border-gray-200 rounded-[14px] p-4 mb-4">
          <div class="text-sm font-medium mb-3">Progress per Surah</div>
          <div v-if="dash.surah_progress.length === 0" class="text-xs text-gray-400">Belum ada hafalan tercatat.</div>
          <div v-else>
            <div v-for="s in dash.surah_progress" :key="s.surah_number" class="mb-3 last:mb-0">
              <div class="flex justify-between text-xs mb-1">
                <span class="font-medium">{{ s.surah_name || `Surah ${s.surah_number}` }}</span>
                <span class="text-gray-500">
                  {{ s.ayat_hafal }}<template v-if="s.ayat_total"> / {{ s.ayat_total }}</template> ayat
                  <AppBadge class="ml-1.5" :variant="s.status === 'hafal' ? 'green' : 'blue'">{{ s.status }}</AppBadge>
                </span>
              </div>
              <AppProgressBar :percent="s.ayat_total ? Math.min(100, s.ayat_hafal / s.ayat_total * 100) : 0" :height="5" />
            </div>
          </div>
        </div>

        <!-- Aktivitas terakhir -->
        <div class="bg-white border border-gray-200 rounded-[14px] p-4">
          <div class="text-sm font-medium mb-3">Aktivitas Terakhir</div>
          <div v-if="dash.last_activity.length === 0" class="text-xs text-gray-400">Belum ada aktivitas.</div>
          <div v-else>
            <div v-for="(a, i) in dash.last_activity" :key="i"
              class="flex items-center gap-3 py-2 border-b border-gray-50 last:border-0">
              <div class="text-[11px] text-gray-400 w-16 flex-shrink-0">{{ formatDate(a.date) }}</div>
              <div class="flex-1 text-[13px]">{{ a.detail }}</div>
              <AppBadge :variant="a.status === 'hafal' ? 'green' : a.status === 'sedang' ? 'blue' : 'gray'">{{ a.status }}</AppBadge>
            </div>
          </div>
        </div>

        <p class="text-center text-[11px] text-gray-300 mt-6">
          Halaman ini hanya-baca dan dibagikan lewat link pribadi · WAPHafiz
        </p>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { IconBook2, IconLinkOff } from '@tabler/icons-vue'

// Halaman publik: tanpa layout aplikasi, tanpa auth middleware.
definePageMeta({ layout: false })

interface PublicDashboard {
  member_name: string
  label: string
  stats: { total_hafal_ayat: number; total_hafal_surah: number; percent_hafal: number; streak: number }
  murajaah_today: { done: number; total: number }
  surah_progress: { surah_number: number; surah_name: string; ayat_hafal: number; ayat_total: number; status: string }[]
  last_activity: { date: string; type: string; detail: string; status: string }[]
  generated_at: string
}

const route = useRoute()
const config = useRuntimeConfig()

useHead({
  title: 'Progress Hafalan — WAPHafiz',
  meta: [{ name: 'robots', content: 'noindex, nofollow' }],
})

// $fetch polos (bukan useApi): endpoint publik, tanpa Authorization header,
// dan 404 di sini tidak boleh memicu redirect ke /login.
const { data, pending, error } = await useAsyncData(
  `share-${route.params.token}`,
  () => $fetch<{ data: PublicDashboard }>(`/public/shares/${route.params.token}`, {
    baseURL: config.public.apiBase as string,
  }),
)

const dash = computed(() => data.value?.data ?? null)

const murajaahPercent = computed(() => {
  const t = dash.value?.murajaah_today
  return t && t.total > 0 ? (t.done / t.total) * 100 : 0
})
const murajaahAllDone = computed(() => {
  const t = dash.value?.murajaah_today
  return !!t && t.total > 0 && t.done >= t.total
})

useHead(() => ({
  title: dash.value ? `Progress Hafalan ${dash.value.member_name} — WAPHafiz` : 'Progress Hafalan — WAPHafiz',
}))

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
}
function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}
</script>
