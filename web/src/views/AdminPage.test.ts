import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminPage from './AdminPage.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

describe('AdminPage', () => {
  it('shows the login form when listFriends is unauthorized', async () => {
    vi.mocked(api.listFriends).mockRejectedValue(new Error('not authenticated'))
    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.find('#username').exists()).toBe(true)
  })

  it('shows the friend management UI once authenticated', async () => {
    vi.mocked(api.listFriends).mockResolvedValue([
      { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '' },
    ])
    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Freunde verwalten')
    expect(wrapper.text()).toContain('Alex')
  })

  it('reloads the friend list after creating a friend', async () => {
    vi.mocked(api.listFriends).mockResolvedValueOnce([]).mockResolvedValueOnce([
      { id: '1', name: 'Bea', vehicleName: 'vehicle_2', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '' },
    ])
    vi.mocked(api.createFriend).mockResolvedValue({ id: '1', pin: '123456', link: 'http://x/l/1' })

    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    await wrapper.find('#name').setValue('Bea')
    await wrapper.find('#vehicleName').setValue('vehicle_2')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Bea')
    expect(wrapper.text()).toContain('http://x/l/1')
  })

  it('reflects a row update without re-fetching the whole list', async () => {
    vi.mocked(api.listFriends).mockResolvedValue([
      { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '' },
    ])
    vi.mocked(api.updateFriend).mockResolvedValue({
      id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.5, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '',
    })

    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    const saveButton = wrapper.findAll('button').find((b) => b.text() === 'Speichern')
    await saveButton!.trigger('click')
    await flushMicrotasks()

    expect(api.updateFriend).toHaveBeenCalled()
    expect(wrapper.text()).toContain('Gespeichert.')
  })
})
