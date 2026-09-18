import { describe, expect, it } from 'vitest'
import { compoundLetter, formatGap, formatLap, formatSector } from './live'

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
})
