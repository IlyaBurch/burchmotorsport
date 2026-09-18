<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useIntervalFn } from '@vueuse/core'
import { BmCard, BmChip } from '@/shared/ui'
import {
  FIRST_SEASON,
  fetchLive,
  fetchMeetings,
  fetchSessions,
  formatGap,
  formatLap,
  isFinished,
  type Live,
  type Meeting,
  type Session,
} from '@/shared/api/live'

const route = useRoute()
const router = useRouter()

const live = ref<Live | null>(null)
const error = ref<string | null>(null)

const sessionKey = computed(() => String(route.query.session ?? 'latest'))

async function refresh() {
  try {
    live.value = await fetchLive(sessionKey.value)
    error.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

const poll = useIntervalFn(refresh, 5000, { immediate: false })
watch(
  sessionKey,
  async () => {
    live.value = null
    await refresh()
    const l = live.value as Live | null
    if (l && !isFinished(l.session)) poll.resume()
    else poll.pause()
  },
  { immediate: true },
)

// --- pickers -------------------------------------------------------------
const seasons = Array.from(
  { length: new Date().getFullYear() - FIRST_SEASON + 1 },
  (_, i) => FIRST_SEASON + i,
).reverse()

const year = ref<number | null>(null)
const meetingKey = ref<number | null>(null)
const meetings = ref<Meeting[]>([])
const sessions = ref<Session[]>([])

watch(year, async (y) => {
  meetings.value = y ? await fetchMeetings(y) : []
  if (!meetings.value.some((m) => m.meeting_key === meetingKey.value)) meetingKey.value = null
})
watch(meetingKey, async (mk) => {
  sessions.value = mk ? await fetchSessions(mk) : []
})

// sync pickers to whatever session is on screen (deep links, "latest")
watch(live, (l) => {
  if (!l) return
  year.value = new Date(l.session.date_start).getFullYear()
  meetingKey.value = l.session.meeting_key
})

function pickSession(e: Event) {
  const v = (e.target as HTMLSelectElement).value
  router.replace({ query: v === 'latest' ? {} : { session: v } })
}

// --- derived -------------------------------------------------------------
const bestLapNumber = computed(() => {
  const d = live.value?.drivers.filter((x) => x.bestLap != null) ?? []
  return d.length ? d.reduce((a, b) => (a.bestLap! <= b.bestLap! ? a : b)).number : null
})

const leaderLap = computed(() => live.value?.drivers[0]?.lap ?? 0)

const updated = computed(() =>
  live.value
    ? new Date(live.value.updatedAt).toLocaleTimeString('ru-RU', { timeZone: 'Europe/Moscow' }) + ' МСК'
    : '',
)

const finished = computed(() => !!live.value && isFinished(live.value.session))
</script>

<template>
  <section class="live">
    <BmCard class="live__head">
      <div class="bm-card__eyebrow">{{ finished ? 'Архив' : 'Live' }}</div>
      <h1 class="bm-card__title">
        {{ live ? `${live.session.circuit_short_name} · ${live.session.session_name}` : 'Телеметрия' }}
      </h1>
      <p class="body-sm live__meta">
        <template v-if="live && finished">{{ leaderLap }} кругов · сессия завершена</template>
        <template v-else-if="live">Круг {{ leaderLap }} · обновлено {{ updated }}</template>
        <template v-else-if="error">Нет связи с данными</template>
        <template v-else>Загрузка</template>
      </p>
      <BmChip v-if="error" variant="red">{{ error }}</BmChip>

      <div class="live__pickers">
        <label>
          <span class="display-sm">Сезон</span>
          <select v-model="year" class="live__select body">
            <option v-for="y in seasons" :key="y" :value="y">{{ y }}</option>
          </select>
        </label>
        <label>
          <span class="display-sm">Гран-при</span>
          <select v-model="meetingKey" class="live__select body" :disabled="!meetings.length">
            <option v-for="m in meetings" :key="m.meeting_key" :value="m.meeting_key">
              {{ m.meeting_name }}
            </option>
          </select>
        </label>
        <label>
          <span class="display-sm">Сессия</span>
          <select
            class="live__select body"
            :disabled="!sessions.length"
            :value="sessionKey"
            @change="pickSession"
          >
            <option value="latest">Текущая</option>
            <option v-for="s in sessions" :key="s.session_key" :value="String(s.session_key)">
              {{ s.session_name }}
            </option>
          </select>
        </label>
      </div>
    </BmCard>

    <div v-if="live" class="live__table-wrap">
      <table class="live__table">
        <thead>
          <tr class="display-sm">
            <th>P</th>
            <th>Пилот</th>
            <th class="live__hide-sm">Команда</th>
            <th class="live__num">Отрыв</th>
            <th class="live__num live__hide-sm">Интервал</th>
            <th class="live__num">{{ finished ? 'Последний' : 'Круг' }}</th>
            <th class="live__num live__hide-sm">Лучший</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in live.drivers" :key="d.number">
            <td class="display-sm">{{ d.position || '—' }}</td>
            <td class="body-strong">
              {{ d.acronym }}
              <span class="body-sm live__hide-sm">{{ d.name }}</span>
            </td>
            <td class="body-sm live__hide-sm">{{ d.team }}</td>
            <td class="timing live__num">{{ formatGap(d.gap) }}</td>
            <td class="timing live__num live__hide-sm">{{ formatGap(d.interval) }}</td>
            <td class="timing live__num">{{ formatLap(d.lastLap) }}</td>
            <td class="timing live__num live__hide-sm">
              <BmChip v-if="d.number === bestLapNumber" variant="purple">{{ formatLap(d.bestLap) }}</BmChip>
              <template v-else>{{ formatLap(d.bestLap) }}</template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.live {
  display: grid;
  gap: var(--space-6);
}

.live__meta {
  margin-top: var(--space-2);
}

.live__pickers {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--space-4);
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: var(--border-thin) solid var(--border);
}

.live__pickers label {
  display: grid;
  gap: var(--space-2);
}

/* input recipe from the guide, applied to a native select */
.live__select {
  min-height: 44px;
  padding: var(--space-3) var(--space-4);
  background: var(--surface-raised);
  color: var(--text);
  border: var(--border-thin) solid var(--border);
  border-radius: var(--radius-sm);
}

.live__select:disabled {
  background: var(--surface-sunken);
  color: var(--text-muted);
}

.live__table-wrap {
  background: var(--surface-sunken);
  border: var(--border-thick) solid var(--border);
  box-shadow: var(--shadow-hard);
  overflow-x: auto;
}

.live__table {
  width: 100%;
  border-collapse: collapse;
}

.live__table th,
.live__table td {
  padding: var(--space-3) var(--space-4);
  text-align: left;
  white-space: nowrap;
  vertical-align: middle;
}

.live__table thead th {
  color: var(--text-muted);
  border-bottom: var(--border-thin) solid var(--border);
}

.live__table tbody tr + tr td {
  border-top: 1px solid var(--border);
}

.live__num {
  text-align: right;
}

.live__hide-sm {
  display: none;
}

@media (min-width: 768px) {
  .live__hide-sm {
    display: table-cell;
  }

  span.live__hide-sm {
    display: inline;
    margin-left: var(--space-2);
  }
}
</style>
