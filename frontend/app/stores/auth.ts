interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'member'
  familyGroupId: string
  familyGroupName: string
}

interface AuthState {
  token: string | null
  user: User | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: null,
    user: null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'admin',
  },

  actions: {
    setAuth(token: string, user: User) {
      this.token = token
      this.user = user
      if (import.meta.client) {
        localStorage.setItem('waphafiz_token', token)
        localStorage.setItem('waphafiz_user', JSON.stringify(user))
      }
    },

    restoreFromStorage() {
      if (!import.meta.client) return
      const token = localStorage.getItem('waphafiz_token')
      const userRaw = localStorage.getItem('waphafiz_user')
      if (token && userRaw) {
        this.token = token
        this.user = JSON.parse(userRaw)
      }
    },

    logout() {
      this.token = null
      this.user = null
      if (import.meta.client) {
        localStorage.removeItem('waphafiz_token')
        localStorage.removeItem('waphafiz_user')
      }
    },
  },
})
