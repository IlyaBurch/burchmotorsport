<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useIntervalFn } from '@vueuse/core'
import { BmCard, BmChip } from '@/shared/ui'
import {
  FIRST_SEASON,
  compoundLetter,
  fetchLive,
  fetchMeetings,
  fetchSessions,
  formatGap,
  formatLap,
  formatSector,
  hasStarted,
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
// Only user actions navigate. Loading a session just syncs the pickers.
const seasons = Array.from(
  { length: new Date().getFullYear() - FIRST_SEASON + 1 },
  (_, i) => FIRST_SEASON + i,
).reverse()

const year = ref<number | null>(null)
const meetingKey = ref<number | null>(null)
const meetings = ref<Meeting[]>([])
const sessions = ref<Session[]>([])

const loadMeetings = async (y: number) => (meetings.value = (await fetchMeetings(y)).filter(hasStarted))
const loadSessions = async (mk: number) => (sessions.value = (await fetchSessions(mk)).filter(hasStarted))
const goTo = (key: number) => router.replace({ query: { session: String(key) } })

async function onYear(e: Event) {
  year.value = Number((e.target as HTMLSelectElement).value)
  await loadMeetings(year.value)
  const last = meetings.value.at(-1)
  if (last) await onMeeting(last.meeting_key)
}

async function onMeeting(mk: number) {
  meetingKey.value = mk
  await loadSessions(mk)
  const last = sessions.value.at(-1)
  if (last) goTo(last.session_key)
}

watch(live, async (l) => {
  if (!l) return
  const y = new Date(l.session.date_start).getFullYear()
  if (y !== year.value) {
    year.value = y
    await loadMeetings(y)
  }
  if (l.session.meeting_key !== meetingKey.value) {
    meetingKey.value = l.session.meeting_key
    await loadSessions(meetingKey.value)
  }
})

// --- derived -------------------------------------------------------------
const finished = computed(() => !!live.value && isFinished(live.value.session))
const leaderLap = computed(() => live.value?.drivers[0]?.lap ?? 0)

const updated = computed(() =>
  live.value
    ? new Date(live.value.updatedAt).toLocaleTimeString('ru-RU', { timeZone: 'Europe/Moscow' }) + ' МСК'
    : '',
)

const min = (xs: (number | null | undefined)[]) => {
  const v = xs.filter((x): x is number => x != null)
  return v.length ? Math.min(...v) : null
}
const overallBestLap = computed(() => min(live.value?.drivers.map((d) => d.bestLap) ?? []))
const overallBestSectors = computed(() =>
  [0, 1, 2].map((i) => min(live.value?.drivers.map((d) => d.bestSectors[i]!) ?? [])),
)

type Tone = 'purple' | 'green' | 'default'
const sectorTone = (d: Live['drivers'][number], i: number): Tone => {
  const s = d.sectors[i]
  if (s == null) return 'default'
  if (s === overallBestSectors.value[i]) return 'purple'
  if (s === d.bestSectors[i]) return 'green'
  return 'default'
}

const trackFlag = computed(() => {
  const f = live.value?.raceControl.find((m) => m.category === 'Flag' || m.category === 'SafetyCar')
  return f?.flag ?? null
})
const flagTone = (flag: string | null): Tone | 'yellow' | 'red' =>
  flag === 'RED' ? 'red' : flag && /YELLOW|SC|VSC/.test(flag) ? 'yellow' : flag === 'CHEQUERED' ? 'purple' : 'green'
</script>

<template>
  <section class="live">
    <BmCard>
      <div class="live__title-row">
        <div>
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
        </div>
        <BmChip v-if="trackFlag" :variant="flagTone(trackFlag)">{{ trackFlag }}</BmChip>
      </div>

      <dl v-if="live?.weather" class="live__weather">
        <div><dt class="display-sm">Воздух</dt><dd class="timing">{{ live.weather.air_temperature }}°</dd></div>
        <div><dt class="display-sm">Трасса</dt><dd class="timing">{{ live.weather.track_temperature }}°</dd></div>
        <div><dt class="display-sm">Влажн.</dt><dd class="timing">{{ live.weather.humidity }}%</dd></div>
        <div><dt class="display-sm">Ветер</dt><dd class="timing">{{ live.weather.wind_speed }} м/с</dd></div>
        <div><dt class="display-sm">Дождь</dt><dd class="timing">{{ live.weather.rainfall ? 'Да' : 'Нет' }}</dd></div>
      </dl>

      <div class="live__pickers">
        <label>
          <span class="display-sm">Сезон</span>
          <select class="live__select body" :value="year ?? ''" @change="onYear">
            <option v-for="y in seasons" :key="y" :value="y">{{ y }}</option>
          </select>
        </label>
        <label>
          <span class="display-sm">Гран-при</span>
          <select
            class="live__select body"
            :disabled="!meetings.length"
            :value="meetingKey ?? ''"
            @change="onMeeting(Number(($event.target as HTMLSelectElement).value))"
          >
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
            :value="live?.session.session_key ?? ''"
            @change="goTo(Number(($event.target as HTMLSelectElement).value))"
          >
            <option v-for="s in sessions" :key="s.session_key" :value="s.session_key">
              {{ s.session_name }}
            </option>
          </select>
        </label>
      </div>
    </BmCard>

    <div class="live__table-wrap">
      <table class="live__table">
        <thead>
          <tr class="display-sm">
            <th>P</th>
            <th>Пилот</th>
            <th>Шины</th>
            <th class="live__num">Пит</th>
            <th class="live__num">Отрыв</th>
            <th class="live__num">Интервал</th>
            <th class="live__num">Круг</th>
            <th class="live__num">S1</th>
            <th class="live__num">S2</th>
            <th class="live__num">S3</th>
            <th class="live__num">Лучший</th>
            <th class="live__num">Трап</th>
          </tr>
        </thead>
        <tbody v-if="live">
          <tr v-for="d in live.drivers" :key="d.number">
            <td class="display-sm">{{ d.position || '—' }}</td>
            <td class="body-strong">
              {{ d.acronym }}
              <span class="body-sm live__name">{{ d.name }}</span>
            </td>
            <td class="timing">
              <template v-if="d.compound">{{ compoundLetter(d.compound) }} {{ d.tyreAge }}</template>
              <template v-else>—</template>
            </td>
            <td class="timing live__num">{{ d.pits }}</td>
            <td class="timing live__num">{{ formatGap(d.gap) }}</td>
            <td class="timing live__num">{{ formatGap(d.interval) }}</td>
            <td class="timing live__num">{{ formatLap(d.lastLap) }}</td>
            <td v-for="i in [0, 1, 2]" :key="i" class="timing live__num">
              <BmChip v-if="sectorTone(d, i) !== 'default'" :variant="sectorTone(d, i)">
                {{ formatSector(d.sectors[i]) }}
              </BmChip>
              <template v-else>{{ formatSector(d.sectors[i]) }}</template>
            </td>
            <td class="timing live__num">
              <BmChip v-if="d.bestLap != null && d.bestLap === overallBestLap" variant="purple">
                {{ formatLap(d.bestLap) }}
              </BmChip>
              <template v-else>{{ formatLap(d.bestLap) }}</template>
            </td>
            <td class="timing live__num">{{ d.speedTrap || '—' }}</td>
          </tr>
        </tbody>
        <tbody v-else aria-busy="true" aria-label="Загрузка">
          <tr v-for="i in 20" :key="i">
            <td colspan="12"><div class="live__skeleton" /></td>
          </tr>
        </tbody>
      </table>
    </div>

    <BmCard v-if="live?.raceControl.length">
      <div class="bm-card__eyebrow">Race control</div>
      <ol class="live__rc">
        <li v-for="m in live.raceControl.slice(0, 8)" :key="m.date + m.message">
          <span class="timing-sm">L{{ m.lap_number || '—' }}</span>
          <span class="body">{{ m.message }}</span>
        </li>
      </ol>
    </BmCard>
  </section>
</template>

<style scoped>
.live {
  display: grid;
  gap: var(--space-6);
}

.live__title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
}

.live__meta {
  margin-top: var(--space-2);
}

.live__weather {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-6);
  margin: var(--space-4) 0 0;
  padding-top: var(--space-4);
  border-top: var(--border-thin) solid var(--border);
}

.live__weather dt {
  color: var(--text-muted);
}

.live__weather dd {
  margin: 0;
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
  padding: var(--space-2) var(--space-3);
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

.live__name {
  display: none;
  margin-left: var(--space-2);
}

@media (min-width: 1440px) {
  .live__name {
    display: inline;
  }
}

.live__skeleton {
  height: 24px;
  background: var(--surface-raised);
  animation: live-blink 1s steps(2) infinite;
}

@keyframes live-blink {
  to {
    background: var(--surface-sunken);
  }
}

@media (prefers-reduced-motion: reduce) {
  .live__skeleton {
    animation: none;
  }
}

.live__rc {
  list-style: none;
  margin: var(--space-3) 0 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.live__rc li {
  display: grid;
  grid-template-columns: 48px 1fr;
  gap: var(--space-3);
  align-items: baseline;
}
</style>
