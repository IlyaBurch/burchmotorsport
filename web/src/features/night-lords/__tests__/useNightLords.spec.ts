import { describe, expect, it, vi } from 'vitest'
import { useNightLords } from '../model/useNightLords'
import { useTheme } from '@/features/theme-toggle'

describe('useNightLords', () => {
  it('switches on after 8 quick taps and back after 8 more', () => {
    vi.useFakeTimers()
    const egg = useNightLords()
    const { theme, set } = useTheme()
    set('dark')
    for (let i = 0; i < 7; i++) egg.tap()
    expect(theme.value).toBe('dark')
    egg.tap()
    expect(theme.value).toBe('night-lords')
    expect(egg.toast.value).toBe(true)
    vi.advanceTimersByTime(3000)
    expect(egg.toast.value).toBe(false)
    for (let i = 0; i < 8; i++) egg.tap()
    expect(theme.value).toBe('dark')
    vi.useRealTimers()
  })

  it('ignores slow taps', () => {
    vi.useFakeTimers()
    const egg = useNightLords()
    const { theme, set } = useTheme()
    set('dark')
    for (let i = 0; i < 8; i++) {
      egg.tap()
      vi.advanceTimersByTime(1000)
    }
    expect(theme.value).toBe('dark')
    vi.useRealTimers()
  })
})
