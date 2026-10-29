import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import FriendPage from './FriendPage.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

describe('FriendPage', () => {
  it('shows the login form when not authenticated', async () => {
    vi.mocked(api.getFriendStatus).mockResolvedValue({ authenticated: false })
    const wrapper = mount(FriendPage, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    expect(wrapper.find('#pin').exists()).toBe(true)
  })

  it('shows the dashboard when already authenticated', async () => {
    vi.mocked(api.getFriendStatus).mockResolvedValue({ authenticated: true })
    vi.mocked(api.getDashboard).mockResolvedValue({
      friend: { id: 'abc', name: 'Alex', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.3 },
      today: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisWeek: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisMonth: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisYear: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
    })
    vi.mocked(api.getSessions).mockResolvedValue([])

    const wrapper = mount(FriendPage, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Hallo Alex')
  })

  it('switches from login to dashboard and back to login on logout', async () => {
    vi.mocked(api.getFriendStatus).mockResolvedValue({ authenticated: false })
    vi.mocked(api.friendLogin).mockResolvedValue({ name: 'Alex' })
    vi.mocked(api.getDashboard).mockResolvedValue({
      friend: { id: 'abc', name: 'Alex', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.3 },
      today: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisWeek: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisMonth: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisYear: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
    })
    vi.mocked(api.getSessions).mockResolvedValue([])
    vi.mocked(api.friendLogout).mockResolvedValue(undefined)

    const wrapper = mount(FriendPage, { props: { uuid: 'abc' } })
    await flushMicrotasks()
    expect(wrapper.find('#pin').exists()).toBe(true)

    await wrapper.find('#pin').setValue('123456')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()
    expect(wrapper.text()).toContain('Hallo Alex')

    await wrapper.find('button').trigger('click') // "Abmelden" inside FriendDashboard
    await flushMicrotasks()
    expect(wrapper.find('#pin').exists()).toBe(true)
  })

  it('shows an invalid-link message for an unknown uuid', async () => {
    vi.mocked(api.getFriendStatus).mockRejectedValue(new Error('unknown link'))
    const wrapper = mount(FriendPage, { props: { uuid: 'ghost' } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('ungültig')
  })
})
