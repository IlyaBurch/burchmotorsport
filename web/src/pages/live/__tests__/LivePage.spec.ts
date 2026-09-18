import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import LivePage from '../ui/LivePage.vue'

const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

async function mountWith(handler: (url: string) => Response) {
  vi.stubGlobal('fetch', vi.fn(async (url: string) => handler(url)))
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/live', component: LivePage }] })
  await router.push('/live?session=1')
  const w = mount(LivePage, { global: { plugins: [router] } })
  await flushPromises()
  return w
}

describe('LivePage states', () => {
  it('shows a countdown for an upcoming session', async () => {
    const w = await mountWith((url) =>
      url.startsWith('/api/live')
        ? json({
            session: { session_key: 1, meeting_key: 1, session_name: 'Race', circuit_short_name: 'Baku', date_start: '2099-01-01T11:00:00+00:00', date_end: '2099-01-01T13:00:00+00:00' },
            upcoming: true,
            drivers: [],
            teams: [],
            radio: [],
            weather: null,
            raceControl: [],
            updatedAt: '',
          })
        : json([]),
    )
    expect(w.text()).toContain('До старта')
    expect(w.find('table').exists()).toBe(false)
  })

  it('apologises when the api has no data', async () => {
    const w = await mountWith((url) => (url.startsWith('/api/live') ? json('no data', 404) : json([])))
    expect(w.text()).toContain('Данные потерялись')
  })
})
