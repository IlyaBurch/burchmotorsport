import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import WatchPage from '../ui/WatchPage.vue'

const id = '0123456789abcdef0123456789abcdef'

async function mountAt(path: string, handler: (url: string) => Response = () => new Response('no data', { status: 404 })) {
  vi.stubGlobal('fetch', vi.fn(async (url: string) => handler(url)))
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/watch', component: WatchPage }] })
  await router.push(path)
  const w = mount(WatchPage, { global: { plugins: [router], stubs: { Teleport: true } } })
  await flushPromises()
  return { w, router }
}

/** what the RuTube iframe posts once getPlayOptions=title is loaded */
async function playerTitle(title: string) {
  window.dispatchEvent(
    new MessageEvent('message', { origin: 'https://rutube.ru', data: JSON.stringify({ type: 'player:playOptionLoaded', data: { title } }) }),
  )
  await flushPromises()
}

describe('WatchPage', () => {
  it('asks for a link and rejects non-rutube ones', async () => {
    const { w, router } = await mountAt('/watch')
    expect(w.find('iframe').exists()).toBe(false)
    await w.find('input').setValue('https://vk.com/video')
    await w.find('form').trigger('submit')
    expect(w.text()).toContain('Это не ссылка на RuTube')
    await router.replace({ query: { session: '9947' } })
    await w.find('input').setValue(`https://rutube.ru/video/${id}/`)
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(router.currentRoute.value.query.v).toBe(id)
    expect(router.currentRoute.value.query.session).toBeUndefined()
  })

  it('embeds the player when the url carries a video id', async () => {
    const { w } = await mountAt(`/watch?v=${id}&session=1`)
    expect(w.find('iframe').attributes('src')).toContain(`/play/embed/${id}`)
  })

  it('resolves the session from the video title when the url has none', async () => {
    const { w, router } = await mountAt(`/watch?v=${id}`, (url) =>
      url.startsWith('/api/resolve')
        ? new Response(JSON.stringify({ title: 't', session: { session_key: 9947 }, f1: true }), { status: 200 })
        : new Response('no data', { status: 404 }),
    )
    await playerTitle('Формула 1 - Гран-При Великобритании 2025 - Гонка')
    expect(router.currentRoute.value.query.session).toBe('9947')
    expect(w.text()).toContain('По видео')
  })

  it('sends non-F1 videos back to the form', async () => {
    const { w, router } = await mountAt(`/watch?v=${id}`, (url) =>
      url.startsWith('/api/resolve')
        ? new Response(JSON.stringify({ title: 'Котики', session: null, f1: false }), { status: 200 })
        : new Response('no data', { status: 404 }),
    )
    await playerTitle('Котики играют')
    expect(router.currentRoute.value.query.v).toBeUndefined()
    expect(w.find('iframe').exists()).toBe(false)
    expect(w.text()).toContain('не Формула 1')
  })
})
