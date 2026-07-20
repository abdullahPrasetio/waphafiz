<template>
  <div v-if="open" class="fixed inset-0 z-50 flex items-end md:items-center justify-center bg-black/40 p-0 md:p-4" @click.self="close">
    <div class="bg-white w-full md:max-w-lg rounded-t-[20px] md:rounded-[14px] max-h-[90vh] overflow-y-auto">
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 sticky top-0 bg-white">
        <div>
          <div class="text-sm font-semibold">Link Pantau</div>
          <div class="text-[11px] text-gray-400">{{ memberName || 'Progress hafalan saya' }}</div>
        </div>
        <button class="text-gray-400 hover:text-gray-600 border-none bg-none" @click="close"><IconX :size="18" /></button>
      </div>

      <div class="p-5">
        <!-- Token baru: hanya ditampilkan sekali -->
        <div v-if="createdResult" class="bg-green-50 border border-green-200 rounded-[10px] p-4 mb-5">
          <div class="flex items-center gap-1.5 text-[13px] font-medium text-green-900 mb-2">
            <IconCircleCheck :size="15" class="text-green-500" /> Link berhasil dibuat
          </div>
          <div class="flex items-center gap-2 bg-white border border-gray-200 rounded-[10px] px-3 py-2 mb-2">
            <span class="text-[11.5px] text-gray-600 truncate flex-1">{{ createdUrl }}</span>
            <button class="text-green-500 hover:text-green-600 border-none bg-none flex-shrink-0" @click="copyUrl">
              <IconCopy v-if="!copied" :size="15" />
              <IconCheck v-else :size="15" />
            </button>
          </div>
          <div class="flex gap-2">
            <a :href="waLink" target="_blank" rel="noopener"
              class="flex-1 text-center text-[12px] font-medium text-white bg-green-500 hover:bg-green-600 rounded-[10px] py-2">
              Bagikan via WhatsApp
            </a>
          </div>
          <p class="text-[11px] text-amber-700 mt-2.5 flex items-start gap-1">
            <IconAlertTriangle :size="12" class="mt-0.5 flex-shrink-0" />
            Simpan link ini sekarang — demi keamanan, link tidak bisa ditampilkan lagi setelah jendela ini ditutup.
          </p>
        </div>

        <!-- Form buat link -->
        <form v-if="!createdResult" class="mb-5" @submit.prevent="submitCreate">
          <label class="block text-[12px] font-medium text-gray-700 mb-1">Label penerima</label>
          <input v-model="form.label" type="text" required maxlength="100"
            class="w-full px-3 py-2 text-[13px] border border-gray-200 rounded-[10px] mb-1"
            placeholder='Contoh: "Ustadz Ahmad", "Nenek"' />
          <p v-if="formError" class="text-[11px] text-red-500 mb-2">{{ formError }}</p>

          <label class="block text-[12px] font-medium text-gray-700 mb-1 mt-3">Masa berlaku</label>
          <select v-model.number="form.expires_in_days" class="w-full px-3 py-2 text-[13px] border border-gray-200 rounded-[10px] bg-white">
            <option :value="7">7 hari</option>
            <option :value="30">30 hari</option>
            <option :value="90">90 hari</option>
            <option :value="0">Tanpa batas waktu</option>
          </select>

          <button type="submit" :disabled="creating"
            class="w-full mt-4 py-2.5 text-[13px] font-medium text-white bg-green-500 hover:bg-green-600 disabled:opacity-50 rounded-[10px] border-none">
            {{ creating ? 'Membuat...' : 'Buat Link Pantau' }}
          </button>
        </form>
        <button v-else class="w-full mb-5 py-2 text-[12.5px] font-medium text-green-600 bg-green-50 hover:bg-green-100 rounded-[10px] border-none"
          @click="createdResult = null">
          + Buat link lain
        </button>

        <!-- Daftar link -->
        <div class="text-[12px] font-medium text-gray-700 mb-2">Link yang pernah dibuat</div>
        <div v-if="shareStore.loading" class="text-xs text-gray-400 py-3">Memuat...</div>
        <div v-else-if="shareStore.error" class="text-xs text-red-500 py-3">{{ shareStore.error }}</div>
        <div v-else-if="shareStore.list.length === 0" class="text-xs text-gray-400 py-3">
          Belum ada link pantau. Buat satu untuk dibagikan ke guru ngaji atau keluarga.
        </div>
        <div v-else>
          <div v-for="s in shareStore.list" :key="s.id"
            class="flex items-center gap-3 py-2.5 border-b border-gray-50 last:border-0">
            <div class="flex-1 min-w-0">
              <div class="text-[13px] font-medium truncate">{{ s.label }}</div>
              <div class="text-[11px] text-gray-400">
                {{ s.view_count }}× dilihat
                <template v-if="s.last_viewed_at"> · terakhir {{ formatDate(s.last_viewed_at) }}</template>
                <template v-if="s.expires_at"> · s.d. {{ formatDate(s.expires_at) }}</template>
              </div>
            </div>
            <AppBadge :variant="s.status === 'aktif' ? 'green' : 'gray'">{{ s.status }}</AppBadge>
            <button v-if="s.status === 'aktif'"
              class="text-[11.5px] text-red-500 hover:text-red-600 border-none bg-none flex-shrink-0"
              @click="confirmRevoke(s)">
              Cabut
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { IconX, IconCopy, IconCheck, IconCircleCheck, IconAlertTriangle } from '@tabler/icons-vue'
import type { ShareInfo, CreateShareResult } from '~/stores/share'

const props = defineProps<{
  open: boolean
  // kosong = kelola link milik sendiri; diisi = admin mengelola link anggota lain
  userId?: string
  memberName?: string
}>()
const emit = defineEmits<{ close: [] }>()

const shareStore = useShareStore()

const form = reactive({ label: '', expires_in_days: 90 })
const formError = ref<string | null>(null)
const creating = ref(false)
const createdResult = ref<CreateShareResult | null>(null)
const copied = ref(false)

const createdUrl = computed(() =>
  createdResult.value ? `${window.location.origin}/s/${createdResult.value.token}` : '')

const waLink = computed(() =>
  `https://wa.me/?text=${encodeURIComponent(`Assalamu'alaikum, silakan pantau progress hafalan ${props.memberName || 'saya'} lewat link ini: ${createdUrl.value}`)}`)

watch(() => props.open, (open) => {
  if (open) {
    createdResult.value = null
    formError.value = null
    form.label = ''
    form.expires_in_days = 90
    shareStore.fetchList(props.userId)
  }
})

function close() {
  emit('close')
}

async function submitCreate() {
  formError.value = null
  if (!form.label.trim()) {
    formError.value = 'Label wajib diisi'
    return
  }
  creating.value = true
  try {
    createdResult.value = await shareStore.create({
      user_id: props.userId || undefined,
      label: form.label.trim(),
      expires_in_days: form.expires_in_days,
    })
    copied.value = false
    await shareStore.fetchList(props.userId)
  } catch (e: any) {
    formError.value = e?.data?.message || 'Gagal membuat link pantau'
  } finally {
    creating.value = false
  }
}

async function copyUrl() {
  await navigator.clipboard.writeText(createdUrl.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

async function confirmRevoke(s: ShareInfo) {
  const ok = await useConfirm().confirm({
    title: `Cabut link "${s.label}"?`,
    message: 'Pemegang link tidak akan bisa melihat progress lagi.',
    confirmText: 'Cabut link',
    variant: 'danger',
  })
  if (!ok) return
  try {
    await shareStore.revoke(s.id)
    await shareStore.fetchList(props.userId)
  } catch {
    // biarkan list apa adanya; error berikutnya terlihat saat refresh list
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>
