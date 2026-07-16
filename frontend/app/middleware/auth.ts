export default defineNuxtRouteMiddleware((to) => {
  // Token hanya ada di localStorage (lihat stores/auth.ts), yang tidak bisa
  // diakses saat SSR. Guard ini sengaja dilewati di server agar full page
  // reload tidak salah dianggap "belum login" — client akan re-check setelah hydration.
  if (import.meta.server) return

  const authStore = useAuthStore()
  authStore.restoreFromStorage()

  const publicRoutes = ['/login', '/register']
  if (publicRoutes.includes(to.path)) return

  if (!authStore.isLoggedIn) {
    return navigateTo('/login')
  }

  if (to.path.startsWith('/admin') && !authStore.isAdmin) {
    return navigateTo('/dashboard')
  }
})
