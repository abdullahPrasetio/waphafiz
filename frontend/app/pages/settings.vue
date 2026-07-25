<template>
  <div class="p-5 md:p-8 max-w-lg">
    <div class="mb-6">
      <h1 class="text-xl font-semibold">Pengaturan Akun</h1>
      <p class="text-[13px] text-gray-500 mt-0.5">Kelola informasi keamanan akun kamu</p>
    </div>

    <div class="card">
      <div class="flex items-center gap-2.5 mb-1">
        <IconLock :size="17" class="text-green-500" />
        <span class="text-sm font-medium">Ganti Password</span>
      </div>
      <p class="text-[12.5px] text-gray-500 mb-5">Gunakan password baru yang kuat dan belum pernah dipakai sebelumnya.</p>

      <form @submit.prevent="handleChangePassword">
        <div class="mb-4">
          <label class="form-label">Password saat ini</label>
          <div class="relative">
            <IconLock class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
            <input
              v-model="form.currentPassword"
              :type="showCurrent ? 'text' : 'password'"
              class="form-input pr-10"
              placeholder="Masukkan password saat ini"
              required
              autocomplete="current-password"
            />
            <button
              type="button"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 bg-none border-none"
              @click="showCurrent = !showCurrent"
            >
              <IconEye v-if="!showCurrent" :size="16" />
              <IconEyeOff v-else :size="16" />
            </button>
          </div>
        </div>

        <div class="mb-2">
          <label class="form-label">Password baru</label>
          <div class="relative">
            <IconKey class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
            <input
              v-model="form.newPassword"
              :type="showNew ? 'text' : 'password'"
              class="form-input pr-10"
              placeholder="Min. 8 karakter"
              required
              minlength="8"
              autocomplete="new-password"
            />
            <button
              type="button"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 bg-none border-none"
              @click="showNew = !showNew"
            >
              <IconEye v-if="!showNew" :size="16" />
              <IconEyeOff v-else :size="16" />
            </button>
          </div>
        </div>

        <!-- Password strength hint -->
        <div class="flex gap-1 mb-4 mt-2">
          <div
            v-for="i in 4"
            :key="i"
            class="h-1 flex-1 rounded-full transition-colors"
            :class="i <= strengthScore ? strengthColor : 'bg-gray-100'"
          />
        </div>

        <div class="mb-6">
          <label class="form-label">Konfirmasi password baru</label>
          <div class="relative">
            <IconKey class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
            <input
              v-model="form.confirmPassword"
              type="password"
              class="form-input"
              placeholder="Ulangi password baru"
              required
              minlength="8"
              autocomplete="new-password"
            />
          </div>
          <p v-if="form.confirmPassword && form.confirmPassword !== form.newPassword" class="text-red-500 text-[11.5px] mt-1.5">
            Konfirmasi password tidak cocok
          </p>
        </div>

        <p v-if="error" class="text-red-500 text-xs mb-4">{{ error }}</p>
        <p v-if="success" class="text-green-600 text-xs mb-4 flex items-center gap-1.5">
          <IconCheck :size="14" /> {{ success }}
        </p>

        <button type="submit" class="btn-primary w-full" :disabled="loading || !canSubmit">
          {{ loading ? 'Menyimpan...' : 'Simpan Password Baru' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { IconLock, IconKey, IconEye, IconEyeOff, IconCheck } from '@tabler/icons-vue'

definePageMeta({ middleware: 'auth' })

const { apiFetch } = useApi()

const form = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const showCurrent = ref(false)
const showNew = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

const strengthScore = computed(() => {
  const pw = form.newPassword
  if (!pw) return 0
  let score = 0
  if (pw.length >= 8) score++
  if (pw.length >= 12) score++
  if (/[A-Z]/.test(pw) && /[a-z]/.test(pw)) score++
  if (/[0-9]/.test(pw) && /[^A-Za-z0-9]/.test(pw)) score++
  return score
})

const strengthColor = computed(() => {
  if (strengthScore.value <= 1) return 'bg-red-400'
  if (strengthScore.value === 2) return 'bg-amber-400'
  return 'bg-green-500'
})

const canSubmit = computed(() =>
  form.currentPassword.length > 0 &&
  form.newPassword.length >= 8 &&
  form.newPassword === form.confirmPassword
)

async function handleChangePassword() {
  error.value = ''
  success.value = ''

  if (form.newPassword !== form.confirmPassword) {
    error.value = 'Konfirmasi password baru tidak cocok'
    return
  }
  if (form.newPassword === form.currentPassword) {
    error.value = 'Password baru tidak boleh sama dengan password lama'
    return
  }

  loading.value = true
  try {
    await apiFetch('/auth/change-password', {
      method: 'POST',
      body: {
        current_password: form.currentPassword,
        new_password: form.newPassword,
      },
    })
    success.value = 'Password berhasil diubah.'
    form.currentPassword = ''
    form.newPassword = ''
    form.confirmPassword = ''
  } catch (e: any) {
    error.value = e?.data?.message || 'Gagal mengubah password, coba lagi'
  } finally {
    loading.value = false
  }
}
</script>
