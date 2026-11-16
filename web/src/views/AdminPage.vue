<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { adminLogout, listGuests } from '../lib/api'
import type { CreatedGuest, Guest } from '../lib/types'
import AdminLogin from '../components/AdminLogin.vue'
import AdminGuestForm from '../components/AdminGuestForm.vue'
import AdminGuestRow from '../components/AdminGuestRow.vue'
import CopyButton from '../components/CopyButton.vue'
import BrandHeader from '../components/BrandHeader.vue'

const checking = ref(true)
const authenticated = ref(false)
const guests = ref<Guest[]>([])
const error = ref('')
const lastCreated = ref<CreatedGuest | null>(null)
const createOpen = ref(false)
const notice = ref('')
const noticeLink = ref('')

function guestName(id: string): string {
  return guests.value.find((f) => f.id === id)?.name ?? 'neuen Gast'
}

async function loadGuests() {
  error.value = ''
  try {
    guests.value = await listGuests()
    authenticated.value = true
  } catch {
    authenticated.value = false
  } finally {
    checking.value = false
  }
}

function onLoginSuccess() {
  authenticated.value = true
  loadGuests()
}

function clearNotice() {
  notice.value = ''
  noticeLink.value = ''
}

function onCreated(guest: CreatedGuest) {
  lastCreated.value = guest
  createOpen.value = false
  clearNotice()
  loadGuests()
}

// The guest's id is the link slug, so a regenerated link changes the id the
// row was keyed on - reload the list to pick up the new one.
function onRegenerated(link: string) {
  lastCreated.value = null
  notice.value = 'Neuer Link:'
  noticeLink.value = link
  loadGuests()
}

function onDeleted(guest: Guest) {
  guests.value = guests.value.filter((f) => f.id !== guest.id)
  if (lastCreated.value?.id === guest.id) lastCreated.value = null
  notice.value = `${guest.name} wurde gelöscht.`
  noticeLink.value = ''
}

async function logout() {
  try {
    await adminLogout()
  } finally {
    authenticated.value = false
    guests.value = []
    lastCreated.value = null
    clearNotice()
  }
}

function onUpdated(updated: Guest) {
  const i = guests.value.findIndex((f) => f.id === updated.id)
  if (i !== -1) guests.value[i] = updated
}

onMounted(loadGuests)
</script>

<template>
  <main>
    <div v-if="checking" class="flex min-h-dvh items-center justify-center text-sm text-gray-400">Lade...</div>
    <AdminLogin v-else-if="!authenticated" @success="onLoginSuccess" />
    <div v-else class="mx-auto max-w-5xl px-4 py-6 sm:py-10">
      <div class="mb-6 flex items-center justify-between gap-4">
        <BrandHeader subtitle="Admin" />
        <button
          type="button"
          class="btn-outline"
          @click="logout"
        >
          Abmelden
        </button>
      </div>
      <div v-if="lastCreated" class="alert-success mb-6 space-y-3">
        <p class="font-medium">Zugang für {{ guestName(lastCreated.id) }} - die PIN wird nur jetzt angezeigt.</p>
        <div class="flex flex-wrap items-center gap-3">
          <a :href="lastCreated.link" target="_blank" rel="noopener" class="break-all underline">{{ lastCreated.link }}</a>
          <CopyButton :text="lastCreated.link" label="Link kopieren" />
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <span>PIN:</span>
          <span class="font-mono text-xl font-semibold tracking-widest select-all">{{ lastCreated.pin }}</span>
          <CopyButton :text="lastCreated.pin" label="PIN kopieren" />
        </div>
      </div>
      <div v-if="notice" class="alert-success mb-6 flex flex-wrap items-center gap-3">
        <span>{{ notice }}</span>
        <template v-if="noticeLink">
          <span class="break-all select-all">{{ noticeLink }}</span>
          <CopyButton :text="noticeLink" label="Link kopieren" />
        </template>
      </div>
      <div v-if="createOpen" class="box mb-6">
        <h2 class="heading mb-4">Neuen Gast anlegen</h2>
        <AdminGuestForm @created="onCreated" @cancel="createOpen = false" />
      </div>
      <div class="box">
        <div class="mb-4 flex items-center justify-between gap-4">
          <h2 class="heading">Gäste verwalten</h2>
          <button v-if="!createOpen" type="button" class="btn-primary" @click="createOpen = true">+ Neuer Gast</button>
        </div>
        <table class="block w-full text-left text-sm md:table">
          <thead class="hidden md:table-header-group">
            <tr class="border-b border-gray-200 text-xs font-bold tracking-wide text-gray-400 uppercase dark:border-gray-800">
              <th class="px-3 py-2">Name</th>
              <th class="px-3 py-2">Fahrzeug</th>
              <th class="px-3 py-2">Sonne (EUR/kWh)</th>
              <th class="px-3 py-2">Netz (EUR/kWh)</th>
              <th class="px-3 py-2"><span class="sr-only">Aktionen</span></th>
            </tr>
          </thead>
          <tbody class="block md:table-row-group">
            <tr v-if="guests.length === 0" class="block md:table-row">
              <td colspan="5" class="block px-3 py-6 text-center text-gray-400 md:table-cell">
                Noch keine Gäste angelegt.
              </td>
            </tr>
            <AdminGuestRow v-for="f in guests" :key="f.id" :guest="f" @updated="onUpdated" @regenerated="onRegenerated" @deleted="onDeleted" />
          </tbody>
        </table>
      </div>
      <p v-if="error" role="alert" class="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
    </div>
  </main>
</template>
