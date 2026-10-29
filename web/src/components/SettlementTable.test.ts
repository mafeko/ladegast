import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SettlementTable from './SettlementTable.vue'

describe('SettlementTable totals', () => {
  const data = {
    settlements: [{ id: 'a', at: '2026-10-06T20:00:00Z', chargedKWh: 178, solarKWh: 0, gridKWh: 0, amountEUR: 49.35 }],
    open: { since: '2026-10-06T20:00:00Z', chargedKWh: 99.8, solarKWh: 0, gridKWh: 0, amountEUR: 30.78 },
  }

  it('shows settled, open and grand total', () => {
    const w = mount(SettlementTable, { props: { data } })
    expect(w.find('[data-testid="settled-total"]').text()).toContain('49,35')
    expect(w.find('[data-testid="open-total"]').text()).toContain('30,78')
    expect(w.find('[data-testid="grand-total"]').text()).toContain('80,13')
    expect(w.find('[data-testid="grand-total"]').text()).toContain('277,8')
  })
})

describe('SettlementTable solar/grid split', () => {
  it('lists solar and grid kWh in brackets with icons', () => {
    const data = {
      settlements: [{ id: 'a', at: '2026-10-06T20:00:00Z', chargedKWh: 10, solarKWh: 6, gridKWh: 4, amountEUR: 2 }],
      open: { since: null, chargedKWh: 5, solarKWh: 1, gridKWh: 4, amountEUR: 1 },
    }
    const w = mount(SettlementTable, { props: { data } })
    const split = w.find('tbody tr:nth-child(2) [data-testid="split"]')
    expect(split.text()).toContain('6,0')
    expect(split.text()).toContain('4,0')
    expect(split.find('[data-icon="sun"]').exists()).toBe(true)
    expect(split.find('[data-icon="plug"]').exists()).toBe(true)
    expect(w.find('[data-testid="grand-total"] [data-testid="split"]').text()).toContain('7,0')
  })
})
