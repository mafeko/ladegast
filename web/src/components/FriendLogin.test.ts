import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import FriendLogin from './FriendLogin.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

describe('FriendLogin', () => {
  it('emits success with the friend name on correct PIN', async () => {
    vi.mocked(api.friendLogin).mockResolvedValue({ name: 'Alex' })
    const wrapper = mount(FriendLogin, { props: { uuid: 'abc' } })

    await wrapper.find('#pin').setValue('123456')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(api.friendLogin).toHaveBeenCalledWith('abc', '123456')
    expect(wrapper.emitted('success')).toEqual([['Alex']])
  })

  it('shows the error message on wrong PIN', async () => {
    vi.mocked(api.friendLogin).mockRejectedValue(new Error('invalid PIN'))
    const wrapper = mount(FriendLogin, { props: { uuid: 'abc' } })

    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('invalid PIN')
    expect(wrapper.emitted('success')).toBeUndefined()
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
