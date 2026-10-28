<script setup lang="ts">
import { ref } from 'vue'
import { deleteFriend, regenerateLink, resetPin, updateFriend } from '../lib/api'
import type { Friend } from '../lib/types'
import AdminSettlements from './AdminSettlements.vue'
import CopyButton from './CopyButton.vue'

const props = defineProps<{ friend: Friend }>()
const emit = defineEmits<{
  updated: [friend: Friend]
  regenerated: [link: string]
  deleted: [friend: Friend]
}>()

const name = ref(props.friend.name)
const vehicleName = ref(props.friend.vehicleName)
const rateSolarPerKWh = ref(props.friend.rateSolarPerKWh)
const rateGridPerKWh = ref(props.friend.rateGridPerKWh)
const message = ref('')
const newPin = ref('')
const error = ref('')
const pinFormOpen = ref(false)
const pinInput = ref('')
const settlementsOpen = ref(false)

function resetFeedback() {
  error.value = ''
  message.value = ''
  newPin.value = ''
}

async function save() {
  resetFeedback()
  try {
    const updated = await updateFriend(props.friend.id, {
      name: name.value,
      vehicleName: vehicleName.value,
      rateSolarPerKWh: rateSolarPerKWh.value,
      rateGridPerKWh: rateGridPerKWh.value,
    })
    emit('updated', updated)
    message.value = 'Gespeichert.'
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Speichern fehlgeschlagen'
  }
}

async function onRegenerateLink() {
  if (!window.confirm(`Neuen Link für ${props.friend.name} erzeugen? Der alte Link wird sofort ungültig.`)) return
  resetFeedback()
  try {
    const result = await regenerateLink(props.friend.id)
    emit('regenerated', result.link)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Fehlgeschlagen'
  }
}

// chosen === undefined asks the server for a random PIN.
async function setPin(chosen?: string) {
  if (!window.confirm(`PIN für ${props.friend.name} ändern? Die alte PIN wird sofort ungültig.`)) return
  resetFeedback()
  try {
    const result = await resetPin(props.friend.id, chosen)
    newPin.value = result.pin
    pinFormOpen.value = false
    pinInput.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Fehlgeschlagen'
  }
}

async function onDelete() {
  if (!window.confirm(`${props.friend.name} endgültig löschen? Link und PIN funktionieren danach nicht mehr.`)) return
  resetFeedback()
  try {
    await deleteFriend(props.friend.id)
    emit('deleted', props.friend)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Löschen fehlgeschlagen'
  }
}

const inputClass =
  'rounded-lg border border-gray-300 bg-white min-h-11 px-2 py-1 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950'
const buttonClass =
  'rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-1.5 text-xs font-medium sm:min-h-0 text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800'
const cellClass =
  'flex items-center justify-between gap-4 px-3 py-1 before:text-xs before:text-gray-500 before:content-[attr(data-label)] md:table-cell md:py-2 md:before:content-none dark:before:text-gray-400'
</script>

<template>
  <tr class="block border-b border-gray-100 py-3 last:border-0 md:table-row md:py-0 dark:border-gray-800">
    <td data-label="Name" :class="cellClass">
      <input v-model="name" aria-label="Name" required class="w-36" :class="inputClass" />
    </td>
    <td data-label="Fahrzeug" :class="cellClass">
      <input v-model="vehicleName" aria-label="Fahrzeug" required class="w-32" :class="inputClass" />
    </td>
    <td data-label="Sonne (EUR/kWh)" :class="cellClass">
      <input v-model.number="rateSolarPerKWh" aria-label="Preis Sonne" type="number" step="0.01" min="0" class="w-24" :class="inputClass" />
    </td>
    <td data-label="Netz (EUR/kWh)" :class="cellClass">
      <input v-model.number="rateGridPerKWh" aria-label="Preis Netz" type="number" step="0.01" min="0" class="w-24" :class="inputClass" />
    </td>
    <td class="flex flex-wrap items-center gap-2 px-3 py-2 md:table-cell md:space-x-2 md:space-y-2 md:whitespace-normal">
      <button
        type="button"
        class="rounded-lg bg-indigo-600 min-h-11 px-3 py-1.5 text-xs font-medium sm:min-h-0 text-white transition hover:bg-indigo-500"
        @click="save"
      >
        Speichern
      </button>
      <CopyButton :text="friend.link" label="Link kopieren" />
      <button type="button" :class="buttonClass" @click="onRegenerateLink">Link neu</button>
      <button type="button" :class="buttonClass" @click="pinFormOpen = !pinFormOpen">PIN ändern</button>
      <button type="button" :class="buttonClass" :aria-expanded="settlementsOpen" @click="settlementsOpen = !settlementsOpen">
        Abrechnung
      </button>
      <RouterLink :to="`/admin/view-as/${friend.id}`" class="text-xs font-medium text-indigo-600 hover:underline dark:text-indigo-400">
        Ansehen als
      </RouterLink>
      <button
        type="button"
        class="rounded-lg border border-red-300 bg-white min-h-11 px-3 py-1.5 text-xs font-medium sm:min-h-0 text-red-600 transition hover:bg-red-50 dark:border-red-900 dark:bg-gray-900 dark:text-red-400 dark:hover:bg-red-950/40"
        @click="onDelete"
      >
        Löschen
      </button>
    </td>
  </tr>
  <tr v-if="pinFormOpen || settlementsOpen || message || newPin || error" class="block md:table-row">
    <td colspan="5" class="block space-y-2 px-3 pb-3 md:table-cell">
      <form v-if="pinFormOpen" class="flex flex-wrap items-center gap-2" @submit.prevent="setPin(pinInput)">
        <input
          v-model="pinInput"
          aria-label="Neue PIN"
          inputmode="numeric"
          pattern="[0-9]{4,12}"
          maxlength="12"
          required
          autocomplete="off"
          placeholder="Neue PIN (4-12 Ziffern)"
          class="w-56"
          :class="inputClass"
        />
        <button type="submit" class="rounded-lg bg-indigo-600 min-h-11 px-3 py-1.5 text-xs font-medium sm:min-h-0 text-white transition hover:bg-indigo-500">
          PIN setzen
        </button>
        <button type="button" :class="buttonClass" @click="setPin()">Zufällige PIN</button>
        <button type="button" :class="buttonClass" @click="pinFormOpen = false">Abbrechen</button>
      </form>
      <AdminSettlements v-if="settlementsOpen" :friend-id="friend.id" :friend-name="friend.name" />
      <p v-if="message" class="text-sm text-gray-500 dark:text-gray-400">{{ message }}</p>
      <div v-if="newPin" class="flex flex-wrap items-center gap-3 rounded-lg bg-indigo-600 px-4 py-3 text-white">
        <span class="text-sm">Neue PIN für {{ friend.name }}:</span>
        <span class="font-mono text-xl font-semibold tracking-widest select-all">{{ newPin }}</span>
        <CopyButton :text="newPin" on-color />
      </div>
      <p v-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
    </td>
  </tr>
</template>
