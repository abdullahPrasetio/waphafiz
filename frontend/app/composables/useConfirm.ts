interface ConfirmOptions {
  title: string
  message?: string
  confirmText?: string
  cancelText?: string
  variant?: 'danger' | 'default'
}

interface ConfirmState extends Required<ConfirmOptions> {
  open: boolean
}

const defaultState: ConfirmState = {
  open: false,
  title: '',
  message: '',
  confirmText: 'Ya, lanjutkan',
  cancelText: 'Batal',
  variant: 'default',
}

// Singleton state + pending resolver — dipakai bersama oleh useConfirm() (pemanggil)
// dan AppConfirmDialog.vue (yang me-render), satu dialog untuk seluruh aplikasi.
const state = reactive<ConfirmState>({ ...defaultState })
let pendingResolve: ((value: boolean) => void) | null = null

export function useConfirmState() {
  function resolve(value: boolean) {
    state.open = false
    pendingResolve?.(value)
    pendingResolve = null
  }
  return { state, resolve }
}

// Pengganti window.confirm(): await useConfirm().confirm({ title, message }).
// Mengembalikan true jika user menekan tombol konfirmasi, false jika batal/klik luar.
export function useConfirm() {
  function confirm(options: ConfirmOptions): Promise<boolean> {
    Object.assign(state, defaultState, options, { open: true })
    return new Promise<boolean>((resolve) => {
      pendingResolve = resolve
    })
  }
  return { confirm }
}
