import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GuestPage from './GuestPage.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

describe('GuestPage', () => {
  it('shows the login form when not authenticated', async () => {
    vi.mocked(api.getGuestStatus).mockResolvedValue({ authenticated: false })
    const wrapper = mount(GuestPage, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    expect(wrapper.find('#pin').exists()).toBe(true)
  })

  it('shows the dashboard when already authenticated', async () => {
    vi.mocked(api.getGuestStatus).mockResolvedValue({ authenticated: true })
    vi.mocked(api.getDashboard).mockResolvedValue({
      guest: { id: 'abc', name: 'Alex', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.3 },
      today: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisWeek: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisMonth: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisYear: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
    })
    vi.mocked(api.getSessions).mockResolvedValue([])

    const wrapper = mount(GuestPage, { props: { uuid: 'abc' } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('Hallo Alex')
  })

  it('switches from login to dashboard and back to login on logout', async () => {
    vi.mocked(api.getGuestStatus).mockResolvedValue({ authenticated: false })
    vi.mocked(api.guestLogin).mockResolvedValue({ name: 'Alex' })
    vi.mocked(api.getDashboard).mockResolvedValue({
      guest: { id: 'abc', name: 'Alex', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.3 },
      today: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisWeek: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisMonth: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
      thisYear: { chargedKWh: 0, solarKWh: 0, gridKWh: 0, costEUR: 0 },
    })
    vi.mocked(api.getSessions).mockResolvedValue([])
    vi.mocked(api.guestLogout).mockResolvedValue(undefined)

    const wrapper = mount(GuestPage, { props: { uuid: 'abc' } })
    await flushMicrotasks()
    expect(wrapper.find('#pin').exists()).toBe(true)

    await wrapper.find('#pin').setValue('123456')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()
    expect(wrapper.text()).toContain('Hallo Alex')

    await wrapper.find('button').trigger('click') // "Abmelden" inside GuestDashboard
    await flushMicrotasks()
    expect(wrapper.find('#pin').exists()).toBe(true)
  })

  it('shows an invalid-link message for an unknown uuid', async () => {
    vi.mocked(api.getGuestStatus).mockRejectedValue(new Error('unknown link'))
    const wrapper = mount(GuestPage, { props: { uuid: 'ghost' } })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('ungültig')
  })
})
