<script setup lang="ts">
import { computed, ref } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { countdown, fetchNext, formatMsk, sessionLabel, type NextSession } from '@/shared/api/live'

const next = ref<NextSession | null>(null)
fetchNext()
  .then((n) => (next.value = n))
  .catch(() => {})

const now = ref(Date.now())
useIntervalFn(() => (now.value = Date.now()), 30_000)

const startMs = computed(() => (next.value ? Date.parse(next.value.date_start) : 0))

const startMsk = computed(() => (next.value ? formatMsk(next.value.date_start) : ''))
const left = computed(() => countdown(startMs.value - now.value))
</script>

<template>
  <RouterLink
    v-if="next"
    class="next"
    :to="{ path: '/live', query: { session: String(next.session_key) } }"
    :title="startMsk"
  >
    <span class="body-sm next__what">
      <span class="next__prefix">Далее: </span>{{ next.country_name }} · {{ sessionLabel(next.session_name) }}
    </span>
    <span class="display-sm next__left">{{ left }}</span>
  </RouterLink>
</template>

<style scoped>
.next {
  display: grid;
  justify-items: center;
  padding: var(--space-1) var(--space-3);
  border: var(--border-thin) solid var(--border);
  background: var(--surface-raised);
  color: var(--text);
  text-decoration: none;
  white-space: nowrap;
  line-height: 1.2;
}

.next:hover {
  background: var(--surface-raised);
  color: var(--text);
}

.next__what {
  color: var(--text-muted);
}

@media (max-width: 639px) {
  .next {
    padding-inline: var(--space-2);
  }

  .next__what {
    font-size: 12px;
  }

  .next__prefix {
    display: none;
  }
}
</style>
