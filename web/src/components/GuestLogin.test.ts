import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GuestLogin from './GuestLogin.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

describe('GuestLogin', () => {
  it('emits success with the guest name on correct PIN', async () => {
    vi.mocked(api.guestLogin).mockResolvedValue({ name: 'Alex' })
    const wrapper = mount(GuestLogin, { props: { uuid: 'abc' } })

    await wrapper.find('#pin').setValue('123456')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(api.guestLogin).toHaveBeenCalledWith('abc', '123456')
    expect(wrapper.emitted('success')).toEqual([['Alex']])
  })

  it('shows the error message on wrong PIN', async () => {
    vi.mocked(api.guestLogin).mockRejectedValue(new Error('invalid PIN'))
    const wrapper = mount(GuestLogin, { props: { uuid: 'abc' } })

    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('invalid PIN')
    expect(wrapper.emitted('success')).toBeUndefined()
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
