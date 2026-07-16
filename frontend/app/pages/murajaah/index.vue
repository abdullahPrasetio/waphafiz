<template>
  <div class="p-5 md:p-8">
    <div class="mb-5 flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold">Muraja'ah</h1>
        <p class="text-[13px] text-gray-500 mt-0.5">Jadwal pengulangan hafalan hari ini, {{ today }}</p>
      </div>
      <NuxtLink to="/murajaah/history" class="text-xs text-green-500 hover:underline">Riwayat →</NuxtLink>
    </div>

    <!-- Summary chips -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-2.5 mb-6">
      <div class="bg-white border border-gray-200 rounded-[10px] px-3.5 py-3">
        <div class="text-xl font-semibold">{{ store.completedCount }} / {{ store.totalCount }}</div>
        <div class="text-[11px] text-gray-500 mt-0.5">Selesai hari ini</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[10px] px-3.5 py-3">
        <div class="text-xl font-semibold">{{ store.sabqi.length }}</div>
        <div class="text-[11px] text-gray-500 mt-0.5">Sesi Sabqi</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[10px] px-3.5 py-3">
        <div class="text-xl font-semibold">{{ store.manzil.length }}</div>
        <div class="text-[11px] text-gray-500 mt-0.5">Sesi Manzil</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-[10px] px-3.5 py-3">
        <div class="text-xl font-semibold">{{ totalAyat }}</div>
        <div class="text-[11px] text-gray-500 mt-0.5">Total hari ini</div>
      </div>
    </div>

    <div v-if="store.loading" class="text-sm text-gray-400">Memuat jadwal...</div>

    <div v-else-if="store.error" class="bg-red-50 border border-red-200 rounded-[10px] px-4 py-3 text-sm text-red-600 mb-4">
      {{ store.error }}
    </div>

    <div v-else-if="store.today.length === 0" class="text-center py-12">
      <p class="text-sm text-gray-400 mb-3">Belum ada jadwal muraja'ah.</p>
      <p class="text-xs text-gray-400">Catat hafalan terlebih dahulu agar jadwal terbuat otomatis.</p>
      <NuxtLink to="/hafalan" class="inline-block mt-3 text-sm text-green-500 font-medium">Catat hafalan →</NuxtLink>
    </div>

    <template v-else>
      <div v-if="store.sabqi.length" class="mb-5">
        <div class="text-[11px] text-gray-400 uppercase tracking-wider mb-2.5">Sabqi — hafalan 7 hari terakhir</div>
        <MurajaahCard v-for="item in store.sabqi" :key="item.id" :item="item" @complete="store.markComplete(item.id)" />
      </div>

      <div v-if="store.manzil.length">
        <div class="text-[11px] text-gray-400 uppercase tracking-wider mb-2.5">Manzil — siklus hafalan lama</div>
        <MurajaahCard v-for="item in store.manzil" :key="item.id" :item="item" @complete="store.markComplete(item.id)" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })

const store = useMurajaahStore()
const today = ref('')
onMounted(() => {
  today.value = new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
})

const totalAyat = computed(() =>
  store.today.reduce((sum, s) => sum + (s.ayat_end - s.ayat_start + 1), 0)
)

onMounted(() => store.fetchToday())
</script>
