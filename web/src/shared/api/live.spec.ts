import { describe, expect, it } from 'vitest'
import { compoundLetter, formatGap, formatLap, formatSector, countdown, inkOn, penaltiesFrom, pointsFor, projectStandings, projectTeams, sessionLabel, trackLimitsFrom, type LiveDriver, type RaceControl } from './live'

describe('live formatters', () => {
  it('formats laps as m:ss.mmm', () => {
    expect(formatLap(96.03)).toBe('1:36.030')
    expect(formatLap(59.9995)).toBe('1:00.000')
    expect(formatLap(null)).toBe('—')
  })
  it('formats gaps', () => {
    expect(formatGap(0)).toBe('—')
    expect(formatGap(4.351)).toBe('+4.351')
    expect(formatGap('+1 LAP')).toBe('+1 LAP')
    expect(formatGap(null)).toBe('—')
  })
  it('formats sectors and compounds', () => {
    expect(formatSector(30.1)).toBe('30.100')
    expect(compoundLetter('INTERMEDIATE')).toBe('I')
    expect(compoundLetter('SOFT')).toBe('S')
  })
  it('localizes sessions and picks readable ink', () => {
    expect(sessionLabel('Race')).toBe('Гонка')
    expect(sessionLabel('Weird')).toBe('Weird')
    expect(inkOn('FFFFFF')).toBe('var(--ink)')
    expect(inkOn('0000FF')).toBe('var(--paper)')
  })
  it('projects standings', () => {
    expect(pointsFor('Race', 1)).toBe(25)
    expect(pointsFor('Sprint', 8)).toBe(1)
    expect(pointsFor('Qualifying', 1)).toBe(0)
    expect(pointsFor('Race', 15)).toBe(0)
    const d = (number: number, position: number, champPoints: number, champPos: number) =>
      ({ number, position, champPoints, champPos }) as LiveDriver
    const p = projectStandings('Race', [d(1, 2, 100, 1), d(2, 1, 90, 2)])
    expect(p.get(1)).toMatchObject({ gain: 18, points: 118, pos: 1, delta: 0 })
    expect(p.get(2)).toMatchObject({ gain: 25, points: 115, pos: 2, delta: 0 })
  })
  it('projects teams', () => {
    const d = (number: number, position: number, team: string) => ({ number, position, team }) as LiveDriver
    const t = projectTeams('Race', [d(1, 1, 'A'), d(2, 2, 'A'), d(3, 3, 'B')], [
      { name: 'A', colour: '', champPos: 2, champPoints: 0 },
      { name: 'B', colour: '', champPos: 1, champPoints: 20 },
    ])
    expect(t[0]).toMatchObject({ name: 'A', gain: 43, points: 43, pos: 1, delta: 1 })
    expect(t[1]).toMatchObject({ name: 'B', gain: 15, points: 35, pos: 2, delta: -1 })
  })
  it('parses penalties and track limits', () => {
    const rc = (message: string, lap = 1) => ({ message, lap_number: lap, date: 'd', category: 'Other', flag: '' }) as RaceControl
    const p = penaltiesFrom([
      rc('FIA STEWARDS: 5 SECOND TIME PENALTY FOR CAR 10 (GAS) - SPEEDING IN THE PIT LANE', 57),
      rc('FIA STEWARDS: PENALTY SERVED - 5 SECOND TIME PENALTY FOR CAR 5 (BOR) - UNSAFE RELEASE'),
      rc('FIA STEWARDS: TURN 5 INCIDENT INVOLVING CAR 3 (VER) REVIEWED NO FURTHER INVESTIGATION'),
    ])
    expect(p).toEqual([{ key: expect.any(String), car: 10, what: '5 SECOND TIME PENALTY', why: 'SPEEDING IN THE PIT LANE', lap: 57 }])
    const tl = trackLimitsFrom([
      rc('CAR 81 (PIA) LAP DELETED - TRACK LIMITS AT TURN 12 LAP 44'),
      rc('CAR 81 (PIA) TIME 1:29.379 DELETED - TRACK LIMITS AT TURN 6 LAP 32'),
      rc('CAR 4 (NOR) TIME 1:45.587 DELETED - TRACK LIMITS AT TURN 12 LAP 44'),
    ])
    expect([...tl]).toEqual([[81, 2], [4, 1]])
  })
  it('formats countdowns', () => {
    expect(countdown(0)).toBe('Сейчас')
    expect(countdown(90 * 60_000)).toBe('1ч 30м')
    expect(countdown((5 * 24 + 21) * 3_600_000 + 27 * 60_000)).toBe('5 дн 21ч 27м')
  })
})
