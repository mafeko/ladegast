import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import UsageExplorer from './UsageExplorer.vue'
import { clearSessionCache } from '../lib/sessionCache'
import { rangeFor, startOfMonth, addMonths } from '../lib/range'
import type { SessionView } from '../lib/types'

const flush = () => new Promise((resolve) => setTimeout(resolve, 0))

function sessionIn(monthOffset: number, day: number, solar = 4, grid = 1): SessionView {
  const t = addMonths(startOfMonth(new Date()), monthOffset)
  t.setDate(day)
  t.setHours(12)
  return { created: t.toISOString(), finished: t.toISOString(), chargedKWh: solar + grid, solarKWh: solar, gridKWh: grid, costEUR: 1 }
}

beforeEach(() => clearSessionCache())

function mountExplorer(load: () => Promise<SessionView[]>, key = 'k') {
  return mount(UsageExplorer, { props: { cacheKey: key, load } })
}

describe('UsageExplorer', () => {
  it('shows only sessions of the selected range (default: this month)', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2), sessionIn(0, 3), sessionIn(-2, 5)])
    const wrapper = mountExplorer(load)
    await flush()

    expect(wrapper.text()).toContain('Ladevorgänge im Zeitraum (2)')
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)
  })

  it('filters the list and chart when the range changes without calling the API again', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2), sessionIn(-1, 5), sessionIn(-1, 6), sessionIn(-1, 7)])
    const wrapper = mountExplorer(load)
    await flush()
    expect(wrapper.text()).toContain('(1)')

    await wrapper.find('button[aria-label="Voriger Zeitraum"]').trigger('click')

    expect(wrapper.text()).toContain('(3)')
    expect(wrapper.text()).toContain('15,00 kWh gesamt')
    expect(load).toHaveBeenCalledTimes(1)
  })

  it('reuses the cache across mounts and reloads on "Aktualisieren"', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2)])
    mountExplorer(load)
    await flush()
    mountExplorer(load)
    await flush()
    expect(load).toHaveBeenCalledTimes(1)

    const wrapper = mountExplorer(load)
    await flush()
    await wrapper.findAll('button').find((b) => b.text() === 'Aktualisieren')!.trigger('click')
    await flush()
    expect(load).toHaveBeenCalledTimes(2)
  })

  it('shows an error when loading fails', async () => {
    const wrapper = mountExplorer(vi.fn().mockRejectedValue(new Error('EVCC ist nicht erreichbar.')))
    await flush()
    expect(wrapper.text()).toContain('EVCC ist nicht erreichbar.')
  })

  it('zooms into a bar: month -> that day by hour', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2, 6, 2)])
    const wrapper = mountExplorer(load)
    await flush()

    const bars = wrapper.findAll('rect[role="button"]')
    expect(bars.length).toBeGreaterThanOrEqual(28)
    await bars[1].trigger('click') // day 2
    await wrapper.findAll('button').find((b) => b.text() === 'Hineinzoomen')!.trigger('click')

    expect(wrapper.findAll('rect[role="button"]')).toHaveLength(24)
    expect(wrapper.text()).toContain('8,00 kWh gesamt')
  })
})

describe('range helper used by explorer', () => {
  it('rangeFor month ends at the next month start', () => {
    const r = rangeFor('month', new Date())
    expect(r.to.getDate()).toBe(1)
  })
})

describe('UsageExplorer without settlements', () => {
  it('renders chart and table when the friend has no settlement yet', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2)])
    const loadSettlements = vi.fn().mockResolvedValue({ settlements: [], open: { since: null, chargedKWh: 5, amountEUR: 1 } })
    const wrapper = mount(UsageExplorer, { props: { cacheKey: 'ns', load, loadSettlements } })
    await flush()
    expect(wrapper.find('svg').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Laden fehlgeschlagen')
  })

  it('tolerates a null settlements list', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2)])
    const loadSettlements = vi.fn().mockResolvedValue({ settlements: null, open: { since: null, chargedKWh: 0, amountEUR: 0 } })
    const wrapper = mount(UsageExplorer, { props: { cacheKey: 'ns2', load, loadSettlements } })
    await flush()
    expect(wrapper.find('svg').exists()).toBe(true)
  })
})

describe('UsageExplorer settlements failure', () => {
  it('still renders the chart when loading settlements fails', async () => {
    const load = vi.fn().mockResolvedValue([sessionIn(0, 2)])
    const loadSettlements = vi.fn().mockRejectedValue(new Error('boom'))
    const wrapper = mount(UsageExplorer, { props: { cacheKey: 'ns3', load, loadSettlements } })
    await flush()
    expect(wrapper.find('svg').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('boom')
  })
})
