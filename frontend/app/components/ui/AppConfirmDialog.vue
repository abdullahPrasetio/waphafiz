<template>
  <Teleport to="body">
    <Transition name="confirm-backdrop">
      <div v-if="state.open" class="fixed inset-0 z-[300] flex items-end md:items-center justify-center bg-black/40 p-0 md:p-4"
        @click.self="cancel">
        <Transition name="confirm-card" appear>
          <div class="bg-white w-full md:max-w-sm rounded-t-[20px] md:rounded-[14px] p-5">
            <div class="flex items-start gap-3 mb-4">
              <div class="w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0"
                :class="state.variant === 'danger' ? 'bg-red-50 text-red-500' : 'bg-green-50 text-green-500'">
                <IconAlertTriangle v-if="state.variant === 'danger'" :size="18" />
                <IconInfoCircle v-else :size="18" />
              </div>
              <div>
                <div class="text-[14px] font-semibold text-gray-900">{{ state.title }}</div>
                <p v-if="state.message" class="text-[13px] text-gray-500 mt-1">{{ state.message }}</p>
              </div>
            </div>
            <div class="flex gap-2.5">
              <button
                class="flex-1 py-2 text-[13px] font-medium text-gray-600 bg-gray-50 hover:bg-gray-100 rounded-[10px] border-none"
                @click="cancel">
                {{ state.cancelText }}
              </button>
              <button
                class="flex-1 py-2 text-[13px] font-medium text-white rounded-[10px] border-none"
                :class="state.variant === 'danger' ? 'bg-red-500 hover:bg-red-600' : 'bg-green-500 hover:bg-green-600'"
                @click="confirm">
                {{ state.confirmText }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { IconAlertTriangle, IconInfoCircle } from '@tabler/icons-vue'

const { state, resolve } = useConfirmState()

function confirm() {
  resolve(true)
}
function cancel() {
  resolve(false)
}
</script>

<style scoped>
.confirm-backdrop-enter-active, .confirm-backdrop-leave-active { transition: opacity 0.15s ease; }
.confirm-backdrop-enter-from, .confirm-backdrop-leave-to { opacity: 0; }
.confirm-card-enter-active { transition: all 0.18s ease; }
.confirm-card-enter-from { opacity: 0; transform: translateY(12px); }
</style>
