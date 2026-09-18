import { computed, onScopeDispose, ref, watch, type Ref } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { fetchLive, fetchTrack, isFinished, NoDataError, RateLimitError, type Live } from '@/shared/api/live'

/**
 * Loads a session by key and keeps it fresh while it is running.
 * Shared by the telemetry page and the watch page.
 */
export function useLiveSession(sessionKey: Ref<string>) {
  const live = ref<Live | null>(null)
  const outline = ref<[number, number][]>([])
  const error = ref<string | null>(null)
  const noData = ref(false)

  let retryTimer: ReturnType<typeof setTimeout> | undefined
  async function refresh() {
    clearTimeout(retryTimer)
    try {
      live.value = await fetchLive(sessionKey.value)
      error.value = null
      noData.value = false
    } catch (e) {
      noData.value = e instanceof NoDataError
      if (e instanceof RateLimitError) {
        error.value = `Лимит OpenF1, повтор через ${e.seconds} с`
        retryTimer = setTimeout(refresh, e.seconds * 1000)
      } else {
        error.value = noData.value ? null : 'Нет связи с данными'
      }
    }
  }

  const poll = useIntervalFn(refresh, 10_000, { immediate: false })
  onScopeDispose(() => clearTimeout(retryTimer))

  watch(
    sessionKey,
    async (key) => {
      live.value = null
      outline.value = []
      noData.value = false
      await refresh()
      const l = live.value as Live | null
      if (l && !l.upcoming) fetchTrack(key).then((o) => (outline.value = o)).catch(() => {}) // map is optional
      if (l && !l.upcoming && !isFinished(l.session)) poll.resume()
      else poll.pause()
    },
    { immediate: true },
  )

  const upcoming = computed(() => !!live.value?.upcoming)
  const finished = computed(() => !!live.value && !live.value.upcoming && isFinished(live.value.session))
  const loading = computed(() => !live.value && !error.value && !noData.value)

  return { live, outline, error, noData, upcoming, finished, loading, refresh }
}
