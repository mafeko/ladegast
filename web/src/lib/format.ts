// Small presentation helpers shared by the dashboard views.

const kWhFormat = new Intl.NumberFormat('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

export function formatKWh(value: number): string {
  return `${kWhFormat.format(value)} kWh`
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
