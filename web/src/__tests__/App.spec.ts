import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from '@/app/App.vue'
import router from '@/app/router'

describe('App', () => {
  it('mounts properly', async () => {
    await router.push('/')
    const wrapper = mount(App, {
      global: {
        plugins: [router],
        stubs: { RouterView: true, BmHeader: true },
      },
    })
    expect(wrapper.exists()).toBe(true)
  })
})
