<template>
  <div class="p-5 md:p-8">
    <div class="mb-5">
      <h1 class="text-xl font-semibold">Manajemen Anggota</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">Kelola anggota keluarga dan pantau aktivitas mereka</p>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-5">
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <div class="text-xs text-gray-500 mb-1">Total anggota</div>
        <div class="text-[22px] font-semibold">{{ members.length }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <div class="text-xs text-gray-500 mb-1">Aktif hari ini</div>
        <div class="text-[22px] font-semibold">{{ activeCount }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <div class="text-xs text-gray-500 mb-1">Tidak aktif 7 hari</div>
        <div class="text-[22px] font-semibold">{{ inactiveCount }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[14px] p-4">
        <div class="text-xs text-gray-500 mb-1">Rata-rata hafalan</div>
        <div class="text-[22px] font-semibold">-</div>
      </div>
    </div>

    <!-- Toolbar -->
    <div class="flex flex-col md:flex-row gap-2.5 mb-4">
      <div class="relative flex-1">
        <IconSearch class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="15" />
        <input v-model="search" type="text" class="w-full pl-8 pr-3 py-2 text-[13px] border border-gray-200 rounded-[10px] bg-white" placeholder="Cari nama atau email..." />
      </div>
      <div class="flex items-center gap-2 bg-white border border-gray-200 rounded-[10px] px-3 py-2 text-[12.5px] text-gray-500">
        <IconKey class="text-green-500" :size="15" />
        Kode undangan: <span class="font-semibold text-green-500 tracking-wide">{{ inviteCode }}</span>
        <button class="text-gray-400 hover:text-gray-600 border-none bg-none" @click="copyCode"><IconCopy :size="15" /></button>
        <button class="text-gray-400 hover:text-gray-600 border-none bg-none" @click="regenerateCode"><IconRefresh :size="15" /></button>
      </div>
    </div>

    <!-- Tabel -->
    <div class="bg-white border border-gray-200 rounded-[14px] overflow-hidden">
      <div v-if="loading" class="p-6 text-sm text-gray-400">Memuat...</div>
      <table v-else class="w-full text-[13px] border-collapse">
        <thead>
          <tr class="bg-gray-50 border-b border-gray-200">
            <th class="text-left px-4 py-2.5 text-[11px] font-medium text-gray-500 uppercase tracking-wide">Anggota</th>
            <th class="text-left px-4 py-2.5 text-[11px] font-medium text-gray-500 uppercase tracking-wide">Role</th>
            <th class="text-left px-4 py-2.5 text-[11px] font-medium text-gray-500 uppercase tracking-wide">Status</th>
            <th class="text-left px-4 py-2.5 text-[11px] font-medium text-gray-500 uppercase tracking-wide">Hafalan</th>
            <th class="text-left px-4 py-2.5 text-[11px] font-medium text-gray-500 uppercase tracking-wide">Terakhir aktif</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in filteredMembers" :key="m.id" class="border-b border-gray-50 hover:bg-gray-50 last:border-0">
            <td class="px-4 py-3">
              <div class="flex items-center gap-2.5">
                <AppAvatar :name="m.name" size="sm" />
                <div>
                  <div class="font-medium text-gray-900">{{ m.name }}</div>
                  <div class="text-[11px] text-gray-400">{{ m.email }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-3">
              <AppBadge :variant="m.role === 'admin' ? 'purple' : 'gray'">
                {{ m.role === 'admin' ? 'Admin' : 'Anggota' }}
              </AppBadge>
            </td>
            <td class="px-4 py-3">
              <AppBadge :variant="m.is_active ? 'green' : 'gray'">
                {{ m.is_active ? 'Aktif' : 'Tidak aktif' }}
              </AppBadge>
            </td>
            <td class="px-4 py-3 text-gray-500">-</td>
            <td class="px-4 py-3 text-[12px] text-gray-500">{{ m.last_seen || '-' }}</td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2 justify-end">
                <button class="text-gray-400 hover:text-green-500 border-none bg-none" title="Link pantau"
                  @click="openShareModal(m)">
                  <IconShare :size="15" />
                </button>
                <button class="text-gray-400 hover:text-gray-600 border-none bg-none"><IconDots :size="15" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <ShareManagerModal
      :open="shareModalOpen"
      :user-id="shareTarget?.id"
      :member-name="shareTarget?.name"
      @close="shareModalOpen = false" />
  </div>
</template>

<script setup lang="ts">
import { IconSearch, IconKey, IconCopy, IconRefresh, IconDots, IconShare } from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

interface Member {
  id: string
  name: string
  email: string
  role: 'admin' | 'member'
  is_active: boolean
  last_seen?: string
}

const { apiFetch } = useApi()
const members = ref<Member[]>([])
const loading = ref(false)
const search = ref('')
const inviteCode = ref('FAM-2026-????')
const shareModalOpen = ref(false)
const shareTarget = ref<Member | null>(null)

function openShareModal(m: Member) {
  shareTarget.value = m
  shareModalOpen.value = true
}

const activeCount = computed(() => members.value.filter((m) => m.is_active).length)
const inactiveCount = computed(() => members.value.filter((m) => !m.is_active).length)
const filteredMembers = computed(() => {
  if (!search.value) return members.value
  const q = search.value.toLowerCase()
  return members.value.filter((m) => m.name.toLowerCase().includes(q) || m.email.toLowerCase().includes(q))
})

async function fetchMembers() {
  loading.value = true
  try {
    const res = await apiFetch<{ data: { members: Member[]; invite_code: string } }>('/family/members')
    members.value = res.data.members
    inviteCode.value = res.data.invite_code
  } catch {} finally {
    loading.value = false
  }
}

async function copyCode() {
  await navigator.clipboard.writeText(inviteCode.value)
}

async function regenerateCode() {
  const ok = await useConfirm().confirm({
    title: 'Generate ulang kode undangan?',
    message: 'Kode lama tidak bisa dipakai lagi.',
    confirmText: 'Generate ulang',
  })
  if (!ok) return
  try {
    const res = await apiFetch<{ data: { invite_code: string } }>('/family/invite-code/regenerate', { method: 'POST' })
    inviteCode.value = res.data.invite_code
  } catch {}
}

onMounted(() => fetchMembers())
</script>
