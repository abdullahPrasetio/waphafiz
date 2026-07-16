<template>
  <div class="p-5 md:p-8">
    <div class="mb-6">
      <h1 class="text-xl font-semibold">Riwayat Muraja'ah</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">Rekap sesi muraja'ah yang telah diselesaikan</p>
    </div>

    <!-- Filter tipe -->
    <div class="flex gap-2 mb-5">
      <button
        v-for="opt in filterOpts"
        :key="opt.value"
        class="px-3 py-1.5 rounded-full text-[12.5px] font-medium border transition-colors"
        :class="filter === opt.value
          ? 'bg-green-500 text-white border-green-500'
          : 'bg-white text-gray-500 border-gray-200 hover:border-gray-300'"
        @click="filter = opt.value"
      >
        {{ opt.label }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 5" :key="i" class="bg-white border border-gray-200 rounded-[14px] p-4 animate-pulse">
        <div class="h-3 bg-gray-100 rounded w-1/3 mb-2"></div>
        <div class="h-2 bg-gray-100 rounded w-1/2"></div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="filtered.length === 0" class="text-center py-16 text-gray-400">
      <IconCalendarOff class="mx-auto mb-3 text-gray-300" :size="40" />
      <p class="text-sm font-medium text-gray-500">Belum ada riwayat muraja'ah</p>
      <p class="text-xs mt-1">Selesaikan muraja'ah hari ini, lalu riwayat akan muncul di sini.</p>
      <NuxtLink to="/murajaah" class="inline-block mt-4 text-xs text-green-500 hover:underline">
        Lihat jadwal hari ini →
      </NuxtLink>
    </div>

    <!-- List berdasarkan tanggal -->
    <div v-else class="space-y-6">
      <div v-for="(group, date) in groupedByDate" :key="date">
        <div class="text-xs font-semibold text-gray-400 uppercase tracking-widest mb-2">{{ formatDate(date) }}</div>
        <div class="bg-white border border-gray-200 rounded-[14px] overflow-hidden">
          <div
            v-for="(item, idx) in group"
            :key="item.id"
            class="flex items-center gap-3 px-4 py-3"
            :class="idx < group.length - 1 ? 'border-b border-gray-50' : ''"
          >
            <div class="w-8 h-8 rounded-full bg-green-50 flex items-center justify-center flex-shrink-0">
              <IconCheck class="text-green-500" :size="15" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-[13px] font-medium text-gray-900 truncate">{{ surahName(item.surah_number) }}</div>
              <div class="text-[11px] text-gray-400">Ayat {{ item.ayat_start }}–{{ item.ayat_end }}</div>
            </div>
            <AppBadge :variant="item.type === 'sabqi' ? 'green' : 'blue'">
              {{ item.type === 'sabqi' ? 'Sabqi' : 'Manzil' }}
            </AppBadge>
            <div class="text-[11px] text-gray-400 flex-shrink-0">{{ formatTime(item.completed_at) }}</div>
          </div>
        </div>
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="text-center">
        <button
          class="text-[13px] text-green-500 hover:underline"
          :disabled="loadingMore"
          @click="loadMore"
        >
          {{ loadingMore ? 'Memuat...' : 'Muat lebih banyak' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { IconCalendarOff, IconCheck } from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

const quranStore = useQuranStore()
function surahName(num: number) {
  return quranStore.surahList.find((s) => s.number === num)?.englishName || `Surah ${num}`
}

interface HistoryItem {
  id: string
  type: 'sabqi' | 'manzil'
  surah_number: number
  ayat_start: number
  ayat_end: number
  scheduled_date: string
  completed_at: string
}

interface HistoryResponse {
  data: HistoryItem[]
  pagination: { page: number; per_page: number; total: number; total_pages: number }
}

const { apiFetch } = useApi()

const history = ref<HistoryItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const page = ref(1)
const limit = 20
const total = ref(0)
const filter = ref<'all' | 'sabqi' | 'manzil'>('all')

const filterOpts = [
  { value: 'all', label: 'Semua' },
  { value: 'sabqi', label: 'Sabqi' },
  { value: 'manzil', label: 'Manzil' },
]

const hasMore = computed(() => history.value.length < total.value)
const filtered = computed(() =>
  filter.value === 'all' ? history.value : history.value.filter((h) => h.type === filter.value)
)

const groupedByDate = computed(() => {
  const groups: Record<string, HistoryItem[]> = {}
  for (const item of filtered.value) {
    const date = item.completed_at.split('T')[0]
    if (!groups[date]) groups[date] = []
    groups[date].push(item)
  }
  return groups
})

async function fetchHistory(reset = false) {
  if (reset) {
    page.value = 1
    loading.value = true
  } else {
    loadingMore.value = true
  }
  try {
    const res = await apiFetch<HistoryResponse>(`/murajaah/history?page=${page.value}&limit=${limit}`)
    if (reset) {
      history.value = res.data
    } else {
      history.value.push(...res.data)
    }
    total.value = res.pagination.total
  } catch {
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function loadMore() {
  page.value++
  await fetchHistory(false)
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
}

function formatTime(isoStr: string) {
  return new Date(isoStr).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

watch(filter, () => fetchHistory(true))

onMounted(() => {
  fetchHistory(true)
  quranStore.fetchSurahList()
})
</script>
