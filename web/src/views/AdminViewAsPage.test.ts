import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminViewAsPage from './AdminViewAsPage.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

describe('AdminViewAsPage', () => {
  it('loads and renders the dashboard for the given friend id', async () => {
    vi.mocked(api.viewAsDashboard).mockResolvedValue({
      friend: { id: '1', name: 'Alex' },
      today: { chargedKWh: 1, solarKWh: 1, gridKWh: 0, costEUR: 0.1 },
      thisWeek: { chargedKWh: 1, solarKWh: 1, gridKWh: 0, costEUR: 0.1 },
      thisMonth: { chargedKWh: 1, solarKWh: 1, gridKWh: 0, costEUR: 0.1 },
      thisYear: { chargedKWh: 1, solarKWh: 1, gridKWh: 0, costEUR: 0.1 },
    })
    vi.mocked(api.viewAsSessions).mockResolvedValue([])

    const wrapper = mount(AdminViewAsPage, {
      props: { id: '1' },
      global: { stubs: { RouterLink: true } },
    })
    await flushMicrotasks()

    expect(api.viewAsDashboard).toHaveBeenCalledWith('1')
    expect(wrapper.text()).toContain('Ansicht als Alex')
  })

  it('shows an error message when loading fails', async () => {
    vi.mocked(api.viewAsDashboard).mockRejectedValue(new Error('friend not found'))
    vi.mocked(api.viewAsSessions).mockResolvedValue([])

    const wrapper = mount(AdminViewAsPage, {
      props: { id: 'missing' },
      global: { stubs: { RouterLink: true } },
    })
    await flushMicrotasks()

    expect(wrapper.text()).toContain('friend not found')
  })
})
