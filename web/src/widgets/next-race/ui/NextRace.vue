<script setup lang="ts">
import { computed, ref } from 'vue'
import { useIntervalFn } from '@vueuse/core'

interface Next {
  session_key: number
  location: string
  date_start: string
}

const next = ref<Next | null>(null)
fetch('/api/next')
  .then((r) => (r.ok ? r.json() : null))
  .then((n) => (next.value = n))
  .catch(() => {})

const now = ref(Date.now())
useIntervalFn(() => (now.value = Date.now()), 30_000)

const startMs = computed(() => (next.value ? Date.parse(next.value.date_start) : 0))

const startMsk = computed(() =>
  next.value
    ? new Date(startMs.value).toLocaleString('ru-RU', {
        timeZone: 'Europe/Moscow',
        weekday: 'short',
        day: 'numeric',
        month: 'short',
        hour: '2-digit',
        minute: '2-digit',
      }) + ' МСК'
    : '',
)

const left = computed(() => {
  const ms = startMs.value - now.value
  if (ms <= 0) return 'сейчас'
  const h = Math.floor(ms / 3_600_000)
  const d = Math.floor(h / 24)
  const m = Math.floor((ms % 3_600_000) / 60_000)
  return d > 0 ? `${d}д ${h % 24}ч` : h > 0 ? `${h}ч ${m}м` : `${m}м`
})
</script>

<template>
  <RouterLink v-if="next" class="next" :to="{ path: '/live', query: { session: String(next.session_key) } }" :title="startMsk">
    <span class="display-sm next__label">Гонка</span>
    <span class="body-strong next__where">{{ next.location }}</span>
    <span class="timing next__left">{{ left }}</span>
    <span class="body-sm next__when">{{ startMsk }}</span>
  </RouterLink>
</template>

<style scoped>
.next {
  display: inline-grid;
  grid-template-columns: auto auto;
  column-gap: var(--space-2);
  align-items: baseline;
  padding: var(--space-1) var(--space-3);
  border: var(--border-thin) solid var(--border);
  background: var(--surface-raised);
  color: var(--text);
  text-decoration: none;
  white-space: nowrap;
}

.next:hover {
  background: var(--surface-raised);
  color: var(--text);
}

.next__label {
  color: var(--accent);
}

.next__where,
.next__when {
  display: none;
}

@media (min-width: 640px) {
  .next__where { display: inline; }
}

@media (min-width: 1024px) {
  .next {
    grid-template-columns: auto auto auto auto;
  }
  .next__when { display: inline; }
}
</style>
