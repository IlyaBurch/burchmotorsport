import { describe, expect, it } from 'vitest'
import { compoundLetter, formatGap, formatLap, formatSector, inkOn, pointsFor, projectStandings, sessionLabel, type LiveDriver } from './live'

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
})
