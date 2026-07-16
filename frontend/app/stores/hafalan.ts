export interface HafalanProgress {
  id: string
  surah_number: number
  ayat_start: number
  ayat_end: number
  status: 'belum' | 'sedang' | 'hafal'
  noted_at: string
}

export interface HafalanSummary {
  totalAyat: number
  totalSurah: number
  totalJuz: number
  progressByJuz: Record<number, number>
}

export const useHafalanStore = defineStore('hafalan', {
  state: () => ({
    list: [] as HafalanProgress[],
    summary: null as HafalanSummary | null,
    loading: false,
    error: null as string | null,
  }),

  getters: {
    byStatus: (state) => (status: HafalanProgress['status']) =>
      state.list.filter((h) => h.status === status),
    bySurah: (state) => (surahNumber: number) =>
      state.list.filter((h) => h.surah_number === surahNumber),
  },

  actions: {
    async fetchList() {
      this.loading = true
      this.error = null
      try {
        const { apiFetch } = useApi()
        const res = await apiFetch<{ data: HafalanProgress[] }>('/hafalan')
        this.list = res.data
      } catch (e: any) {
        this.error = e?.data?.message || 'Gagal memuat data hafalan'
      } finally {
        this.loading = false
      }
    },

    async fetchSummary() {
      try {
        const { apiFetch } = useApi()
        const res = await apiFetch<{ data: HafalanSummary }>('/hafalan/summary')
        this.summary = res.data
      } catch {}
    },
  },
})
