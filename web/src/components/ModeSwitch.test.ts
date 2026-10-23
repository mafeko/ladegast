import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ModeSwitch from './ModeSwitch.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

describe('ModeSwitch', () => {
  it('emits changed with the new mode on success', async () => {
    vi.mocked(api.setMode).mockResolvedValue({ mode: 'now' })
    const wrapper = mount(ModeSwitch, { props: { uuid: 'abc' } })

    await wrapper.findAll('button')[1].trigger('click') // "Schnellladen"
    await flushMicrotasks()

    expect(api.setMode).toHaveBeenCalledWith('abc', 'now')
    expect(wrapper.emitted('changed')).toEqual([['now']])
  })

  it('marks the chosen mode as active and confirms it', async () => {
    vi.mocked(api.setMode).mockResolvedValue({ mode: 'now' })
    const wrapper = mount(ModeSwitch, { props: { uuid: 'abc' } })
    const buttons = () => wrapper.findAll('button')
    expect(buttons().map((b) => b.attributes('aria-pressed'))).toEqual(['false', 'false'])

    await buttons()[1].trigger('click')
    await flushMicrotasks()

    expect(buttons().map((b) => b.attributes('aria-pressed'))).toEqual(['false', 'true'])
    expect(wrapper.text()).toContain('Modus gesetzt: Schnellladen')
  })

  it('shows an error message when the switch fails', async () => {
    vi.mocked(api.setMode).mockRejectedValue(new Error('could not set mode'))
    const wrapper = mount(ModeSwitch, { props: { uuid: 'abc' } })

    await wrapper.findAll('button')[0].trigger('click') // "PV-Überschuss"
    await flushMicrotasks()

    expect(wrapper.text()).toContain('could not set mode')
    expect(wrapper.emitted('changed')).toBeUndefined()
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
