import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import AppFooter from './AppFooter.vue'

describe('AppFooter', () => {
  it('links to the GitHub repository', () => {
    const a = mount(AppFooter).find('a')
    expect(a.attributes('href')).toBe('https://github.com/mafeko/ladegast')
    expect(a.attributes('rel')).toContain('noopener')
  })
})
