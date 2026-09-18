import { computed, onScopeDispose, ref, type Ref } from 'vue'
import { useLocalStorage } from '@vueuse/core'

export type PanelMode = 'side' | 'over'

/**
 * Video + telemetry layout: panel beside the video or over it. Fullscreen is
 * requested on the wrapper so the panel stays visible, and forces "over".
 */
export function useTheater(wrapper: Ref<HTMLElement | null>) {
  const mode = useLocalStorage<PanelMode>('bm.watch.mode', 'side')
  const fullscreen = ref(false)

  const sync = () => (fullscreen.value = document.fullscreenElement === wrapper.value && !!wrapper.value)
  document.addEventListener('fullscreenchange', sync)
  onScopeDispose(() => document.removeEventListener('fullscreenchange', sync))

  const enter = () => wrapper.value?.requestFullscreen?.().catch(() => {})
  const exit = () => document.exitFullscreen?.().catch(() => {})

  const effective = computed<PanelMode>(() => (fullscreen.value ? 'over' : mode.value))

  return { mode, fullscreen, effective, enter, exit }
}
