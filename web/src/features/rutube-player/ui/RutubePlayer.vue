<script setup lang="ts">
import { ref, watch } from 'vue'
import { rutubeEmbedUrl } from '@/shared/lib/rutube'
import { useRutube } from '../model/useRutube'

const props = defineProps<{ videoId: string }>()
const emit = defineEmits<{ state: [state: string]; time: [seconds: number]; title: [title: string] }>()

const frame = ref<HTMLIFrameElement | null>(null)
const player = useRutube(frame)
watch(player.state, (s) => emit('state', s))
watch(player.currentTime, (t) => emit('time', t))
watch(player.title, (t) => emit('title', t))

defineExpose({
  unmute: () => player.send('player:unMute'),
  play: () => player.send('player:play'),
  hideControls: () => player.send('player:hideControls'),
  showControls: () => player.send('player:showControls'),
  state: player.state,
})
</script>

<template>
  <div class="player">
    <iframe
      ref="frame"
      :src="rutubeEmbedUrl(props.videoId)"
      title="Трансляция RuTube"
      allow="autoplay; fullscreen; picture-in-picture; encrypted-media"
      allowfullscreen
    />
  </div>
</template>

<style scoped>
.player {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: var(--surface-sunken);
  border: var(--border-thick) solid var(--border);
  box-sizing: border-box;
}

.player iframe {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  border: 0;
}
</style>
