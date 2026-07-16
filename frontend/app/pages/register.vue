<template>
  <div class="bg-white border border-gray-200 rounded-xl p-10 w-full max-w-sm">
    <div class="flex items-center gap-2.5 mb-8">
      <div class="w-10 h-10 bg-green-500 rounded-[10px] flex items-center justify-center text-white font-semibold text-base">W</div>
      <div class="text-xl font-semibold">WAP<span class="text-green-500">Hafiz</span></div>
    </div>

    <!-- Toggle flow -->
    <div class="flex rounded-[10px] border border-gray-200 p-0.5 mb-6">
      <button
        type="button"
        class="flex-1 py-2 text-[13px] font-medium rounded-[8px] transition-all"
        :class="flow === 'join' ? 'bg-green-500 text-white' : 'text-gray-500'"
        @click="flow = 'join'"
      >Gabung Keluarga</button>
      <button
        type="button"
        class="flex-1 py-2 text-[13px] font-medium rounded-[8px] transition-all"
        :class="flow === 'create' ? 'bg-green-500 text-white' : 'text-gray-500'"
        @click="flow = 'create'"
      >Buat Keluarga Baru</button>
    </div>

    <!-- Info banner -->
    <div class="bg-green-50 border border-green-200 rounded-[10px] px-3.5 py-2.5 flex gap-2 mb-4 text-xs text-green-900 leading-relaxed">
      <IconInfoCircle class="text-green-500 flex-shrink-0 mt-0.5" :size="15" />
      <span v-if="flow === 'join'">Masukkan kode undangan dari admin keluarga untuk bergabung.</span>
      <span v-else>Kamu akan membuat keluarga baru dan menjadi admin pertama.</span>
    </div>

    <form @submit.prevent="handleRegister">
      <div v-if="flow === 'join'" class="mb-4">
        <label class="form-label">Kode undangan</label>
        <div class="relative">
          <IconKey class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
          <input v-model="form.inviteCode" type="text" class="form-input" placeholder="FAM-2026-XXXX" required />
        </div>
      </div>

      <div v-if="flow === 'create'" class="mb-4">
        <label class="form-label">Nama keluarga</label>
        <div class="relative">
          <IconHome class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
          <input v-model="form.familyName" type="text" class="form-input" placeholder="Keluarga Prasetio" required />
        </div>
      </div>

      <div class="mb-4">
        <label class="form-label">Nama lengkap</label>
        <div class="relative">
          <IconUser class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
          <input v-model="form.name" type="text" class="form-input" placeholder="Nama kamu" required />
        </div>
      </div>

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
        {{ loading ? 'Memuat...' : flow === 'join' ? 'Bergabung' : 'Buat Keluarga' }}
      </button>
    </form>

    <p class="text-center text-[13px] text-gray-500 mt-5">
      Sudah punya akun? <NuxtLink to="/login" class="text-green-500 font-medium">Masuk</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { IconKey, IconUser, IconMail, IconLock, IconHome, IconInfoCircle } from '@tabler/icons-vue'

definePageMeta({ layout: 'auth' })

const authStore = useAuthStore()
const router = useRouter()

const flow = ref<'join' | 'create'>('join')
const form = reactive({ inviteCode: '', familyName: '', name: '', email: '', password: '' })
const loading = ref(false)
const error = ref('')

async function handleRegister() {
  loading.value = true
  error.value = ''
  try {
    const config = useRuntimeConfig()
    const endpoint = flow.value === 'create' ? '/auth/register/family' : '/auth/register'
    const body = flow.value === 'create'
      ? { family_name: form.familyName, name: form.name, email: form.email, password: form.password }
      : { invite_code: form.inviteCode, name: form.name, email: form.email, password: form.password }

    const res = await $fetch<{ data: { token: string; user: any } }>(endpoint, {
      baseURL: config.public.apiBase as string,
      method: 'POST',
      body,
    })
    authStore.setAuth(res.data.token, res.data.user)
    router.push('/dashboard')
  } catch (e: any) {
    error.value = e?.data?.message || 'Pendaftaran gagal, coba lagi'
  } finally {
    loading.value = false
  }
}
</script>
