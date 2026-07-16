<template>
  <div class="p-5 md:p-8">
    <div class="mb-6">
      <h1 class="text-xl font-semibold">Progress Hafalan Anggota</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">Pantau perkembangan hafalan seluruh anggota keluarga</p>
    </div>

    <!-- Search & filter -->
    <div class="flex flex-col md:flex-row gap-2.5 mb-5">
      <div class="relative flex-1">
        <IconSearch class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="15" />
        <input
          v-model="search"
          type="text"
          class="w-full pl-8 pr-3 py-2 text-[13px] border border-gray-200 rounded-[10px] bg-white"
          placeholder="Cari nama anggota..."
        />
      </div>
      <select
        v-model="sortBy"
        class="px-3 py-2 text-[13px] border border-gray-200 rounded-[10px] bg-white text-gray-600"
      >
        <option value="totalAyat">Urutkan: Total Ayat</option>
        <option value="name">Urutkan: Nama</option>
        <option value="lastActive">Urutkan: Terakhir Aktif</option>
      </select>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 4" :key="i" class="bg-white border border-gray-200 rounded-[14px] p-5 animate-pulse">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-9 h-9 rounded-full bg-gray-100"></div>
          <div class="flex-1">
            <div class="h-3 bg-gray-100 rounded w-1/3 mb-2"></div>
            <div class="h-2 bg-gray-100 rounded w-1/4"></div>
          </div>
        </div>
        <div class="h-2 bg-gray-100 rounded mb-2"></div>
        <div class="h-2 bg-gray-100 rounded w-3/4"></div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredMembers.length === 0" class="text-center py-16 text-gray-400">
      <IconUsers class="mx-auto mb-3 text-gray-300" :size="40" />
      <p class="text-sm font-medium text-gray-500">Tidak ada anggota ditemukan</p>
    </div>

    <!-- Cards anggota -->
    <div v-else class="space-y-3">
      <div
        v-for="m in filteredMembers"
        :key="m.user_id"
        class="bg-white border border-gray-200 rounded-[14px] p-5"
      >
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-3">
            <AppAvatar :name="m.name" size="sm" />
            <div>
              <div class="text-[14px] font-semibold text-gray-900">{{ m.name }}</div>
              <div class="text-[11px] text-gray-400 mt-0.5">{{ m.email }}</div>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="flex items-center gap-1 text-[12px] text-green-600 border border-green-200 bg-green-50 px-2.5 py-1 rounded-[8px] hover:bg-green-100 transition-colors"
              @click="openManage(m)"
            >
              <IconEdit :size="13" />
              Kelola
            </button>
            <div class="text-right">
              <div class="text-[11px] text-gray-400">
                {{ m.active_today ? 'Aktif hari ini' : `Terakhir: ${formatLastSeen(m.last_seen_at)}` }}
              </div>
              <div
                class="inline-block mt-1 w-2 h-2 rounded-full"
                :class="m.active_today ? 'bg-green-400' : 'bg-gray-300'"
              ></div>
            </div>
          </div>
        </div>

        <!-- Stats row -->
        <div class="grid grid-cols-3 gap-3 mb-4">
          <div class="text-center">
            <div class="text-[18px] font-bold text-gray-900">{{ m.total_ayat }}</div>
            <div class="text-[10px] text-gray-400">Ayat</div>
          </div>
          <div class="text-center border-x border-gray-100">
            <div class="text-[18px] font-bold text-gray-900">{{ m.total_surah }}</div>
            <div class="text-[10px] text-gray-400">Surah</div>
          </div>
          <div class="text-center">
            <div class="text-[18px] font-bold text-gray-900">{{ m.total_juz.toFixed(1) }}</div>
            <div class="text-[10px] text-gray-400">Juz</div>
          </div>
        </div>

        <!-- Progress per status -->
        <div class="flex items-center gap-2.5 text-[11px] text-gray-500">
          <span class="flex items-center gap-1">
            <span class="inline-block w-2 h-2 rounded-full bg-green-400"></span>
            Hafal: {{ m.status_counts.hafal }}
          </span>
          <span class="flex items-center gap-1">
            <span class="inline-block w-2 h-2 rounded-full bg-amber-400"></span>
            Sedang: {{ m.status_counts.sedang }}
          </span>
          <span class="flex items-center gap-1">
            <span class="inline-block w-2 h-2 rounded-full bg-gray-300"></span>
            Belum: {{ m.status_counts.belum }}
          </span>
        </div>

        <!-- Progress bar -->
        <div v-if="m.total_ayat > 0" class="mt-3">
          <div class="flex justify-between text-[11px] text-gray-400 mb-1">
            <span>Progress hafal</span>
            <span>{{ percentHafal(m) }}%</span>
          </div>
          <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
            <div
              class="h-full bg-green-400 rounded-full transition-all"
              :style="{ width: `${percentHafal(m)}%` }"
            ></div>
          </div>
        </div>

        <!-- Hafalan terbaru -->
        <div v-if="m.recent_hafalan?.length" class="mt-3 pt-3 border-t border-gray-50">
          <div class="text-[11px] text-gray-400 mb-1.5">Hafalan terbaru</div>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="h in m.recent_hafalan"
              :key="h.id"
              class="text-[11px] bg-gray-50 border border-gray-100 px-2 py-0.5 rounded-full text-gray-600"
            >
              Surah {{ h.surah_number }} {{ h.ayat_start }}–{{ h.ayat_end }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Kelola Hafalan -->
    <Transition name="fade">
      <div v-if="manageModal" class="fixed inset-0 bg-black/40 z-40 flex items-end md:items-center justify-center p-4" @click.self="closeManage">
        <div class="bg-white rounded-[20px] w-full max-w-md max-h-[80vh] flex flex-col">
          <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100">
            <div>
              <div class="text-[15px] font-semibold">Kelola Hafalan</div>
              <div class="text-[12px] text-gray-400">{{ manageModal.name }}</div>
            </div>
            <button class="text-gray-400 hover:text-gray-600" @click="closeManage">
              <IconX :size="18" />
            </button>
          </div>

          <div class="flex-1 overflow-y-auto px-5 py-4">
            <div v-if="manageLoading" class="text-center py-8 text-gray-400 text-sm">Memuat...</div>
            <div v-else-if="manageList.length === 0" class="text-center py-8 text-gray-400 text-sm">Belum ada hafalan tercatat.</div>
            <div v-else class="space-y-2">
              <div
                v-for="h in manageList"
                :key="h.id"
                class="flex items-center gap-3 border border-gray-100 rounded-[12px] px-3.5 py-3"
              >
                <div class="flex-1 min-w-0">
                  <div class="text-[13px] font-medium text-gray-800">Surah {{ h.surah_number }}</div>
                  <div class="text-[11px] text-gray-400">Ayat {{ h.ayat_start }}–{{ h.ayat_end }}</div>
                </div>
                <select
                  :value="h.status"
                  class="text-[12px] border border-gray-200 rounded-[8px] px-2 py-1 bg-white"
                  :disabled="savingId === h.id"
                  @change="onStatusChange(h, $event)"
                >
                  <option value="hafal">Hafal</option>
                  <option value="sedang">Sedang</option>
                  <option value="belum">Belum</option>
                </select>
                <button
                  class="w-7 h-7 flex items-center justify-center text-red-400 hover:text-red-600 hover:bg-red-50 rounded-[8px] transition-colors"
                  :disabled="savingId === h.id || deletingId === h.id"
                  @click="deleteHafalan(h)"
                >
                  <IconLoader2 v-if="deletingId === h.id" :size="14" class="animate-spin" />
                  <IconTrash v-else :size="14" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast" class="fixed bottom-20 left-1/2 -translate-x-1/2 text-white text-[13px] px-4 py-2 rounded-[10px] shadow-lg z-50"
           :class="toast.type === 'error' ? 'bg-red-500' : 'bg-green-500'">
        {{ toast.message }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { IconSearch, IconUsers, IconEdit, IconX, IconTrash, IconLoader2 } from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

interface RecentHafalan {
  id: string
  surah_number: number
  ayat_start: number
  ayat_end: number
}

interface MemberProgress {
  user_id: string
  name: string
  email: string
  total_ayat: number
  total_surah: number
  total_juz: number
  active_today: boolean
  last_seen_at: string | null
  status_counts: { hafal: number; sedang: number; belum: number }
  recent_hafalan: RecentHafalan[]
}

interface HafalanItem {
  id: string
  surah_number: number
  ayat_start: number
  ayat_end: number
  status: string
}

const { apiFetch } = useApi()
const members = ref<MemberProgress[]>([])
const loading = ref(false)
const search = ref('')
const sortBy = ref<'totalAyat' | 'name' | 'lastActive'>('totalAyat')

const manageModal = ref<MemberProgress | null>(null)
const manageList = ref<HafalanItem[]>([])
const manageLoading = ref(false)
const savingId = ref<string | null>(null)
const deletingId = ref<string | null>(null)
const toast = ref<{ message: string; type: 'success' | 'error' } | null>(null)

const filteredMembers = computed(() => {
  let list = members.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter((m) => m.name.toLowerCase().includes(q))
  }
  return [...list].sort((a, b) => {
    if (sortBy.value === 'totalAyat') return b.total_ayat - a.total_ayat
    if (sortBy.value === 'name') return a.name.localeCompare(b.name)
    if (sortBy.value === 'lastActive') {
      const ta = a.last_seen_at ? new Date(a.last_seen_at).getTime() : 0
      const tb = b.last_seen_at ? new Date(b.last_seen_at).getTime() : 0
      return tb - ta
    }
    return 0
  })
})

function percentHafal(m: MemberProgress) {
  if (!m.total_ayat) return 0
  return Math.round((m.status_counts.hafal / m.total_ayat) * 100)
}

function formatLastSeen(iso: string | null) {
  if (!iso) return 'Belum pernah aktif'
  return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

function showToast(message: string, type: 'success' | 'error' = 'success') {
  toast.value = { message, type }
  setTimeout(() => { toast.value = null }, 3000)
}

async function openManage(m: MemberProgress) {
  manageModal.value = m
  manageList.value = []
  manageLoading.value = true
  try {
    const res = await apiFetch<{ data: HafalanItem[] }>(`/admin/hafalan/member/${m.user_id}`)
    manageList.value = res.data ?? []
  } catch {
    showToast('Gagal memuat hafalan', 'error')
  } finally {
    manageLoading.value = false
  }
}

function closeManage() {
  manageModal.value = null
  manageList.value = []
  fetchProgress()
}

function onStatusChange(h: HafalanItem, event: Event) {
  const val = (event.target as HTMLSelectElement).value
  updateHafalan(h, val)
}

async function updateHafalan(h: HafalanItem, newStatus: string) {
  savingId.value = h.id
  try {
    await apiFetch(`/admin/hafalan/${h.id}`, {
      method: 'PATCH',
      body: { status: newStatus },
    })
    h.status = newStatus
    showToast('Status hafalan diperbarui')
  } catch {
    showToast('Gagal memperbarui status', 'error')
  } finally {
    savingId.value = null
  }
}

async function deleteHafalan(h: HafalanItem) {
  if (!confirm(`Hapus hafalan Surah ${h.surah_number} ayat ${h.ayat_start}–${h.ayat_end}?`)) return
  deletingId.value = h.id
  try {
    await apiFetch(`/admin/hafalan/${h.id}`, { method: 'DELETE' })
    manageList.value = manageList.value.filter((x) => x.id !== h.id)
    showToast('Hafalan dihapus')
  } catch {
    showToast('Gagal menghapus hafalan', 'error')
  } finally {
    deletingId.value = null
  }
}

async function fetchProgress() {
  loading.value = true
  try {
    const res = await apiFetch<{ data: MemberProgress[] }>('/admin/hafalan')
    members.value = res.data ?? []
  } catch {
    members.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchProgress())
</script>

<style>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
