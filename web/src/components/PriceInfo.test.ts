import { mount } from '@vue/test-utils'
import { expect, it } from 'vitest'
import PriceInfo from './PriceInfo.vue'

it('shows both rates with two decimals', () => {
  const w = mount(PriceInfo, { props: { friend: { id: '1', name: 'A', rateSolarPerKWh: 0.1, rateGridPerKWh: 0.3 } } })
  expect(w.text()).toContain('0,10')
  expect(w.text()).toContain('0,30')
})
