// Mirrors the JSON shapes returned by internal/httpserver (Go backend).

export type Mode = 'smart' | 'now'

export interface Summary {
  chargedKWh: number
  solarKWh: number
  gridKWh: number
  costEUR: number
}

export interface FriendInfo {
  id: string
  name: string
  rateSolarPerKWh: number
  rateGridPerKWh: number
}

export interface Dashboard {
  friend: FriendInfo
  today: Summary
  thisWeek: Summary
  thisMonth: Summary
  thisYear: Summary
}

export interface SessionView {
  created: string
  finished: string
  chargedKWh: number
  solarKWh: number
  gridKWh: number
  costEUR: number
}

export interface FriendStatus {
  authenticated: boolean
}

export interface Friend {
  id: string
  name: string
  vehicleName: string
  rateSolarPerKWh: number
  rateGridPerKWh: number
  createdAt: string
  linkRegeneratedAt: string
  link: string
}

export interface CreateFriendInput {
  name: string
  vehicleName: string
  rateSolarPerKWh: number
  rateGridPerKWh: number
  /** Optional chosen PIN (4-12 digits); the server generates one if omitted. */
  pin?: string
}

export interface CreatedFriend {
  id: string
  pin: string
  link: string
}

export interface RegeneratedLink {
  id: string
  link: string
}

export interface Settlement {
  id: string
  at: string
  chargedKWh: number
  solarKWh: number
  gridKWh: number
  amountEUR: number
}

/** Charged since the last settlement (or ever) and still unpaid. */
export interface OpenAmount {
  since: string | null
  chargedKWh: number
  solarKWh: number
  gridKWh: number
  amountEUR: number
}

export interface Settlements {
  settlements: Settlement[]
  open: OpenAmount
}
