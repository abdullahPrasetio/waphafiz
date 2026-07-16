export interface MurajaahSchedule {
  id: string
  type: 'sabqi' | 'manzil'
  surah_number: number
  ayat_start: number
  ayat_end: number
  scheduled_date: string
  completed_at: string | null
}

export const useMurajaahStore = defineStore('murajaah', {
  state: () => ({
    today: [] as MurajaahSchedule[],
    loading: false,
    error: null as string | null,
  }),

  getters: {
    sabqi: (state) => state.today.filter((s) => s.type === 'sabqi'),
    manzil: (state) => state.today.filter((s) => s.type === 'manzil'),
    completedCount: (state) => state.today.filter((s) => s.completed_at).length,
    totalCount: (state) => state.today.length,
    allDone: (state) => state.today.length > 0 && state.today.every((s) => s.completed_at),
  },

  actions: {
    async fetchToday() {
      this.loading = true
      this.error = null
      try {
        const { apiFetch } = useApi()
        const res = await apiFetch<{ data: MurajaahSchedule[] }>('/murajaah/today')
        this.today = res.data ?? []
      } catch (e: any) {
        this.error = e?.data?.message || 'Gagal memuat jadwal muraja\'ah'
      } finally {
        this.loading = false
      }
    },

    async markComplete(id: string) {
      // Optimistic update
      const item = this.today.find((s) => s.id === id)
      if (item) item.completed_at = new Date().toISOString()

      try {
        const { apiFetch } = useApi()
        await apiFetch(`/murajaah/${id}/complete`, { method: 'POST' })
      } catch {
        // Rollback
        const item = this.today.find((s) => s.id === id)
        if (item) item.completed_at = null
      }
    },
  },
})
