<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useIntervalFn } from '@vueuse/core'
import { BmCard, BmChip } from '@/shared/ui'
import {
  FIRST_SEASON,
  compoundLetter,
  fetchLive,
  fetchMeetings,
  fetchSessions,
  fetchTrack,
  flagClass,
  formatGap,
  formatLap,
  formatSector,
  formatTime,
  hasStarted,
  isFinished,
  penaltiesFrom,
  projectStandings,
  projectTeams,
  sessionLabel,
  trackLimitsFrom,
  type Live,
  type Meeting,
  type Session,
} from '@/shared/api/live'
import TrackMap from './TrackMap.vue'
import DriverPlate from './DriverPlate.vue'

const route = useRoute()
const router = useRouter()

const live = ref<Live | null>(null)
const outline = ref<[number, number][]>([])
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
  async (key) => {
    live.value = null
    outline.value = []
    fetchTrack(key).then((o) => (outline.value = o)).catch(() => {}) // map is optional
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
  const last = meetings.value[meetings.value.length - 1]
  if (last) await onMeeting(last.meeting_key)
}

async function onMeeting(mk: number) {
  meetingKey.value = mk
  await loadSessions(mk)
  const last = sessions.value[sessions.value.length - 1]
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

const byNumber = computed(() => new Map(live.value?.drivers.map((d) => [d.number, d]) ?? []))
const plate = (car: number) => ({
  label: byNumber.value.get(car)?.acronym ?? String(car),
  colour: byNumber.value.get(car)?.teamColour,
})

const sessionName = computed(() => live.value?.session.session_name ?? '')
const scoring = computed(() => ['Race', 'Sprint'].includes(sessionName.value))
const standings = computed(() =>
  live.value
    ? [...projectStandings(sessionName.value, live.value.drivers).values()].sort((a, b) => a.pos - b.pos)
    : [],
)
const teams = computed(() => (live.value ? projectTeams(sessionName.value, live.value.drivers, live.value.teams) : []))
const fmtDelta = (n: number) => (n > 0 ? `▲${n}` : n < 0 ? `▼${-n}` : '')

const penalties = computed(() => penaltiesFrom(live.value?.raceControl ?? []))
const trackLimits = computed(() =>
  [...trackLimitsFrom(live.value?.raceControl ?? [])].sort((a, b) => b[1] - a[1]),
)

const trackFlag = computed(() => {
  const f = live.value?.raceControl.find((m) => m.category === 'Flag' || m.category === 'SafetyCar')
  return f?.flag ?? null
})
const flagTone = (flag: string | null): Tone | 'yellow' | 'red' =>
  flag === 'RED' ? 'red' : flag && /YELLOW|SC|VSC/.test(flag) ? 'yellow' : flag === 'CHEQUERED' ? 'purple' : 'green'

// flag banner under the header: shows on change, green/clear hides itself after 5s
const banner = ref<string | null>(null)
let bannerTimer: ReturnType<typeof setTimeout> | undefined
watch(
  () => trackFlag.value,
  (flag) => {
    clearTimeout(bannerTimer)
    banner.value = flag
    if (flag === 'GREEN' || flag === 'CLEAR') bannerTimer = setTimeout(() => (banner.value = null), 5000)
  },
)
onScopeDispose(() => clearTimeout(bannerTimer))
</script>

<template>
  <section class="live">
    <div v-if="finished" class="live__flag live__flag--chequered bm-checker display-md" role="status">
      <span>{{ sessionLabel(live!.session.session_name) }} завершена</span>
    </div>
    <div v-else-if="banner" class="live__flag display-md" :class="`live__flag--${flagTone(banner)}`" role="status">
      {{ banner }}
    </div>

    <BmCard class="live__head">
      <div class="live__title-row">
        <div>
          <div class="bm-card__eyebrow">{{ finished ? 'Архив' : 'Live' }}</div>
          <h1 class="bm-card__title">
            {{ live ? `${live.session.circuit_short_name} · ${sessionLabel(live.session.session_name)}` : 'Телеметрия' }}
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
              {{ sessionLabel(s.session_name) }}
            </option>
          </select>
        </label>
      </div>
    </BmCard>

    <div class="live__table-wrap">
      <table class="live__table">
        <thead>
          <tr class="display-sm">
            <th class="live__sticky live__sticky--pos">P</th>
            <th class="live__sticky live__sticky--drv">Пилот</th>
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
            <td class="display-sm live__sticky live__sticky--pos">{{ d.position || '—' }}</td>
            <td class="live__sticky live__sticky--drv">
              <DriverPlate :label="d.acronym" :colour="d.teamColour" />
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

    <aside class="live__side">
      <TrackMap v-if="outline.length && live" :outline="outline" :drivers="live.drivers" />
      <div v-else class="live__skeleton live__skeleton--map" aria-busy="true" aria-label="Загрузка карты" />

      <dl v-if="live?.weather" class="live__weather">
        <div><dt class="display-sm">Воздух</dt><dd class="timing">{{ live.weather.air_temperature }}°</dd></div>
        <div><dt class="display-sm">Трасса</dt><dd class="timing">{{ live.weather.track_temperature }}°</dd></div>
        <div><dt class="display-sm">Влажн.</dt><dd class="timing">{{ live.weather.humidity }}%</dd></div>
        <div><dt class="display-sm">Ветер</dt><dd class="timing">{{ live.weather.wind_speed }} м/с</dd></div>
        <div><dt class="display-sm">Дождь</dt><dd class="timing">{{ live.weather.rainfall ? 'Да' : 'Нет' }}</dd></div>
      </dl>
    </aside>

    <BmCard v-if="standings.length" class="live__drivers">
      <div class="bm-card__eyebrow">Личный зачёт{{ scoring ? ' · прогноз' : '' }}</div>
      <ol class="live__list">
        <li v-for="s in standings" :key="s.number" class="live__row">
          <span class="display-sm">{{ s.pos }}</span>
          <DriverPlate :label="byNumber.get(s.number)?.name ?? ''" :colour="plate(s.number).colour" wide />
          <span class="timing live__num">{{ s.points }}</span>
          <span class="timing-sm live__num live__gain">{{ s.gain ? `+${s.gain}` : '' }}</span>
          <span class="timing-sm live__num">{{ fmtDelta(s.delta) }}</span>
        </li>
      </ol>
    </BmCard>

    <BmCard v-if="teams.length" class="live__teams">
      <div class="bm-card__eyebrow">Кубок конструкторов{{ scoring ? ' · прогноз' : '' }}</div>
      <ol class="live__list">
        <li v-for="t in teams" :key="t.name" class="live__row">
          <span class="display-sm">{{ t.pos }}</span>
          <DriverPlate :label="t.name" :colour="t.colour" wide />
          <span class="timing live__num">{{ t.points }}</span>
          <span class="timing-sm live__num live__gain">{{ t.gain ? `+${t.gain}` : '' }}</span>
          <span class="timing-sm live__num">{{ fmtDelta(t.delta) }}</span>
        </li>
      </ol>
    </BmCard>

    <div class="live__pen">
      <BmCard>
        <div class="bm-card__eyebrow">Штрафы</div>
        <p v-if="!penalties.length" class="body-sm live__empty">Пока чисто</p>
        <ol v-else class="live__list">
          <li v-for="p in penalties" :key="p.key" class="live__penalty">
            <DriverPlate v-bind="plate(p.car)" />
            <span>
              <span class="body-strong">{{ p.what }}</span>
              <span v-if="p.why" class="body-sm"> · {{ p.why }}</span>
              <span class="timing-sm"> · L{{ p.lap }}</span>
            </span>
          </li>
        </ol>
      </BmCard>

      <BmCard>
        <div class="bm-card__eyebrow">Лимиты трассы</div>
        <p v-if="!trackLimits.length" class="body-sm live__empty">Ни одного удалённого круга</p>
        <div v-else class="live__limits">
          <span v-for="[car, n] in trackLimits" :key="car" class="live__limit">
            <DriverPlate v-bind="plate(car)" />
            <span class="timing">{{ n }}</span>
          </span>
        </div>
      </BmCard>
    </div>

    <BmCard v-if="live?.raceControl.length" class="live__rc">
      <div class="bm-card__eyebrow">Race control</div>
      <ol class="live__list live__feed">
        <li v-for="m in live.raceControl" :key="m.date + m.message" class="live__msg">
          <span class="timing-sm">L{{ m.lap_number || '—' }}</span>
          <i v-if="flagClass(m.flag)" class="live__sq" :class="`live__sq--${flagClass(m.flag)}`" aria-hidden="true" />
          <span class="body-sm">{{ m.message }}</span>
        </li>
      </ol>
    </BmCard>

    <BmCard v-if="live?.radio.length" class="live__radio">
      <div class="bm-card__eyebrow">Радио</div>
      <ol class="live__list live__feed">
        <li v-for="r in live.radio" :key="r.recording_url" class="live__radio-row">
          <span class="timing-sm">{{ formatTime(r.date) }}</span>
          <DriverPlate v-bind="plate(r.driver_number)" />
          <!-- ponytail: native player; transcript/translation hooks in here later -->
          <audio :src="r.recording_url" controls preload="none" class="live__audio" />
        </li>
      </ol>
    </BmCard>
  </section>
</template>

<style scoped>
/* mobile first, single column; desktop lays the blocks out in three columns */
.live {
  display: grid;
  gap: var(--space-4);
  grid-template-areas: 'flag' 'head' 'table' 'side' 'drivers' 'teams' 'pen' 'radio' 'rc';
}

.live__head { grid-area: head; }
.live__table-wrap { grid-area: table; }
.live__side { grid-area: side; display: grid; gap: var(--space-4); align-content: start; }
.live__drivers { grid-area: drivers; }
.live__teams { grid-area: teams; }
.live__pen { grid-area: pen; display: grid; gap: var(--space-4); align-content: start; }
.live__rc { grid-area: rc; }
.live__radio { grid-area: radio; }

@media (min-width: 1024px) {
  .live {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 360px;
    grid-template-areas:
      'flag flag flag'
      'head head head'
      'table table side'
      'drivers teams pen'
      'rc rc radio';
    gap: var(--space-6);
  }
}

.live__flag {
  grid-area: flag;
  padding: var(--space-3) var(--space-4);
  text-align: center;
  color: var(--ink);
  border: var(--border-thick) solid var(--border);
  box-shadow: var(--shadow-hard);
}

.live__flag--green { background: var(--timing-green); }
.live__flag--yellow { background: var(--timing-yellow); }
.live__flag--red { background: var(--flag-red); }
.live__flag--purple { background: var(--timing-purple); }
.live__flag--default { background: var(--surface-raised); color: var(--text); }

/* chequered: the brand checker pattern from bm.css, text on a solid plate so it stays readable */
.live__flag--chequered {
  padding: var(--space-3);
  background-size: 40px 40px; /* sector bar uses 12px, too busy at banner size */
}

.live__flag--chequered span {
  display: inline-block;
  padding: var(--space-2) var(--space-4);
  background: var(--paper);
  color: var(--ink);
  border: var(--border-thin) solid var(--border);
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
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(96px, 1fr));
  gap: var(--space-3);
  margin: 0;
  padding: var(--space-3) var(--space-4);
  background: var(--surface-raised);
  border: var(--border-thin) solid var(--border);
}

.live__weather dt {
  color: var(--text-muted);
}

.live__weather dd {
  margin: 0;
}

.live__pickers {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-3);
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: var(--border-thin) solid var(--border);
}

.live__pickers label {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

/* input recipe from the guide, applied to a native select */
.live__select {
  width: 100%;
  min-width: 0;
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
  border-collapse: separate;
  border-spacing: 0;
}

.live__table th,
.live__table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  white-space: nowrap;
  vertical-align: middle;
  border-bottom: 1px solid var(--border);
}

.live__table thead th {
  color: var(--text-muted);
  border-bottom: var(--border-thin) solid var(--border);
}

.live__table tbody tr:last-child td {
  border-bottom: 0;
}

/* P + driver stay put while the rest scrolls on narrow screens */
.live__sticky {
  position: sticky;
  background: var(--surface-sunken);
  z-index: 1;
}

.live__sticky--pos { left: 0; }
.live__sticky--drv { left: 44px; }

.live__sticky--drv::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  border-right: 1px solid var(--border);
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

.live__skeleton--map {
  aspect-ratio: 4 / 3;
  height: auto;
  border: var(--border-thick) solid var(--border);
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

/* lists inside cards */
.live__list {
  list-style: none;
  margin: var(--space-3) 0 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.live__feed {
  max-height: 360px;
  overflow-y: auto;
}

.live__empty {
  margin: var(--space-3) 0 0;
}

.live__row {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 48px 40px 36px;
  gap: var(--space-2);
  align-items: center;
}

.live__gain {
  color: var(--text-muted);
}

.live__penalty {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--space-3);
  align-items: baseline;
}

.live__limits {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-3);
}

.live__limit {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding-right: var(--space-2);
  border: var(--border-thin) solid var(--border);
}

.live__msg {
  display: grid;
  grid-template-columns: 44px auto 1fr;
  gap: var(--space-2);
  align-items: baseline;
}

.live__sq {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 1px solid var(--border);
  align-self: center;
}

.live__sq--yellow { background: var(--timing-yellow); }
.live__sq--green { background: var(--timing-green); }
.live__sq--red { background: var(--flag-red); }
.live__sq--blue { background: var(--accent); }
.live__sq--plain { background: var(--paper); }
.live__sq--chequered {
  background: conic-gradient(var(--ink) 0.25turn, var(--paper) 0.25turn 0.5turn, var(--ink) 0.5turn 0.75turn, var(--paper) 0.75turn);
  background-size: 7px 7px;
}

.live__radio-row {
  display: grid;
  grid-template-columns: 44px auto 1fr;
  gap: var(--space-2);
  align-items: center;
}

.live__audio {
  width: 100%;
  min-width: 0;
  height: 36px;
}
</style>
