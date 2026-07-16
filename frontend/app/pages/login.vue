<template>
  <div class="bg-white border border-gray-200 rounded-xl p-10 w-full max-w-sm">
    <div class="flex items-center gap-2.5 mb-8">
      <div class="w-10 h-10 bg-green-500 rounded-[10px] flex items-center justify-center text-white font-semibold text-base">W</div>
      <div class="text-xl font-semibold">WAP<span class="text-green-500">Hafiz</span></div>
    </div>

    <h1 class="text-[22px] font-semibold mb-1">Masuk ke akun</h1>
    <p class="text-[13.5px] text-gray-500 mb-6">Selamat datang kembali, hafidz.</p>

    <form @submit.prevent="handleLogin">
      <div class="mb-4">
        <label class="form-label">Email</label>
        <div class="relative">
          <IconMail class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
          <input v-model="form.email" type="email" class="form-input" placeholder="email@contoh.com" required />
        </div>
      </div>

      <div class="mb-6">
        <label class="form-label">Password</label>
        <div class="relative">
          <IconLock class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
          <input v-model="form.password" type="password" class="form-input" placeholder="Min. 8 karakter" required minlength="8" />
        </div>
      </div>

      <p v-if="error" class="text-red-500 text-xs mb-4">{{ error }}</p>

      <button type="submit" class="btn-primary w-full" :disabled="loading">
        {{ loading ? 'Memuat...' : 'Masuk' }}
      </button>
    </form>

    <p class="text-center text-[13px] text-gray-500 mt-5">
      Belum punya akun?
      <NuxtLink to="/register" class="text-green-500 font-medium">Daftar dengan kode undangan</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { IconMail, IconLock } from '@tabler/icons-vue'

definePageMeta({ layout: 'auth' })

const authStore = useAuthStore()
const router = useRouter()

const form = reactive({ email: '', password: '' })
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    const config = useRuntimeConfig()
    const res = await $fetch<{ data: { token: string; user: any } }>('/auth/login', {
      baseURL: config.public.apiBase as string,
      method: 'POST',
      body: form,
    })
    authStore.setAuth(res.data.token, res.data.user)
    router.push('/dashboard')
  } catch (e: any) {
    error.value = e?.data?.message || 'Email atau password salah'
  } finally {
    loading.value = false
  }
}
</script>
