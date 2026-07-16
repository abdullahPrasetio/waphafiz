<template>
  <div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" @click.self="$emit('close')">
    <div class="bg-white rounded-[14px] border border-gray-200 p-6 w-full max-w-sm max-h-[90vh] overflow-y-auto">
      <h2 class="text-[15px] font-semibold mb-1">Tambah hafalan baru</h2>
      <p class="text-[12.5px] text-gray-500 mb-5">Pilih surah dan rentang ayat yang sudah dihafal</p>

      <form @submit.prevent="handleSave">
        <div class="mb-4">
          <label class="form-label">Surah</label>
          <select v-model="form.surahNumber" class="w-full py-2.5 px-3 border border-gray-200 rounded-[10px] text-[13.5px] bg-white focus:outline-none focus:border-green-500" required>
            <option value="">Pilih surah...</option>
            <option v-for="s in quranStore.surahList" :key="s.number" :value="s.number">
              {{ s.number }}. {{ s.englishName }}
            </option>
          </select>
        </div>

        <!-- Ayat yang sudah dihafal pada surah ini -->
        <div v-if="existingRanges.length" class="mb-4 bg-amber-50 border border-amber-200 rounded-[10px] px-3 py-2.5">
          <p class="text-[11px] font-medium text-amber-700 mb-1.5">Sudah tercatat di surah ini:</p>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="r in existingRanges"
              :key="r.id"
              class="text-[11px] px-2 py-0.5 rounded-full font-medium"
              :class="r.status === 'hafal' ? 'bg-green-100 text-green-700' : 'bg-amber-100 text-amber-700'"
            >
              {{ r.ayat_start }}–{{ r.ayat_end }}
              <span class="opacity-70">({{ r.status }})</span>
            </span>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-2.5 mb-1">
          <div>
            <label class="form-label">Ayat mulai</label>
            <input v-model.number="form.ayatStart" type="number" class="form-input" style="padding-left:12px" min="1" :max="selectedSurah?.numberOfAyahs || 999" required />
          </div>
          <div>
            <label class="form-label">Ayat selesai</label>
            <input v-model.number="form.ayatEnd" type="number" class="form-input" style="padding-left:12px" :min="form.ayatStart || 1" :max="selectedSurah?.numberOfAyahs || 999" required />
          </div>
        </div>
        <p v-if="selectedSurah" class="text-[11px] text-gray-400 mb-3">Total ayat surah ini: {{ selectedSurah.numberOfAyahs }}</p>

        <p v-if="validationError" class="text-red-500 text-xs mb-3">{{ validationError }}</p>

        <div class="mb-5">
          <label class="form-label">Status</label>
          <select v-model="form.status" class="w-full py-2.5 px-3 border border-gray-200 rounded-[10px] text-[13.5px] bg-white focus:outline-none focus:border-green-500">
            <option value="hafal">Hafal</option>
            <option value="sedang">Sedang dihafal</option>
          </select>
        </div>

        <p v-if="error" class="text-red-500 text-xs mb-3">{{ error }}</p>

        <div class="flex gap-2">
          <button type="button" class="btn-outline flex-1" @click="$emit('close')">Batal</button>
          <button type="submit" class="btn-primary flex-1" :disabled="loading || !!validationError">
            {{ loading ? 'Menyimpan...' : 'Simpan' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const emit = defineEmits<{ close: []; saved: [] }>()

const quranStore = useQuranStore()
const hafalanStore = useHafalanStore()
const { apiFetch } = useApi()

const form = reactive({ surahNumber: '' as number | '', ayatStart: 1, ayatEnd: 1, status: 'hafal' })
const loading = ref(false)
const error = ref('')

const selectedSurah = computed(() =>
  form.surahNumber ? quranStore.surahList.find((s) => s.number === form.surahNumber) : null
)

const existingRanges = computed(() =>
  form.surahNumber ? hafalanStore.bySurah(form.surahNumber as number) : []
)

const validationError = computed(() => {
  if (!form.ayatStart || !form.ayatEnd) return ''
  if (form.ayatStart > form.ayatEnd)
    return 'Ayat mulai tidak boleh lebih besar dari ayat selesai'
  if (selectedSurah.value && form.ayatEnd > selectedSurah.value.numberOfAyahs)
    return `Surah ini hanya punya ${selectedSurah.value.numberOfAyahs} ayat`
  const overlap = existingRanges.value.find(
    (r) => form.ayatStart <= r.ayat_end && form.ayatEnd >= r.ayat_start
  )
  if (overlap)
    return `Ayat ${overlap.ayat_start}–${overlap.ayat_end} sudah tercatat (${overlap.status})`
  return ''
})

async function handleSave() {
  if (validationError.value) return
  loading.value = true
  error.value = ''
  try {
    await apiFetch('/hafalan', {
      method: 'POST',
      body: {
        surah_number: form.surahNumber,
        ayat_start: form.ayatStart,
        ayat_end: form.ayatEnd,
        status: form.status,
      },
    })
    emit('saved')
  } catch (e: any) {
    error.value = e?.data?.message || 'Gagal menyimpan hafalan'
  } finally {
    loading.value = false
  }
}

onMounted(() => quranStore.fetchSurahList())
</script>
