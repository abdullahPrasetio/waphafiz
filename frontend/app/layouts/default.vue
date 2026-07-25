<template>
  <div class="flex min-h-screen">
    <!-- Sidebar desktop -->
    <aside class="hidden md:flex flex-col w-56 bg-white border-r border-gray-200 sticky top-0 h-screen flex-shrink-0">
      <div class="p-5 border-b border-gray-200">
        <div class="text-[17px] font-semibold text-gray-900">
          WAP<span class="text-green-500">Hafiz</span>
        </div>
        <ClientOnly>
          <div class="text-[11px] text-gray-400 mt-0.5">{{ authStore.user?.familyGroupName }}</div>
        </ClientOnly>
      </div>

      <nav class="flex-1 py-3 overflow-y-auto">
        <NuxtLink to="/dashboard" class="nav-item" activeClass="active">
          <IconLayoutDashboard :size="18" /> Dashboard
        </NuxtLink>
        <NuxtLink to="/quran" class="nav-item" activeClass="active">
          <IconBook :size="18" /> Al-Quran
        </NuxtLink>
        <NuxtLink to="/hafalan" class="nav-item" activeClass="active">
          <IconChecklist :size="18" /> Hafalan Saya
        </NuxtLink>
        <NuxtLink to="/murajaah" class="nav-item" activeClass="active">
          <IconRefresh :size="18" /> Muraja'ah
        </NuxtLink>

        <ClientOnly>
          <template v-if="authStore.isAdmin">
            <div class="text-[10px] text-gray-400 px-5 pt-4 pb-1 uppercase tracking-wider">Admin</div>
            <NuxtLink to="/admin/members" class="nav-item" activeClass="active">
              <IconUsers :size="18" /> Anggota
            </NuxtLink>
            <NuxtLink to="/admin/dashboard" class="nav-item" activeClass="active">
              <IconLayoutDashboard :size="18" /> Dashboard
            </NuxtLink>
            <NuxtLink to="/admin/progress" class="nav-item" activeClass="active">
              <IconChartBar :size="18" /> Progress
            </NuxtLink>
          </template>
        </ClientOnly>
      </nav>

      <ClientOnly>
        <div class="p-4 border-t border-gray-200 flex items-center gap-2.5">
          <NuxtLink to="/settings" class="flex items-center gap-2.5 flex-1 min-w-0 no-underline text-inherit">
            <AppAvatar :name="authStore.user?.name || ''" size="sm" />
            <div class="flex-1 min-w-0">
              <div class="text-[13px] font-medium text-gray-900 truncate">{{ authStore.user?.name }}</div>
              <div class="text-[11px] text-gray-400 capitalize">{{ authStore.user?.role }}</div>
            </div>
          </NuxtLink>
          <NuxtLink to="/settings" class="text-gray-400 hover:text-green-500 flex-shrink-0" title="Pengaturan Akun">
            <IconSettings :size="17" />
          </NuxtLink>
          <button
            class="text-gray-400 hover:text-red-500 border-none bg-none flex-shrink-0"
            title="Keluar"
            @click="handleLogout">
            <IconLogout :size="17" />
          </button>
        </div>
      </ClientOnly>
    </aside>

    <!-- Top bar mobile -->
    <header class="md:hidden fixed top-0 left-0 right-0 z-40 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
      <div class="text-[15px] font-semibold text-gray-900">
        WAP<span class="text-green-500">Hafiz</span>
      </div>
      <div class="flex items-center gap-3">
        <NuxtLink to="/settings" class="text-gray-400 hover:text-green-500" title="Pengaturan Akun">
          <IconSettings :size="19" />
        </NuxtLink>
        <button class="text-gray-400 hover:text-red-500 border-none bg-none" title="Keluar" @click="handleLogout">
          <IconLogout :size="19" />
        </button>
      </div>
    </header>

    <!-- Main -->
    <main class="flex-1 bg-gray-50 overflow-y-auto pb-20 md:pb-0 pt-14 md:pt-0">
      <slot />
    </main>

    <!-- Bottom nav mobile -->
    <nav class="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 flex z-50">
      <NuxtLink to="/dashboard" class="flex-1 flex flex-col items-center gap-1 py-2.5 text-[9px] text-gray-400 [&.router-link-active]:text-green-500">
        <IconLayoutDashboard :size="20" />Dashboard
      </NuxtLink>
      <NuxtLink to="/quran" class="flex-1 flex flex-col items-center gap-1 py-2.5 text-[9px] text-gray-400 [&.router-link-active]:text-green-500">
        <IconBook :size="20" />Quran
      </NuxtLink>
      <NuxtLink to="/hafalan" class="flex-1 flex flex-col items-center gap-1 py-2.5 text-[9px] text-gray-400 [&.router-link-active]:text-green-500">
        <IconChecklist :size="20" />Hafalan
      </NuxtLink>
      <NuxtLink to="/murajaah" class="flex-1 flex flex-col items-center gap-1 py-2.5 text-[9px] text-gray-400 [&.router-link-active]:text-green-500">
        <IconRefresh :size="20" />Muraja'ah
      </NuxtLink>
      <ClientOnly>
        <NuxtLink v-if="authStore.isAdmin" to="/admin/members" class="flex-1 flex flex-col items-center gap-1 py-2.5 text-[9px] text-gray-400 [&.router-link-active]:text-green-500">
          <IconUsers :size="20" />Admin
        </NuxtLink>
      </ClientOnly>
    </nav>
  </div>
</template>

<script setup lang="ts">
import {
  IconLayoutDashboard, IconBook, IconChecklist,
  IconRefresh, IconUsers, IconChartBar, IconLogout, IconSettings,
} from '@tabler/icons-vue'

const authStore = useAuthStore()

async function handleLogout() {
  const ok = await useConfirm().confirm({
    title: 'Keluar dari akun?',
    message: 'Kamu perlu login kembali untuk mengakses aplikasi.',
    confirmText: 'Keluar',
    variant: 'danger',
  })
  if (!ok) return
  authStore.logout()
  navigateTo('/login')
}
</script>
