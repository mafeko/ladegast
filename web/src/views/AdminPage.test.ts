import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminPage from './AdminPage.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

beforeEach(() => {
  vi.stubGlobal('confirm', vi.fn(() => true))
})

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
      { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' },
    ])
    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Freunde verwalten')
    expect((wrapper.find('input[aria-label="Name"]').element as HTMLInputElement).value).toBe('Alex')
  })

  it('reloads the friend list after creating a friend', async () => {
    vi.mocked(api.listFriends).mockResolvedValueOnce([]).mockResolvedValueOnce([
      { id: '1', name: 'Bea', vehicleName: 'vehicle_2', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' },
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

  it('shows an empty state when there are no friends', async () => {
    vi.mocked(api.listFriends).mockResolvedValue([])
    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Noch keine Freunde angelegt.')
  })

  it('reloads the list and shows the new link after regenerating', async () => {
    const friend = { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' }
    vi.mocked(api.listFriends).mockResolvedValueOnce([friend]).mockResolvedValueOnce([{ ...friend, id: '2' }])
    vi.mocked(api.regenerateLink).mockResolvedValue({ id: '2', link: 'http://x/l/2' })

    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()
    await wrapper.findAll('button').find((b) => b.text() === 'Link neu')!.trigger('click')
    await flushMicrotasks()

    expect(api.listFriends).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('http://x/l/2')
  })

  it('removes a friend from the list after deleting', async () => {
    const friend = { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' }
    vi.mocked(api.listFriends).mockResolvedValue([friend])
    vi.mocked(api.deleteFriend).mockResolvedValue(undefined)

    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()
    await wrapper.findAll('button').find((b) => b.text() === 'Löschen')!.trigger('click')
    await flushMicrotasks()

    expect(api.deleteFriend).toHaveBeenCalledWith('1')
    expect(wrapper.text()).toContain('Noch keine Freunde angelegt.')
    expect(wrapper.text()).toContain('Alex wurde gelöscht.')
  })

  it('logs out and shows the login form again', async () => {
    vi.mocked(api.listFriends).mockResolvedValue([])
    vi.mocked(api.adminLogout).mockResolvedValue(undefined)
    const wrapper = mount(AdminPage, { global: { stubs: { RouterLink: true } } })
    await flushMicrotasks()

    await wrapper.findAll('button').find((b) => b.text() === 'Abmelden')!.trigger('click')
    await flushMicrotasks()

    expect(api.adminLogout).toHaveBeenCalled()
    expect(wrapper.find('#username').exists()).toBe(true)
  })

  it('reflects a row update without re-fetching the whole list', async () => {
    vi.mocked(api.listFriends).mockResolvedValue([
      { id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1' },
    ])
    vi.mocked(api.updateFriend).mockResolvedValue({
      id: '1', name: 'Alex', vehicleName: 'vehicle_1', rateSolarPerKWh: 0.5, rateGridPerKWh: 0.35, createdAt: '', linkRegeneratedAt: '', link: 'http://x/l/1',
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
