import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminFriendForm from './AdminFriendForm.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

describe('AdminFriendForm', () => {
  it('creates a friend and emits the result', async () => {
    const created = { id: '1', pin: '123456', link: 'http://x/l/1' }
    vi.mocked(api.createFriend).mockResolvedValue(created)

    const wrapper = mount(AdminFriendForm)
    await wrapper.find('#name').setValue('Alex')
    await wrapper.find('#vehicleName').setValue('vehicle_1')
    await wrapper.find('#rateSolar').setValue('0.1')
    await wrapper.find('#rateGrid').setValue('0.35')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(api.createFriend).toHaveBeenCalledWith({
      name: 'Alex',
      vehicleName: 'vehicle_1',
      rateSolarPerKWh: 0.1,
      rateGridPerKWh: 0.35,
    })
    expect(wrapper.emitted('created')).toEqual([[created]])
    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('')
  })

  it('shows an error message on failure', async () => {
    vi.mocked(api.createFriend).mockRejectedValue(new Error('name and vehicleName are required'))
    const wrapper = mount(AdminFriendForm)

    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('name and vehicleName are required')
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
