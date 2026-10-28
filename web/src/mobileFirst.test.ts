import { describe, expect, it } from 'vitest'
import indexHtml from '../index.html?raw'

// jsdom does no layout, so guard the mobile-first conventions at the source
// level: every frontend file is scanned for patterns that break on phones.
const sources = import.meta.glob<string>('./{components,views}/*.vue', { query: '?raw', import: 'default', eager: true })
const vueFiles = Object.entries(sources).map(([name, src]) => ({ name, src }))

describe('mobile-first conventions', () => {
  it('declares the viewport meta tag', () => {
    expect(indexHtml).toMatch(/name="viewport"[^>]*width=device-width/)
  })

  it.each(vueFiles)('$name: no desktop-first max-* breakpoints that fight positioning', ({ src }) => {
    expect(src).not.toMatch(/max-sm:(fixed|absolute|inset|top|bottom)/)
  })

  it.each(vueFiles)('$name: text inputs are 16px on phones (iOS zooms smaller ones)', ({ src }) => {
    for (const m of src.matchAll(/<input[^>]*>/g)) {
      const tag = m[0]
      if (/type="(checkbox|radio)"/.test(tag)) continue
      expect(tag, tag).not.toMatch(/(?<![:\w-])text-sm/)
    }
  })

  it.each(vueFiles)('$name: no fixed pixel widths beyond the viewport', ({ src }) => {
    expect(src).not.toMatch(/(?<![:\w-])w-\[\d{3,}px\]/)
  })

  it.each(vueFiles)('$name: full-height screens use dvh, not vh/screen (mobile browser chrome)', ({ src }) => {
    expect(src).not.toContain('min-h-screen')
  })
})
