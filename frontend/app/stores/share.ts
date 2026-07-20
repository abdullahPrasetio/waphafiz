export interface ShareInfo {
  id: string
  user_id: string
  user_name: string
  label: string
  expires_at: string | null
  last_viewed_at: string | null
  view_count: number
  status: 'aktif' | 'kedaluwarsa' | 'dicabut'
}

export interface CreateShareResult {
  id: string
  label: string
  expires_at: string | null
  token: string // hanya muncul sekali di response create
}

export const useShareStore = defineStore('share', {
  state: () => ({
    list: [] as ShareInfo[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchList(userId?: string) {
      this.loading = true
      this.error = null
      try {
        const { apiFetch } = useApi()
        const path = userId ? `/shares?user_id=${userId}` : '/shares'
        const res = await apiFetch<{ data: ShareInfo[] }>(path)
        this.list = res.data || []
      } catch (e: any) {
        this.error = e?.data?.message || 'Gagal memuat daftar link pantau'
      } finally {
        this.loading = false
      }
    },

    async create(input: { user_id?: string; label: string; expires_in_days: number }): Promise<CreateShareResult> {
      const { apiFetch } = useApi()
      const res = await apiFetch<{ data: CreateShareResult }>('/shares', {
        method: 'POST',
        body: input,
      })
      return res.data
    },

    async revoke(shareId: string) {
      const { apiFetch } = useApi()
      await apiFetch(`/shares/${shareId}`, { method: 'DELETE' })
    },
  },
})
