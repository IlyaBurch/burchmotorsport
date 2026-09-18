import { computed, watchEffect } from 'vue'
import { useStorage, usePreferredDark } from '@vueuse/core'

type Theme = 'dark' | 'light'

const prefersDark = usePreferredDark()
const stored = useStorage<Theme | null>('bm-theme', null)

const theme = computed<Theme>({
  get: () => stored.value ?? (prefersDark.value ? 'dark' : 'light'),
  set: (v) => { stored.value = v },
})

watchEffect(() => {
  document.documentElement.setAttribute('data-theme', theme.value)
})

export function useTheme() {
  function toggle() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function set(t: Theme) {
    theme.value = t
  }

  return { theme, toggle, set }
}
