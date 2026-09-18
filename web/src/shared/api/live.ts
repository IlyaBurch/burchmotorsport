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

export interface Live {
  session: LiveSession
  drivers: LiveDriver[]
  weather: Weather | null
  raceControl: RaceControl[]
  updatedAt: string
}

export interface Meeting {
  meeting_key: number
  meeting_name: string
  country_name: string
  date_start: string
}

export interface Session {
  session_key: number
  session_name: string
  date_start: string
}

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
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
