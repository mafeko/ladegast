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

describe('SessionTable settlements', () => {
  const session = { created: '2026-10-17T10:00:00Z', finished: '2026-10-17T12:00:00Z', chargedKWh: 4, solarKWh: 2, gridKWh: 2, costEUR: 1.5 }
  const settlement = { id: 'a', at: '2026-10-18T10:00:00Z', chargedKWh: 4, solarKWh: 0, gridKWh: 0, amountEUR: 1.5 }

  it('lists a settlement with its sum, newest first, outside the total', () => {
    const wrapper = mount(SessionTable, { props: { sessions: [session], settlements: [settlement] } })
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0].attributes('data-testid')).toBe('settlement-row')
    expect(rows[0].text()).toContain('Abrechnung')
    expect(rows[0].text()).toContain('1,50')
    expect(wrapper.find('[data-testid="session-total"]').text()).toContain('4,0')
  })

  it('shows settlements even without sessions', () => {
    const wrapper = mount(SessionTable, { props: { sessions: [], settlements: [settlement] } })
    expect(wrapper.text()).not.toContain('Keine Ladevorgänge')
    expect(wrapper.findAll('[data-testid="settlement-row"]')).toHaveLength(1)
  })
})
