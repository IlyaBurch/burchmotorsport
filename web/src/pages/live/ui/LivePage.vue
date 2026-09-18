<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useIntervalFn } from '@vueuse/core'
import { BmCard, BmChip } from '@/shared/ui'
import {
  FIRST_SEASON,
  NoDataError,
  RateLimitError,
  compoundLetter,
  countdown,
  fetchLive,
  fetchMeetings,
  fetchSessions,
  fetchTrack,
  findPreviousRace,
  flagClass,
  formatGap,
  formatLap,
  formatMsk,
  formatSector,
  formatTime,
  inkOn,
  isFinished,
  penaltiesFrom,
  projectStandings,
  projectTeams,
  radioSrc,
  sessionLabel,
  trackLimitsFrom,
  type Live,
  type Meeting,
  type Session,
} from '@/shared/api/live'
import TrackMap from './TrackMap.vue'
import DriverPlate from './DriverPlate.vue'
import PositionsChart from './PositionsChart.vue'
import TyreChart from './TyreChart.vue'

const route = useRoute()
const router = useRouter()

const live = ref<Live | null>(null)
const outline = ref<[number, number][]>([])
const error = ref<string | null>(null)
const noData = ref(false)
const previous = ref<Live | null>(null) // last year's race here, for upcoming previews
const now = ref(Date.now())
useIntervalFn(() => (now.value = Date.now()), 30_000)

const sessionKey = computed(() => String(route.query.session ?? 'latest'))

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
onScopeDispose(() => clearTimeout(retryTimer))

const poll = useIntervalFn(refresh, 10_000, { immediate: false })
watch(
  sessionKey,
  async (key) => {
    live.value = null
    outline.value = []
    previous.value = null
    noData.value = false
    await refresh()
    const l = live.value as Live | null
    if (l && !l.upcoming) fetchTrack(key).then((o) => (outline.value = o)).catch(() => {}) // map is optional
    if (l?.upcoming) loadPreview(l)
    if (l && !l.upcoming && !isFinished(l.session)) poll.resume()
    else poll.pause()
  },
  { immediate: true },
)

// --- upcoming session preview: last year's race here, its map and podium ----
async function loadPreview(l: Live) {
  try {
    const race = await findPreviousRace(l.session.circuit_short_name, new Date(l.session.date_start).getFullYear())
    if (!race) return
    const [prev, o] = await Promise.all([fetchLive(race.session_key), fetchTrack(race.session_key).catch(() => [])])
    if (live.value?.session.session_key !== l.session.session_key) return // user moved on
    previous.value = prev
    outline.value = o
  } catch {
    /* preview is optional */
  }
}
const weekend = computed(() =>
  live.value?.upcoming
    ? sessions.value.map((x) => ({ ...x, started: Date.parse(x.date_start) <= now.value }))
    : [],
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

const loadMeetings = async (y: number) => (meetings.value = await fetchMeetings(y))
const loadSessions = async (mk: number) => (sessions.value = await fetchSessions(mk))
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
const finished = computed(() => !!live.value && !live.value.upcoming && isFinished(live.value.session))
const upcoming = computed(() => !!live.value?.upcoming)
const loading = computed(() => !live.value && !error.value && !noData.value)
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

const bestTone = (d: Live['drivers'][number], i: number): Tone =>
  d.bestSectors[i] != null && d.bestSectors[i] === overallBestSectors.value[i] ? 'purple' : 'default'

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
const fmtDelta = (n: number) => (n > 0 ? `▲${n}` : n < 0 ? `▼${-n}` : '—')

const bestSectorHolders = computed(() =>
  [0, 1, 2].map((i) => {
    const best = overallBestSectors.value[i]
    const d = live.value?.drivers.find((x) => x.bestSectors[i] === best)
    return { i, best, d }
  }),
)
const bestLapHolder = computed(() =>
  live.value?.drivers.find((d) => d.bestLap != null && d.bestLap === overallBestLap.value) ?? null,
)
const idealLap = computed(() => {
  const s = overallBestSectors.value
  return s.every((x) => x != null) ? s.reduce((a, b) => a! + b!, 0) : null
})

const laps = computed(() => Math.max(0, ...(live.value?.drivers.map((d) => d.positions.length) ?? [])))

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
            <template v-else-if="upcoming">Старт {{ formatMsk(live!.session.date_start) }}</template>
            <template v-else-if="live">Круг {{ leaderLap }} · обновлено {{ updated }}</template>
            <template v-else-if="noData">Данных нет</template>
            <template v-else-if="error">Ошибка</template>
            <template v-else>Загрузка</template>
          </p>
          <BmChip v-if="error" variant="red">{{ error }}</BmChip>
        </div>
        <BmChip v-if="trackFlag" :variant="flagTone(trackFlag)">{{ trackFlag }}</BmChip>
      </div>

      <details class="live__details">
        <summary class="display-sm live__summary">Другая сессия</summary>
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
      </details>
    </BmCard>

    <section v-if="upcoming" class="live__status live__preview">
      <BmCard class="live__preview-main">
        <div class="bm-card__eyebrow">До старта</div>
        <p class="display-xl live__countdown">{{ countdown(Date.parse(live!.session.date_start) - now) }}</p>
        <p class="body-lg">
          {{ sessionLabel(live!.session.session_name) }} · {{ live!.session.circuit_short_name }}, {{ live!.session.country_name }}
        </p>
        <p class="body">Старт {{ formatMsk(live!.session.date_start) }}</p>

        <div v-if="weekend.length" class="live__weekend">
          <div class="display-sm live__weekend-title">Уикенд</div>
          <ol class="live__list">
            <li v-for="x in weekend" :key="x.session_key" class="live__weekend-row" :class="{ 'live__weekend-row--current': x.session_key === live!.session.session_key }">
              <RouterLink class="body-strong live__weekend-link" :to="{ path: '/live', query: { session: String(x.session_key) } }">
                {{ sessionLabel(x.session_name) }}
              </RouterLink>
              <span class="timing-sm">{{ formatMsk(x.date_start) }}</span>
              <BmChip v-if="x.started" variant="green">Была</BmChip>
            </li>
          </ol>
        </div>
      </BmCard>

      <div class="live__preview-side">
        <TrackMap v-if="outline.length" :outline="outline" :drivers="[]" />
        <div v-else class="live__skeleton live__skeleton--map" aria-busy="true" aria-label="Загрузка карты" />

        <BmCard v-if="previous">
          <div class="bm-card__eyebrow">В прошлом году</div>
          <p class="body-sm">{{ previous.session.circuit_short_name }} · {{ new Date(previous.session.date_start).getFullYear() }}</p>
          <ol class="live__list">
            <li v-for="d in previous.drivers.slice(0, 3)" :key="d.number" class="live__podium">
              <span class="display-sm">{{ d.position }}</span>
              <DriverPlate :label="d.acronym" :colour="d.teamColour" />
              <span class="body-sm">{{ d.name }}</span>
              <span class="timing live__num">{{ formatGap(d.gap) }}</span>
            </li>
          </ol>
          <RouterLink class="bm-btn bm-btn--ghost live__preview-link" :to="{ path: '/live', query: { session: String(previous.session.session_key) } }">
            Вся гонка
          </RouterLink>
        </BmCard>
      </div>
    </section>

    <BmCard v-else-if="noData" class="live__status">
      <div class="bm-card__eyebrow">Box box</div>
      <h2 class="display-md">Данные потерялись</h2>
      <p class="body">Сессия была, а тайминга по ней у нас нет. Извини, разбираемся.</p>
    </BmCard>

    <!-- mobile: one card per driver instead of a 12-column scrolling table -->
    <ol v-if="live && !upcoming" class="live__cards" aria-label="Пилоты">
      <li v-for="d in live.drivers" :key="d.number" class="live__card">
        <div class="live__card-id" :style="d.teamColour ? { background: '#' + d.teamColour, color: inkOn(d.teamColour) } : undefined">
          <span class="display-md">{{ d.position || '—' }}</span>
          <span class="body-strong">{{ d.acronym }}</span>
        </div>
        <div class="live__card-body">
          <div class="live__cell">
            <span class="timing">{{ formatGap(d.interval) }}</span>
            <span class="timing-sm live__muted">{{ formatGap(d.gap) }}</span>
          </div>
          <div class="live__cell">
            <span class="timing">{{ d.compound ? `${compoundLetter(d.compound)} ${d.tyreAge}` : '—' }}</span>
            <span class="timing-sm live__muted">{{ d.pits }} pit</span>
          </div>
          <div class="live__cell">
            <BmChip v-if="d.bestLap != null && d.bestLap === overallBestLap" variant="purple">{{ formatLap(d.lastLap) }}</BmChip>
            <span v-else class="timing">{{ formatLap(d.lastLap) }}</span>
            <span class="timing-sm live__muted">{{ formatLap(d.bestLap) }}</span>
          </div>
          <div class="live__cell live__cell--laps">
            <span class="timing">{{ d.lap }}</span>
            <span class="timing-sm live__muted">laps</span>
          </div>
          <div v-for="i in [0, 1, 2]" :key="i" class="live__cell live__cell--sector">
            <BmChip v-if="sectorTone(d, i) !== 'default'" :variant="sectorTone(d, i)">{{ formatSector(d.sectors[i]) }}</BmChip>
            <span v-else class="timing-sm">{{ formatSector(d.sectors[i]) }}</span>
            <BmChip v-if="bestTone(d, i) !== 'default'" :variant="bestTone(d, i)">{{ formatSector(d.bestSectors[i]) }}</BmChip>
            <span v-else class="timing-sm live__muted">{{ formatSector(d.bestSectors[i]) }}</span>
          </div>
        </div>
      </li>
    </ol>

    <div v-if="loading || (live && !upcoming)" class="live__table-wrap">
      <table class="live__table">
        <thead>
          <tr class="display-sm">
            <th class="live__sticky live__sticky--pos">P</th>
            <th class="live__sticky live__sticky--drv">Пилот</th>
            <th>Шины</th>
            <th class="live__num">Пит</th>
            <th class="live__num">Отрыв</th>
            <th class="live__num">Интервал</th>
            <th class="live__num" title="Последний круг">Круг</th>
            <th class="live__num" title="Сектор последнего круга">S1</th>
            <th class="live__num" title="Сектор последнего круга">S2</th>
            <th class="live__num" title="Сектор последнего круга">S3</th>
            <th class="live__num">Лучший</th>
            <th class="live__num">Скорость</th>
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

    <aside v-if="live && !upcoming" class="live__side">
      <TrackMap v-if="outline.length && live" :outline="outline" :drivers="live.drivers" />
      <div v-else class="live__skeleton live__skeleton--map" aria-busy="true" aria-label="Загрузка карты" />

      <dl v-if="live?.weather" class="live__weather">
        <div><dt class="display-sm">Воздух</dt><dd class="timing">{{ live.weather.air_temperature }}°</dd></div>
        <div><dt class="display-sm">Трасса</dt><dd class="timing">{{ live.weather.track_temperature }}°</dd></div>
        <div><dt class="display-sm">Влажн.</dt><dd class="timing">{{ live.weather.humidity }}%</dd></div>
        <div><dt class="display-sm">Ветер</dt><dd class="timing">{{ live.weather.wind_speed }} м/с</dd></div>
        <div><dt class="display-sm">Дождь</dt><dd class="timing">{{ live.weather.rainfall ? 'Да' : 'Нет' }}</dd></div>
      </dl>

      <BmCard v-if="live" class="live__sectors-card">
        <div class="bm-card__eyebrow">Лучшие за сессию</div>
        <dl class="live__sectors">
          <div v-for="b in bestSectorHolders" :key="b.i" class="live__sector">
            <dt class="display-sm">S{{ b.i + 1 }}</dt>
            <dd><DriverPlate v-if="b.d" :label="b.d.acronym" :colour="b.d.teamColour" /></dd>
            <dd>
              <BmChip v-if="b.best != null" variant="purple">{{ formatSector(b.best) }}</BmChip>
              <span v-else class="body-sm">—</span>
            </dd>
          </div>
          <div class="live__sector">
            <dt class="display-sm">Круг</dt>
            <dd><DriverPlate v-if="bestLapHolder" :label="bestLapHolder.acronym" :colour="bestLapHolder.teamColour" /></dd>
            <dd>
              <BmChip v-if="bestLapHolder" variant="purple">{{ formatLap(bestLapHolder.bestLap) }}</BmChip>
              <span v-else class="body-sm">—</span>
            </dd>
          </div>
          <div class="live__sector">
            <dt class="display-sm" title="Сумма лучших секторов">Сумма</dt>
            <dd></dd>
            <dd class="timing">{{ formatLap(idealLap) }}</dd>
          </div>
        </dl>
      </BmCard>

    </aside>

    <BmCard v-if="live && laps > 1" class="live__chart">
      <div class="bm-card__eyebrow">Позиции по кругам</div>
      <PositionsChart :drivers="live.drivers" />
    </BmCard>

    <BmCard v-if="live && laps > 1" class="live__tyres">
      <div class="bm-card__eyebrow">Стратегия шин</div>
      <TyreChart :drivers="live.drivers" />
    </BmCard>

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

    <div v-if="live && !upcoming" class="live__pen">
      <BmCard class="live__pen-card">
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

      <BmCard class="live__pen-card">
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
      <div class="live__rc-scroll">
        <table class="live__rc-table">
          <thead>
            <tr class="display-sm">
              <th>Круг</th>
              <th>Флаг</th>
              <th>Время</th>
              <th>Сообщение</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in live.raceControl" :key="m.date + m.message">
              <td class="timing-sm">L{{ m.lap_number || '—' }}</td>
              <td><i v-if="flagClass(m.flag)" class="live__sq" :class="`live__sq--${flagClass(m.flag)}`" aria-hidden="true" /></td>
              <td class="timing-sm">{{ formatTime(m.date) }}</td>
              <td class="body-sm live__rc-msg">{{ m.message }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </BmCard>

    <BmCard v-if="live?.radio.length" class="live__radio">
      <div class="bm-card__eyebrow">Радио</div>
      <ol class="live__list live__feed">
        <li v-for="r in live.radio" :key="r.recording_url" class="live__radio-row">
          <span class="timing-sm">{{ formatTime(r.date) }}</span>
          <DriverPlate v-bind="plate(r.driver_number)" />
          <!-- ponytail: native player; transcript/translation hooks in here later -->
          <audio :src="radioSrc(r.recording_url)" controls preload="none" class="live__audio" />
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
  grid-template-areas: 'flag' 'head' 'status' 'cards' 'side' 'chart' 'tyres' 'drivers' 'teams' 'pen' 'rc' 'radio';
}

/* grid items default to min-width:auto and would grow to the widest chart; keep them inside the viewport */
.live > * {
  min-width: 0;
}

.live__head { grid-area: head; }
.live__status { grid-area: status; }

/* upcoming session preview */
.live__preview {
  display: grid;
  gap: var(--space-4);
}

@media (min-width: 1024px) {
  .live__preview {
    grid-template-columns: minmax(0, 1fr) 360px;
    gap: var(--space-6);
  }
}

.live__preview-side {
  display: grid;
  gap: var(--space-4);
  align-content: start;
}

.live__countdown {
  margin: var(--space-3) 0;
}

.live__weekend {
  margin-top: var(--space-6);
  padding-top: var(--space-4);
  border-top: var(--border-thin) solid var(--border);
}

.live__weekend-title {
  color: var(--text-muted);
}

.live__weekend-row {
  max-width: none;
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: var(--space-3);
  align-items: center;
}

.live__weekend-row--current .live__weekend-link {
  background: var(--accent);
  color: var(--on-accent);
}

.live__weekend-link {
  padding: 0 var(--space-2);
}

.live__podium {
  max-width: none;
  display: grid;
  grid-template-columns: 24px 56px 1fr auto;
  gap: var(--space-2);
  align-items: center;
}

.live__preview-link {
  margin-top: var(--space-4);
}
.live__cards { grid-area: cards; }
.live__table-wrap { grid-area: table; display: none; }

@media (min-width: 768px) {
  .live__cards { display: none; }
  .live__table-wrap { display: block; }
}

/* driver cards (mobile) */
.live__cards {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.live__card {
  max-width: none;
  display: grid;
  grid-template-columns: 64px 1fr;
  background: var(--surface-sunken);
  border: var(--border-thin) solid var(--border);
}

.live__card-id {
  display: grid;
  place-content: center;
  text-align: center;
  background: var(--surface-raised);
  border-right: var(--border-thin) solid var(--border);
}

.live__card-body {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1.3fr 0.6fr;
  gap: 1px;
  background: var(--border);
}

.live__cell {
  display: grid;
  align-content: center;
  justify-items: center;
  gap: 2px;
  padding: var(--space-1);
  background: var(--surface-sunken);
  text-align: center;
}

.live__cell--laps {
  background: var(--surface-raised);
}

.live__cell--sector {
  grid-column: span 1;
}

.live__card-body .live__cell--sector:nth-of-type(5) {
  grid-column: 1 / 2;
}

.live__card-body .live__cell--sector:nth-of-type(7) {
  grid-column: 3 / 5;
}

.live__muted {
  color: var(--text-muted);
}
.live__side { grid-area: side; display: flex; flex-direction: column; gap: var(--space-4); }
.live__chart { grid-area: chart; }
.live__tyres { grid-area: tyres; }
.live__drivers { grid-area: drivers; }
.live__teams { grid-area: teams; }
.live__pen {
  grid-area: pen;
  display: grid;
  grid-template-rows: 1fr 1fr; /* two equal cards, as tall as the standings row */
  gap: var(--space-4);
  min-height: 0;
}

.live__pen-card {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.live__pen-card > .live__list,
.live__pen-card > .live__limits {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  align-content: start;
}

.live__rc { grid-area: rc; }

.live__radio { grid-area: radio; }

@media (min-width: 768px) and (max-width: 1023px) {
  .live {
    grid-template-areas: 'flag' 'head' 'status' 'table' 'side' 'chart' 'tyres' 'drivers' 'teams' 'pen' 'rc' 'radio';
  }
}

@media (min-width: 1024px) {
  .live {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 360px;
    grid-template-areas:
      'flag flag flag'
      'head head head'
      'status status status'
      'table table side'
      'chart chart chart'
      'tyres tyres tyres'
      'drivers teams pen'
      'rc rc rc'
      'radio radio radio';
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

.live__sectors-card {
  flex: 1; /* fills the side column down to the table's bottom edge */
  display: flex;
  flex-direction: column;
}

.live__sectors {
  flex: 1;
  display: grid;
  align-content: space-evenly;
  gap: var(--space-3);
  margin: var(--space-3) 0 0;
}

.live__sector {
  display: grid;
  grid-template-columns: 64px 56px 1fr;
  gap: var(--space-3);
  align-items: center;
}

.live__sector dd {
  margin: 0;
}

.live__sector dd:last-child {
  justify-self: end;
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

/* session pickers fold away, native <details> */
.live__details {
  margin-top: var(--space-4);
  border-top: var(--border-thin) solid var(--border);
}

.live__summary {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 44px;
  cursor: pointer;
  color: var(--accent);
  list-style: none;
}

.live__summary::-webkit-details-marker {
  display: none;
}

.live__summary::before {
  content: '';
  width: 10px;
  height: 10px;
  border-right: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
  transform: rotate(-45deg);
  transition: transform 100ms ease-out;
}

.live__details[open] .live__summary::before {
  transform: rotate(45deg);
}

.live__pickers {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-3);
  padding-bottom: var(--space-2);
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
  max-width: none; /* bm.css caps li at 65ch for prose */
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 48px 40px 36px;
  gap: var(--space-2);
  align-items: center;
}

.live__gain {
  color: var(--text-muted);
}

.live__penalty {
  max-width: none;
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

/* race control: wide table, newest first, scrolls inside the card */
.live__rc-scroll {
  margin-top: var(--space-3);
  max-height: 320px;
  overflow: auto;
  background: var(--surface-sunken);
  border: var(--border-thin) solid var(--border);
}

.live__rc-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.live__rc-table th,
.live__rc-table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  white-space: nowrap;
  vertical-align: middle;
  border-bottom: 1px solid var(--border);
}

.live__rc-table thead th {
  position: sticky;
  top: 0;
  background: var(--surface-sunken);
  color: var(--text-muted);
  border-bottom: var(--border-thin) solid var(--border);
}

.live__rc-msg {
  white-space: normal;
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
  max-width: none;
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
