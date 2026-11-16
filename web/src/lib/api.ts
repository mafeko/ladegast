// Thin client for Ladegast's backend JSON API. One function per endpoint
// documented in DESIGN.md section 6 - no business logic here, just
// request/response shaping, so every function is directly unit-testable
// by mocking global fetch.
import type {
  CreateFriendInput,
  CreatedFriend,
  Dashboard,
  Friend,
  FriendStatus,
  Mode,
  RegeneratedLink,
  SessionView,
  Settlements,
} from './types'

export const PUBLIC_PREFIX = '/public'
export const ADMIN_PREFIX = '/admin'

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// Backend error texts are English; users see German.
const ERROR_MESSAGES: Record<string, string> = {
  'invalid PIN': 'Falsche PIN.',
  'invalid credentials': 'Benutzername oder Passwort falsch.',
  'too many attempts, try again later': 'Zu viele Fehlversuche. Bitte später erneut versuchen.',
  'not authenticated': 'Nicht angemeldet.',
  'unknown link': 'Dieser Link ist ungültig.',
  'friend not found': 'Freund nicht gefunden.',
  'could not reach EVCC': 'EVCC ist nicht erreichbar.',
  'could not set mode': 'Lademodus konnte nicht gesetzt werden.',
  'invalid request body': 'Ungültige Eingabe.',
  'mode must be smart or now': 'Ungültiger Lademodus.',
  'name and vehicleName are required': 'Name und Fahrzeugname sind erforderlich.',
  'pin must be 4 to 12 digits': 'Die PIN muss aus 4 bis 12 Ziffern bestehen.',
  'rates must not be negative': 'Preise dürfen nicht negativ sein.',
  'settlement not found': 'Abrechnung nicht gefunden.',
  'settlement time must not be in the future': 'Der Abrechnungszeitpunkt darf nicht in der Zukunft liegen.',
  'internal error': 'Interner Fehler.',
}

async function request<T>(method: string, url: string, body?: unknown): Promise<T> {
  const res = await fetch(url, {
    method,
    credentials: 'include',
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    let message = `Anfrage fehlgeschlagen (Status ${res.status})`
    try {
      const data = (await res.json()) as { error?: string }
      if (data.error) message = ERROR_MESSAGES[data.error] ?? data.error
    } catch {
      // response body wasn't JSON; keep the generic message
    }
    throw new ApiError(message, res.status)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

// --- Friend-facing API (PUBLIC_PREFIX) ---

export function getFriendStatus(uuid: string): Promise<FriendStatus> {
  return request('GET', `${PUBLIC_PREFIX}/l/${uuid}/status`)
}

export function friendLogin(uuid: string, pin: string): Promise<{ name: string }> {
  return request('POST', `${PUBLIC_PREFIX}/l/${uuid}/login`, { pin })
}

export function friendLogout(uuid: string): Promise<void> {
  return request('POST', `${PUBLIC_PREFIX}/l/${uuid}/logout`, {})
}

export function getDashboard(uuid: string): Promise<Dashboard> {
  return request('GET', `${PUBLIC_PREFIX}/l/${uuid}/dashboard`)
}

export function getSessions(uuid: string): Promise<SessionView[]> {
  return request('GET', `${PUBLIC_PREFIX}/l/${uuid}/sessions`)
}

export function getSettlements(uuid: string): Promise<Settlements> {
  return request('GET', `${PUBLIC_PREFIX}/l/${uuid}/settlements`)
}

export function setMode(uuid: string, mode: Mode): Promise<{ mode: Mode }> {
  return request('POST', `${PUBLIC_PREFIX}/l/${uuid}/mode`, { mode })
}

// --- Admin API (ADMIN_PREFIX) ---

export function adminLogin(username: string, password: string): Promise<{ username: string }> {
  return request('POST', `${ADMIN_PREFIX}/login`, { username, password })
}

export function adminLogout(): Promise<void> {
  return request('POST', `${ADMIN_PREFIX}/logout`, {})
}

export function listFriends(): Promise<Friend[]> {
  return request('GET', `${ADMIN_PREFIX}/friends`)
}

/** EVCC vehicle names seen in sessions and not yet assigned to a friend. */
export function listVehicles(): Promise<string[]> {
  return request('GET', `${ADMIN_PREFIX}/vehicles`)
}

export function createFriend(input: CreateFriendInput): Promise<CreatedFriend> {
  return request('POST', `${ADMIN_PREFIX}/friends`, input)
}

export function updateFriend(id: string, input: CreateFriendInput): Promise<Friend> {
  return request('PUT', `${ADMIN_PREFIX}/friends/${id}`, input)
}

export function regenerateLink(id: string): Promise<RegeneratedLink> {
  return request('POST', `${ADMIN_PREFIX}/friends/${id}/regenerate-link`, {})
}

// Without pin the server generates a random one.
export function resetPin(id: string, pin?: string): Promise<{ pin: string }> {
  return request('POST', `${ADMIN_PREFIX}/friends/${id}/reset-pin`, pin ? { pin } : {})
}

export function deleteFriend(id: string): Promise<void> {
  return request('DELETE', `${ADMIN_PREFIX}/friends/${id}`)
}

export function viewAsDashboard(id: string): Promise<Dashboard> {
  return request('GET', `${ADMIN_PREFIX}/friends/${id}/view-as/dashboard`)
}

export function viewAsSessions(id: string): Promise<SessionView[]> {
  return request('GET', `${ADMIN_PREFIX}/friends/${id}/view-as/sessions`)
}

export function viewAsSettlements(id: string): Promise<Settlements> {
  return request('GET', `${ADMIN_PREFIX}/friends/${id}/view-as/settlements`)
}

/** Records that the friend has paid everything up to `at` (ISO string; server default: now). */
export function createSettlement(id: string, at?: string): Promise<{ id: string; at: string }> {
  return request('POST', `${ADMIN_PREFIX}/friends/${id}/settlements`, at ? { at } : {})
}

export function deleteSettlement(id: string, settlementId: string): Promise<void> {
  return request('DELETE', `${ADMIN_PREFIX}/friends/${id}/settlements/${settlementId}`)
}
