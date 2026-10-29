import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SessionTable from './SessionTable.vue'

describe('SessionTable', () => {
  it('renders one row per session', () => {
    const wrapper = mount(SessionTable, {
      props: {
        sessions: [
          { created: '2026-10-17T10:00:00Z', finished: '2026-10-17T12:00:00Z', chargedKWh: 5, solarKWh: 2.5, gridKWh: 2.5, costEUR: 1 },
        ],
      },
    })
    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
    expect(wrapper.text()).toContain('5,0')
  })

  it('shows an empty state when there are no sessions', () => {
    const wrapper = mount(SessionTable, { props: { sessions: [] } })
    expect(wrapper.text()).toContain('Keine Ladevorgänge')
  })
})

describe('SessionTable total', () => {
  const s = (kwh: number, cost: number) => ({
    created: '2026-10-17T10:00:00Z', finished: '2026-10-17T12:00:00Z', chargedKWh: kwh, solarKWh: kwh / 2, gridKWh: kwh / 2, costEUR: cost,
  })

  it('sums all visible sessions below the table', () => {
    const wrapper = mount(SessionTable, { props: { sessions: [s(4, 1.5), s(6, 2)] } })
    const total = wrapper.find('tfoot [data-testid="session-total"]')
    expect(total.text()).toContain('10,0')
    expect(total.text()).toContain('3,50')
  })

  it('shows no total row when empty', () => {
    const wrapper = mount(SessionTable, { props: { sessions: [] } })
    expect(wrapper.find('tfoot').exists()).toBe(false)
  })
})

describe('SessionTable header', () => {
  it('uses icons instead of long titles', () => {
    const wrapper = mount(SessionTable, { props: { sessions: [] } })
    const head = wrapper.find('thead')
    for (const icon of ['car', 'sun', 'plug', 'euro']) expect(head.find(`[data-icon="${icon}"]`).exists()).toBe(true)
    expect(head.text()).not.toMatch(/Geladen|Sonne|Netz|Kosten/)
  })
})
