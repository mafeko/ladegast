import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SummaryCard from './SummaryCard.vue'

describe('SummaryCard', () => {
  it('renders the title and formatted values', () => {
    const wrapper = mount(SummaryCard, {
      props: {
        title: 'Heute',
        summary: { chargedKWh: 10, solarKWh: 5, gridKWh: 5, costEUR: 2.25 },
      },
    })
    expect(wrapper.text()).toContain('Heute')
    expect(wrapper.text()).toContain('10.00 kWh')
    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('2,25')
  })
})
