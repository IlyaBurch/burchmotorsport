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
  gap: number | string | null
  interval: number | string | null
  lastLap: number | null
  bestLap: number | null
  lap: number
}

export interface Live {
  session: LiveSession
  drivers: LiveDriver[]
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
export const fetchSessions = (meetingKey: number) =>
  getJSON<Session[]>(`/api/sessions?meeting_key=${meetingKey}`)

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
