import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from '@/app/App.vue'

describe('App', () => {
  it('mounts properly', () => {
    const wrapper = mount(App, {
      global: {
        stubs: {
          RouterLink: true,
          RouterView: true,
          BmHeader: true,
        },
      },
    })
    expect(wrapper.exists()).toBe(true)
  })
})
