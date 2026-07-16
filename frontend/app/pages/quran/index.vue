<template>
  <div class="p-5 md:p-8">
    <div class="mb-5">
      <h1 class="text-xl font-semibold">Al-Quran</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">114 surah</p>
    </div>

    <div class="relative mb-4">
      <IconSearch class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="15" />
      <input v-model="search" type="text" class="w-full pl-8 pr-3 py-2 text-[13px] border border-gray-200 rounded-[10px] bg-white" placeholder="Cari surah..." />
    </div>

    <div v-if="store.loading" class="text-sm text-gray-400">Memuat surah...</div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
      <NuxtLink
        v-for="s in filtered" :key="s.number"
        :to="`/quran/${s.number}`"
        class="bg-white border border-gray-200 rounded-[14px] px-4 py-3.5 flex items-center gap-3.5 hover:border-green-400 transition-colors"
      >
        <div class="w-9 h-9 rounded-full bg-green-50 flex items-center justify-center text-xs font-medium text-green-700 flex-shrink-0">
          {{ s.number }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium">{{ s.name }}</div>
          <div class="text-[11px] text-gray-400">{{ s.numberOfAyahs }} ayat · {{ s.revelationType === 'Meccan' ? 'Makkiyah' : 'Madaniyah' }}</div>
        </div>
        <div class="font-arabic text-[18px] text-gray-400 flex-shrink-0">{{ s.englishName }}</div>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { IconSearch } from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

const store = useQuranStore()
const search = ref('')

const filtered = computed(() => {
  if (!search.value) return store.surahList
  const q = search.value.toLowerCase()
  return store.surahList.filter((s) =>
    s.name.toLowerCase().includes(q) ||
    s.englishName.toLowerCase().includes(q) ||
    (SURAH_NAMES_ID[s.number] || '').toLowerCase().includes(q) ||
    String(s.number).includes(q)
  )
})

onMounted(() => store.fetchSurahList())
</script>
