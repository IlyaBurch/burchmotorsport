<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLocalStorage } from '@vueuse/core'
import { Columns2, Layers, Maximize, Minimize, PanelRightClose, PanelRightOpen, Volume2 } from 'lucide-vue-next'
import { BmButton, BmCard, BmChip, BmInput, BmModal, BmTabs } from '@/shared/ui'
import { parseRutubeId } from '@/shared/lib/rutube'
import { resolveTitle, sessionLabel } from '@/shared/api/live'
import { useLiveSession } from '@/entities/session'
import { RutubePlayer } from '@/features/rutube-player'
import { useTheater, type PanelMode } from '@/features/theater-mode'
import { LivePanel } from '@/widgets/live-panel'

const route = useRoute()
const router = useRouter()

// --- video from the url ------------------------------------------------------
const videoId = computed(() => parseRutubeId(String(route.query.v ?? '')))
const link = ref('')
const linkError = ref('')
function open() {
  const id = parseRutubeId(link.value)
  if (!id) {
    linkError.value = 'Это не ссылка на RuTube'
    return
  }
  linkError.value = ''
  router.replace({ query: { ...route.query, v: id } })
}

// --- telemetry for the session from the url ------------------------------------
const sessionKey = computed(() => String(route.query.session ?? 'latest'))
const { live, outline, error, upcoming, finished } = useLiveSession(sessionKey)

// no session in the url: the player posts the video title, the api maps it to a session
const resolved = ref<'' | 'found' | 'missed'>('')
async function onTitle(title: string) {
  if (route.query.session) return
  try {
    const r = await resolveTitle(title)
    if (!r.f1 && !r.session) {
      // not Formula 1: back to the form with the reason
      linkError.value = 'Похоже, это не Формула 1. Нужна ссылка на видео или трансляцию F1'
      link.value = ''
      router.replace({ query: { ...route.query, v: undefined } })
      return
    }
    resolved.value = r.session ? 'found' : 'missed'
    if (r.session) router.replace({ query: { ...route.query, session: String(r.session.session_key) } })
  } catch {
    resolved.value = 'missed'
  }
}
watch(videoId, () => (resolved.value = ''))

// --- layout --------------------------------------------------------------------
const wrapper = ref<HTMLElement | null>(null)
const player = ref<InstanceType<typeof RutubePlayer> | null>(null)
const theater = useTheater(wrapper)
const modes = [
  { key: 'side', label: 'Рядом' },
  { key: 'over', label: 'Поверх' },
]
const panelOpen = ref(true)

// sound needs a user gesture; autoplay started muted
function unmute() {
  player.value?.play()
  player.value?.unmute()
}
const toggleFullscreen = () => (theater.fullscreen.value ? theater.exit() : theater.enter())
const toggleMode = () => (theater.mode.value = theater.mode.value === 'side' ? 'over' : 'side')

// --- one-time warning about broadcast vs timing offset ---------------------------
const warned = useLocalStorage('bm.watch.warned', false)
const warning = ref(false)
watch(videoId, (id) => id && !warned.value && (warning.value = true), { immediate: true })
function ack() {
  warned.value = true
  warning.value = false
}
</script>

<template>
  <section class="watch">
    <BmCard v-if="!videoId" class="watch__intro">
      <div class="bm-card__eyebrow">Трансляция</div>
      <h1 class="display-md">Смотри гонку с телеметрией</h1>
      <p class="body">Вставь ссылку на трансляцию или запись с RuTube. Ссылка останется в адресной строке, ей можно делиться.</p>
      <form class="watch__form" @submit.prevent="open">
        <BmInput v-model="link" placeholder="https://rutube.ru/video/…" :error="linkError" />
        <BmButton variant="primary" type="submit">Открыть</BmButton>
      </form>
      <p class="body-sm">Телеметрия: {{ live ? `${live.session.circuit_short_name} · ${sessionLabel(live.session.session_name)}` : '…' }}. Другую сессию можно выбрать на странице телеметрии, ссылка на неё подставится сюда параметром <code>session</code>.</p>
    </BmCard>

    <template v-else>
      <div class="watch__bar">
        <div class="watch__title">
          <span class="display-sm">{{ live ? `${live.session.circuit_short_name} · ${sessionLabel(live.session.session_name)}` : 'Телеметрия' }}</span>
          <BmChip v-if="upcoming" variant="yellow">Ещё не началась</BmChip>
          <BmChip v-else-if="finished" variant="purple">Архив</BmChip>
          <BmChip v-else-if="error" variant="red">{{ error }}</BmChip>
          <BmChip v-if="resolved === 'found'" variant="green">По видео</BmChip>
          <BmChip v-else-if="resolved === 'missed'" variant="yellow">Сессию по видео не нашли</BmChip>
        </div>
        <div class="watch__controls">
          <BmTabs v-model="theater.mode.value" :tabs="modes" />
          <BmButton variant="primary" @click="unmute">
            <Volume2 :size="20" :stroke-width="2.5" />
            Звук
          </BmButton>
          <BmButton aria-label="Во весь экран" @click="toggleFullscreen">
            <Maximize :size="20" :stroke-width="2.5" />
          </BmButton>
        </div>
      </div>

      <div ref="wrapper" class="watch__stage" :class="`watch__stage--${theater.mode.value}`">
        <RutubePlayer ref="player" :video-id="videoId" class="watch__video" @title="onTitle" />
        <div v-if="theater.fullscreen.value" class="watch__fs-controls">
          <BmButton aria-label="Выйти из полного экрана" @click="theater.exit()">
            <Minimize :size="20" :stroke-width="2.5" />
          </BmButton>
          <BmButton :aria-label="theater.mode.value === 'side' ? 'Панель поверх видео' : 'Панель рядом с видео'" @click="toggleMode">
            <Layers v-if="theater.mode.value === 'side'" :size="20" :stroke-width="2.5" />
            <Columns2 v-else :size="20" :stroke-width="2.5" />
          </BmButton>
        </div>

        <div v-if="theater.mode.value === 'over'" class="watch__overlay">
          <BmButton class="watch__toggle" :aria-label="panelOpen ? 'Скрыть телеметрию' : 'Показать телеметрию'" @click="panelOpen = !panelOpen">
            <PanelRightClose v-if="panelOpen" :size="20" :stroke-width="2.5" />
            <PanelRightOpen v-else :size="20" :stroke-width="2.5" />
          </BmButton>
          <LivePanel v-show="panelOpen" :live="live" :outline="outline" compact class="watch__panel" />
        </div>
        <LivePanel v-else :live="live" :outline="outline" class="watch__panel" />
      </div>
    </template>

    <BmModal :open="warning" @close="ack">
      <div class="bm-card__eyebrow">Перед стартом</div>
      <h2 class="display-md">Время может не совпадать</h2>
      <p class="body">Трансляция и телеметрия идут по разным каналам с разной задержкой. Телеметрия может опережать картинку или отставать от неё, и это не в наших руках.</p>
      <BmButton variant="primary" style="margin-top: var(--space-4)" @click="ack">Понял</BmButton>
    </BmModal>
  </section>
</template>

<style scoped>
.watch {
  display: grid;
  gap: var(--space-4);
}

.watch__form {
  display: grid;
  gap: var(--space-3);
  margin-block: var(--space-4);
}

@media (min-width: 640px) {
  .watch__form {
    grid-template-columns: 1fr auto;
    align-items: end;
  }
}

.watch__bar {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
}

.watch__title,
.watch__controls {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

/* side: video + panel; over: panel floats on the video */
.watch__stage {
  position: relative;
  display: grid;
  gap: var(--space-4);
  background: var(--surface);
}

.watch__stage--side {
  grid-template-columns: minmax(0, 1fr);
}

@media (min-width: 1024px) {
  .watch__stage--side {
    grid-template-columns: minmax(0, 1fr) 360px;
    align-items: start;
  }
}

.watch__stage--over .watch__video {
  width: 100%;
}

.watch__overlay {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  bottom: var(--space-4);
  width: min(360px, calc(100% - var(--space-8)));
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  justify-items: end;
  gap: var(--space-2);
  pointer-events: none;
}

.watch__overlay > * {
  pointer-events: auto;
}

.watch__overlay .watch__panel {
  width: 100%;
}

.watch__fs-controls {
  position: absolute;
  top: var(--space-4);
  left: var(--space-4);
  z-index: 2;
  display: flex;
  gap: var(--space-2);
}

/* fullscreen: only video + panel, in whichever mode was picked */
.watch__stage:fullscreen {
  padding: var(--space-4);
  box-sizing: border-box;
  height: 100%;
  align-items: stretch;
}

.watch__stage--over:fullscreen {
  display: block;
  padding: 0;
}

.watch__stage:fullscreen .watch__video {
  height: 100%;
  aspect-ratio: auto;
  border: 0;
}

.watch__stage--side:fullscreen {
  grid-template-columns: minmax(0, 1fr) 360px;
}

.watch__stage--side:fullscreen .watch__panel {
  max-height: 100%;
  overflow-y: auto;
}
</style>
