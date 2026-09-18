import { onScopeDispose, ref, type Ref } from 'vue'

type Command =
  | 'player:play'
  | 'player:pause'
  | 'player:mute'
  | 'player:unMute'
  | 'player:hideControls'
  | 'player:showControls'

/** postMessage bridge to an embedded RuTube player. Docs: rutube.ru/info/embed */
export function useRutube(iframe: Ref<HTMLIFrameElement | null>) {
  const state = ref<'idle' | 'playing' | 'paused' | 'stopped'>('idle')
  const currentTime = ref(0)
  const ready = ref(false)
  const title = ref('')

  const send = (type: Command, data: Record<string, unknown> = {}) =>
    iframe.value?.contentWindow?.postMessage(JSON.stringify({ type, data }), 'https://rutube.ru')

  const onMessage = (e: MessageEvent) => {
    if (e.origin !== 'https://rutube.ru') return
    let msg: { type?: string; data?: { state?: string; time?: number; title?: string } }
    try {
      msg = typeof e.data === 'string' ? JSON.parse(e.data) : e.data
    } catch {
      return
    }
    switch (msg.type) {
      case 'player:ready':
        ready.value = true
        break
      case 'player:changeState':
        state.value = (msg.data?.state as typeof state.value) ?? state.value
        break
      case 'player:currentTime':
        currentTime.value = msg.data?.time ?? currentTime.value
        break
      case 'player:playOptionLoaded':
      case 'player:playOptionsLoaded': // docs spell it both ways
        if (msg.data?.title) title.value = msg.data.title
        break
    }
  }
  window.addEventListener('message', onMessage)
  onScopeDispose(() => window.removeEventListener('message', onMessage))

  return { state, currentTime, ready, title, send }
}
