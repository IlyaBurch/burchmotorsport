export interface LiveSession {
  session_key: number
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

export async function fetchLive(): Promise<Live> {
  const res = await fetch('/api/live')
  if (!res.ok) throw new Error(`live: ${res.status}`)
  return res.json()
}

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
