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
