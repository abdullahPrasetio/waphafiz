<template>
  <div class="p-5 md:p-8">
    <div class="mb-6">
      <h1 class="text-xl font-semibold">Dashboard Admin</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">{{ today }}</p>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
        <div v-for="i in 3" :key="i" class="bg-white border border-gray-200 rounded-[14px] p-4 animate-pulse h-24"></div>
      </div>
    </div>

    <template v-else>
      <!-- Stats -->
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
        <div class="bg-white border border-gray-200 rounded-[14px] p-4">
          <IconUsers class="text-green-500 mb-2" :size="19" />
          <div class="text-xs text-gray-500 mb-1">Total Anggota</div>
          <div class="text-[22px] font-semibold">{{ data?.total_members || 0 }}</div>
        </div>
        <div class="bg-white border border-gray-200 rounded-[14px] p-4">
          <IconActivity class="text-green-500 mb-2" :size="19" />
          <div class="text-xs text-gray-500 mb-1">Aktif Hari Ini</div>
          <div class="text-[22px] font-semibold">{{ data?.active_today || 0 }}</div>
        </div>
        <div class="bg-white border border-gray-200 rounded-[14px] p-4">
          <IconBook2 class="text-green-500 mb-2" :size="19" />
          <div class="text-xs text-gray-500 mb-1">Total Ayat Hafal</div>
          <div class="text-[22px] font-semibold">{{ totalAyatAll }}</div>
        </div>
      </div>

      <!-- Ranking + Tidak aktif -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- Ranking hafalan -->
        <div class="card">
          <div class="flex items-center justify-between mb-4">
            <span class="text-sm font-medium">Ranking Hafalan</span>
            <NuxtLink to="/admin/progress" class="text-xs text-green-500">Detail →</NuxtLink>
          </div>
          <div v-if="!ranked.length" class="text-xs text-gray-400 text-center py-6">Belum ada data.</div>
          <div v-else class="space-y-2">
            <div
              v-for="(m, idx) in ranked"
              :key="m.user_id"
              class="flex items-center gap-3 py-2 border-b border-gray-50 last:border-0"
            >
              <div
                class="w-6 h-6 rounded-full flex items-center justify-center text-[11px] font-bold flex-shrink-0"
                :class="idx === 0
                  ? 'bg-amber-400 text-white'
                  : idx === 1
                    ? 'bg-gray-300 text-white'
                    : idx === 2
                      ? 'bg-amber-700 text-white'
                      : 'bg-gray-100 text-gray-500'"
              >
                {{ idx + 1 }}
              </div>
              <AppAvatar :name="m.user_name" size="xs" />
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-medium truncate">{{ m.user_name }}</div>
                <div class="text-[11px] text-gray-400">{{ m.hafal_ayat }} ayat hafal</div>
              </div>
              <div
                class="w-2 h-2 rounded-full flex-shrink-0"
                :class="m.is_active ? 'bg-green-400' : 'bg-gray-200'"
              ></div>
            </div>
          </div>
        </div>

        <!-- Anggota tidak aktif -->
        <div class="card">
          <div class="flex items-center justify-between mb-4">
            <span class="text-sm font-medium">Tidak Aktif 7 Hari</span>
            <NuxtLink to="/admin/members" class="text-xs text-green-500">Kelola →</NuxtLink>
          </div>
          <div v-if="inactive.length === 0" class="flex flex-col items-center py-6 text-center">
            <IconCircleCheck class="text-green-400 mb-2" :size="28" />
            <p class="text-[13px] font-medium text-gray-600">Semua anggota aktif!</p>
            <p class="text-xs text-gray-400 mt-0.5">Tidak ada anggota yang absen lebih dari 7 hari.</p>
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="m in inactive"
              :key="m.user_id"
              class="flex items-center gap-3 py-2 border-b border-gray-50 last:border-0"
            >
              <AppAvatar :name="m.user_name" size="xs" />
              <div class="flex-1 min-w-0">
                <div class="text-[13px] font-medium truncate">{{ m.user_name }}</div>
                <div class="text-[11px] text-gray-400">
                  {{ m.last_seen ? `Terakhir: ${formatDate(m.last_seen)}` : 'Belum pernah aktif' }}
                </div>
              </div>
              <AppBadge variant="gray">Absen</AppBadge>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { IconUsers, IconActivity, IconBook2, IconCircleCheck } from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

interface MemberProgress {
  user_id: string
  user_name: string
  hafal_ayat: number
  is_active: boolean
  last_seen: string | null
}

interface AdminDashboardData {
  total_members: number
  active_today: number
  member_progress: MemberProgress[]
}

const { apiFetch } = useApi()
const loading = ref(false)
const data = ref<AdminDashboardData | null>(null)

const today = new Date().toLocaleDateString('id-ID', {
  weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
})

const ranked = computed(() =>
  [...(data.value?.member_progress || [])].sort((a, b) => b.hafal_ayat - a.hafal_ayat)
)

const totalAyatAll = computed(() =>
  (data.value?.member_progress || []).reduce((sum, m) => sum + m.hafal_ayat, 0)
)

const inactive = computed(() => {
  const cutoff = Date.now() - 7 * 24 * 60 * 60 * 1000
  return (data.value?.member_progress || []).filter((m) => {
    if (!m.last_seen) return true
    return new Date(m.last_seen).getTime() < cutoff
  })
})

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
}

async function fetchDashboard() {
  loading.value = true
  try {
    const res = await apiFetch<{ data: AdminDashboardData }>('/admin/dashboard')
    data.value = res.data
  } catch {
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchDashboard())
</script>
