import { describe, expect, it } from 'vitest'
import { parseRutubeId } from './rutube'

const id = '0123456789abcdef0123456789abcdef'

describe('parseRutubeId', () => {
  it('accepts the usual url shapes and a bare id', () => {
    expect(parseRutubeId(`https://rutube.ru/video/${id}/`)).toBe(id)
    expect(parseRutubeId(`https://rutube.ru/video/${id}/?r=wd`)).toBe(id)
    expect(parseRutubeId(`https://rutube.ru/play/embed/${id}`)).toBe(id)
    expect(parseRutubeId(`https://rutube.ru/live/video/${id}/`)).toBe(id)
    expect(parseRutubeId(` ${id.toUpperCase()} `)).toBe(id)
  })
  it('rejects everything else', () => {
    expect(parseRutubeId('https://vk.com/video-1_2')).toBeNull()
    expect(parseRutubeId('https://rutube.ru/channel/123/')).toBeNull()
    expect(parseRutubeId('lol')).toBeNull()
  })
})
