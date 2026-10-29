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
  it('sends an optional chosen PIN', async () => {
    vi.mocked(api.createFriend).mockResolvedValue({ id: '1', pin: '2468', link: 'http://x/l/1' })
    const wrapper = mount(AdminFriendForm)

    await wrapper.find('#name').setValue('Bea')
    await wrapper.find('#vehicleName').setValue('vehicle_2')
    await wrapper.find('#pin').setValue('2468')
    await wrapper.find('form').trigger('submit.prevent')
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(api.createFriend).toHaveBeenCalledWith(expect.objectContaining({ pin: '2468' }))
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

describe('AdminFriendForm vehicle suggestions', () => {
  it('offers unassigned EVCC vehicles as suggestions', async () => {
    vi.mocked(api.listVehicles).mockResolvedValue(['vehicle_1', 'vehicle_2'])
    const wrapper = mount(AdminFriendForm)
    await new Promise((resolve) => setTimeout(resolve, 0))

    const options = wrapper.findAll('#vehicleSuggestions option').map((o) => o.attributes('value'))
    expect(options).toEqual(['vehicle_1', 'vehicle_2'])
  })

  it('still works when loading suggestions fails', async () => {
    vi.mocked(api.listVehicles).mockRejectedValue(new Error('boom'))
    const wrapper = mount(AdminFriendForm)
    await new Promise((resolve) => setTimeout(resolve, 0))

    expect(wrapper.findAll('#vehicleSuggestions option')).toHaveLength(0)
  })
})
