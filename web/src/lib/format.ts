// Small presentation helpers shared by the dashboard views.

const kWhFormat = new Intl.NumberFormat('de-DE', { minimumFractionDigits: 1, maximumFractionDigits: 1 })

/** Energy as a bare number, one decimal place. */
export function formatKWhValue(value: number): string {
  return kWhFormat.format(value)
}

export function formatKWh(value: number): string {
  return `${formatKWhValue(value)} kWh`
}

export function formatEUR(value: number): string {
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(value)
}

export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return new Intl.DateTimeFormat('de-DE', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(d)
}

export function solarSharePercent(solarKWh: number, chargedKWh: number): number {
  if (chargedKWh <= 0) return 0
  return Math.round((solarKWh / chargedKWh) * 100)
}
