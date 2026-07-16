import type { FetchOptions } from 'ofetch'

export function useApi() {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()

  const baseURL = config.public.apiBase as string

  async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string> || {}),
    }

    if (authStore.token) {
      headers['Authorization'] = `Bearer ${authStore.token}`
    }

    return $fetch<T>(path, {
      baseURL,
      ...options,
      headers,
      onResponseError({ response }) {
        if (response.status === 401) {
          authStore.logout()
          navigateTo('/login')
        }
      },
    })
  }

  return { apiFetch }
}
