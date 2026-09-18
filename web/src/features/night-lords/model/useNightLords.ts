import { computed, ref } from 'vue'
import { useTheme } from '@/features/theme-toggle'

const TAPS = 8
const WINDOW_MS = 4000

/**
 * Easter egg: tap the logo 8 times within 4 seconds to switch to the
 * Night Lords theme (and back). "Ave Dominus Nox."
 */
export function useNightLords() {
  const { theme, set } = useTheme()
  const active = computed(() => theme.value === 'night-lords')
  const toast = ref(false)
  let taps: number[] = []
  let toastTimer: ReturnType<typeof setTimeout> | undefined

  function tap() {
    const now = Date.now()
    taps = [...taps.filter((t) => now - t < WINDOW_MS), now]
    if (taps.length < TAPS) return
    taps = []
    set(active.value ? 'dark' : 'night-lords')
    toast.value = true
    clearTimeout(toastTimer)
    toastTimer = setTimeout(() => (toast.value = false), 2500)
  }

  return { active, toast, tap }
}
