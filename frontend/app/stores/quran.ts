export interface Surah {
  number: number
  name: string
  englishName: string
  numberOfAyahs: number
  revelationType: string
  juzStart: number
}

export interface Ayah {
  number: number
  numberInSurah: number
  text: string
  translation: string
}

export const useQuranStore = defineStore('quran', {
  state: () => ({
    surahList: [] as Surah[],
    currentSurah: null as (Surah & { ayahs: Ayah[] }) | null,
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchSurahList() {
      if (this.surahList.length) return
      this.loading = true
      try {
        const { apiFetch } = useApi()
        const res = await apiFetch<{ data: Surah[] }>('/quran/surah')
        this.surahList = res.data
      } catch (e: any) {
        this.error = e?.data?.message || 'Gagal memuat daftar surah'
      } finally {
        this.loading = false
      }
    },

    async fetchSurahDetail(number: number) {
      this.loading = true
      this.error = null
      try {
        const { apiFetch } = useApi()
        const res = await apiFetch<{ data: Surah & { ayahs: Ayah[] } }>(`/quran/surah/${number}`)
        this.currentSurah = res.data
      } catch (e: any) {
        this.error = e?.data?.message || 'Audio tidak tersedia saat ini'
      } finally {
        this.loading = false
      }
    },
  },
})
