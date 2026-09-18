export interface LiveSession {
  session_key: number
  meeting_key: number
  session_name: string
  session_type: string
  circuit_short_name: string
  country_name: string
  date_start: string
  date_end: string
}

export interface LiveDriver {
  position: number
  number: number
  acronym: string
  name: string
  team: string
  teamColour: string
  gap: number | string | null
  interval: number | string | null
  lastLap: number | null
  bestLap: number | null
  lap: number
  sectors: [number | null, number | null, number | null]
  bestSectors: [number | null, number | null, number | null]
  speedTrap: number
  compound: string
  tyreAge: number
  pits: number
  stints: { compound: string; from: number; to: number }[] | null
  positions: number[] // after each lap, 0 = unknown
  champPos: number
  champPoints: number
  x: number
  y: number
}

export interface Weather {
  air_temperature: number
  track_temperature: number
  humidity: number
  rainfall: number
  wind_speed: number
}

export interface RaceControl {
  date: string
  category: string
  flag: string
  message: string
  lap_number: number
}

export interface Team {
  name: string
  colour: string
  champPos: number
  champPoints: number
}

export interface Radio {
  date: string
  driver_number: number
  recording_url: string
}

export interface Live {
  session: LiveSession
  upcoming?: boolean // on the calendar, not started: drivers is empty
  drivers: LiveDriver[]
  teams: Team[]
  radio: Radio[]
  weather: Weather | null
  raceControl: RaceControl[]
  updatedAt: string
}

export interface Meeting {
  meeting_key: number
  meeting_name: string
  country_name: string
  circuit_short_name: string
  date_start: string
}

export interface Session {
  session_key: number
  session_name: string
  date_start: string
  date_end: string
}

export interface Resolved {
  title: string
  session: (LiveSession & { country_name: string }) | null
  f1: boolean
}

/** which openf1 session a video is about, judging by its title */
export const resolveTitle = (title: string) => getJSON<Resolved>(`/api/resolve?title=${encodeURIComponent(title)}`)

/** previous year's race on the same circuit, for previews of upcoming sessions */
export async function findPreviousRace(circuit: string, year: number): Promise<Session | null> {
  const meeting = (await fetchMeetings(year - 1)).find((m) => m.circuit_short_name === circuit)
  if (!meeting) return null
  return (await fetchSessions(meeting.meeting_key)).find((s) => s.session_name === 'Race') ?? null
}

/** the api knows the session but has no timing rows for it */
export class NoDataError extends Error {}

/** openf1 budget spent; retry after `seconds` */
export class RateLimitError extends Error {
  constructor(public seconds: number) {
    super(`rate limited for ${seconds}s`)
  }
}

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (res.status === 404) throw new NoDataError(url)
  if (res.status === 503) throw new RateLimitError(Number(res.headers.get('Retry-After')) || 60)
  if (!res.ok) throw new Error(`${url}: ${res.status}`)
  return res.json()
}

export const FIRST_SEASON = 2023 // openf1 has no data before

export const fetchLive = (sessionKey: string | number = 'latest') =>
  getJSON<Live>(`/api/live?session_key=${sessionKey}`)
export const fetchMeetings = (year: number) => getJSON<Meeting[]>(`/api/meetings?year=${year}`)
export const fetchTrack = (sessionKey: string | number) =>
  getJSON<[number, number][]>(`/api/track?session_key=${sessionKey}`)
export const fetchSessions = (meetingKey: number) =>
  getJSON<Session[]>(`/api/sessions?meeting_key=${meetingKey}`)

/** 30.123 → "30.123" */
export const formatSector = (s: number | null) => (s == null ? '—' : s.toFixed(3))

/** openf1 session names → ru */
const SESSION_NAMES: Record<string, string> = {
  'Practice 1': 'Практика 1',
  'Practice 2': 'Практика 2',
  'Practice 3': 'Практика 3',
  Qualifying: 'Квалификация',
  'Sprint Qualifying': 'Спринт-квалификация',
  'Sprint Shootout': 'Спринт-квалификация',
  Sprint: 'Спринт',
  Race: 'Гонка',
  'Day 1': 'День 1',
  'Day 2': 'День 2',
  'Day 3': 'День 3',
}
export const sessionLabel = (name: string) => SESSION_NAMES[name] ?? name

/** "F47600" → readable text colour on top of it */
export const inkOn = (hex: string) => {
  const n = parseInt(hex, 16)
  const lum = (0.2126 * (n >> 16) + 0.7152 * ((n >> 8) & 255) + 0.0722 * (n & 255)) / 255
  return lum > 0.55 ? 'var(--ink)' : 'var(--paper)'
}

const RACE_POINTS = [25, 18, 15, 12, 10, 8, 6, 4, 2, 1]
const SPRINT_POINTS = [8, 7, 6, 5, 4, 3, 2, 1]

/** points a driver takes home if the session ends right now */
export const pointsFor = (sessionName: string, position: number) => {
  const table = sessionName === 'Race' ? RACE_POINTS : sessionName === 'Sprint' ? SPRINT_POINTS : []
  return position > 0 ? (table[position - 1] ?? 0) : 0
}

export interface Projection {
  number: number
  gain: number
  points: number
  pos: number
  delta: number // vs position before the session, + = climbed
}

/** projected standings if nothing changes */
export function projectStandings(sessionName: string, drivers: LiveDriver[]): Map<number, Projection> {
  const rows = drivers.map((d) => {
    const gain = pointsFor(sessionName, d.position)
    return { number: d.number, gain, points: d.champPoints + gain, pos: 0, delta: 0, before: d.champPos }
  })
  rows.sort((a, b) => b.points - a.points)
  rows.forEach((r, i) => {
    r.pos = i + 1
    r.delta = r.before ? r.before - r.pos : 0
  })
  return new Map(rows.map((r) => [r.number, r]))
}

/** same idea for constructors: points before + both cars' projected gain */
export type TeamProjection = Projection & { name: string; colour: string }

export function projectTeams(sessionName: string, drivers: LiveDriver[], teams: Team[]): TeamProjection[] {
  const gain = new Map<string, number>()
  for (const d of drivers) gain.set(d.team, (gain.get(d.team) ?? 0) + pointsFor(sessionName, d.position))
  const rows = teams.map((t, i) => ({
    number: i,
    name: t.name,
    colour: t.colour,
    gain: gain.get(t.name) ?? 0,
    points: t.champPoints + (gain.get(t.name) ?? 0),
    pos: 0,
    delta: 0,
    before: t.champPos,
  }))
  rows.sort((a, b) => b.points - a.points)
  rows.forEach((r, i) => {
    r.pos = i + 1
    r.delta = r.before ? r.before - r.pos : 0
  })
  return rows
}

/** "... CAR 10 (GAS) ..." → [10] */
export const carsIn = (message: string) => [...message.matchAll(/CAR (\d+)/g)].map((m) => Number(m[1]))

export interface Penalty {
  key: string
  car: number
  what: string // "5 SECOND TIME PENALTY"
  why: string // "SPEEDING IN THE PIT LANE"
  lap: number
}

/** stewards' penalties, one per message */
export function penaltiesFrom(rc: RaceControl[]): Penalty[] {
  const out: Penalty[] = []
  for (const m of rc) {
    const hit = /FIA STEWARDS: (.+?) FOR CAR (\d+)(?: \(\w+\))?(?: - (.+))?$/.exec(m.message)
    if (!hit || /SERVED|INVESTIGATION|REVIEWED/.test(m.message)) continue
    out.push({ key: m.date + m.message, car: Number(hit[2]), what: hit[1]!, why: hit[3] ?? '', lap: m.lap_number })
  }
  return out
}

/** deleted laps for track limits, per car */
export function trackLimitsFrom(rc: RaceControl[]): Map<number, number> {
  const out = new Map<number, number>()
  for (const m of rc) {
    if (!/DELETED - TRACK LIMITS/.test(m.message)) continue
    for (const car of carsIn(m.message)) out.set(car, (out.get(car) ?? 0) + 1)
  }
  return out
}

/** race control flag → chip tone class suffix */
export const flagClass = (flag: string) =>
  /DOUBLE YELLOW|YELLOW/.test(flag) ? 'yellow'
  : /GREEN|CLEAR/.test(flag) ? 'green'
  : flag === 'RED' ? 'red'
  : flag === 'BLUE' ? 'blue'
  : flag === 'CHEQUERED' ? 'chequered'
  : flag ? 'plain' : ''

export const radioSrc = (url: string) => `/api/radio?url=${encodeURIComponent(url)}`

/** ms until → "5 дн 21ч 27м" */
export const countdown = (ms: number) => {
  if (ms <= 0) return 'Сейчас'
  const h = Math.floor(ms / 3_600_000)
  const d = Math.floor(h / 24)
  const m = Math.floor((ms % 3_600_000) / 60_000)
  return d > 0 ? `${d} дн ${h % 24}ч ${m}м` : `${h}ч ${m}м`
}

export const formatMsk = (iso: string) =>
  new Date(iso).toLocaleString('ru-RU', {
    timeZone: 'Europe/Moscow',
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }) + ' МСК'

export const formatTime = (iso: string) =>
  new Date(iso).toLocaleTimeString('ru-RU', { timeZone: 'Europe/Moscow', hour: '2-digit', minute: '2-digit' })

/** compound → css class suffix; colours are the F1 convention, mapped to tokens in css */
export const compoundClass = (c: string) =>
  ({ SOFT: 'soft', MEDIUM: 'medium', HARD: 'hard', INTERMEDIATE: 'inter', WET: 'wet' })[c] ?? 'unknown'

/** SOFT → "S" */
export const compoundLetter = (c: string) => (c === 'INTERMEDIATE' ? 'I' : c.charAt(0))

/** date_start already passed */
export const hasStarted = (x: { date_start: string }) => Date.parse(x.date_start) <= Date.now()

/** session is over (with an hour of slack) → no need to poll */
export const isFinished = (s: LiveSession) =>
  Date.parse(s.date_end) + 3600_000 < Date.now()

/** 96.03 → "1:36.030" */
export function formatLap(s: number | null): string {
  if (s == null) return '—'
  const ms = Math.round(s * 1000)
  const m = Math.floor(ms / 60000)
  const rest = ((ms % 60000) / 1000).toFixed(3).padStart(6, '0')
  return `${m}:${rest}`
}

/** 4.351 → "+4.351", "+1 LAP" → "+1 LAP", leader → "—" */
export function formatGap(g: number | string | null): string {
  if (g == null) return '—'
  if (typeof g === 'string') return g
  return g === 0 ? '—' : `+${g.toFixed(3)}`
}
