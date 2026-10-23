<script setup lang="ts">
import { ref } from 'vue'
import { regenerateLink, resetPin, updateFriend } from '../lib/api'
import type { Friend } from '../lib/types'

const props = defineProps<{ friend: Friend }>()
const emit = defineEmits<{ updated: [friend: Friend] }>()

const rateSolarPerKWh = ref(props.friend.rateSolarPerKWh)
const rateGridPerKWh = ref(props.friend.rateGridPerKWh)
const message = ref('')
const error = ref('')

async function saveRates() {
  error.value = ''
  message.value = ''
  try {
    const updated = await updateFriend(props.friend.id, {
      name: props.friend.name,
      vehicleName: props.friend.vehicleName,
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
  error.value = ''
  try {
    const result = await regenerateLink(props.friend.id)
    message.value = `Neuer Link: ${result.link}`
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Fehlgeschlagen'
  }
}

async function onResetPin() {
  error.value = ''
  try {
    const result = await resetPin(props.friend.id)
    message.value = `Neue PIN: ${result.pin}`
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Fehlgeschlagen'
  }
}
</script>

<template>
  <tr class="border-b border-gray-100 last:border-0 dark:border-gray-800">
    <td class="px-3 py-2 font-medium">{{ friend.name }}</td>
    <td class="px-3 py-2 text-gray-500 dark:text-gray-400">{{ friend.vehicleName }}</td>
    <td class="px-3 py-2">
      <input
        v-model.number="rateSolarPerKWh"
        type="number"
        step="0.01"
        min="0"
        class="w-24 rounded-lg border border-gray-300 bg-white px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
      />
    </td>
    <td class="px-3 py-2">
      <input
        v-model.number="rateGridPerKWh"
        type="number"
        step="0.01"
        min="0"
        class="w-24 rounded-lg border border-gray-300 bg-white px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
      />
    </td>
    <td class="space-x-2 px-3 py-2 whitespace-nowrap">
      <button
        type="button"
        class="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-indigo-500"
        @click="saveRates"
      >
        Speichern
      </button>
      <button
        type="button"
        class="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
        @click="onRegenerateLink"
      >
        Link neu
      </button>
      <button
        type="button"
        class="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
        @click="onResetPin"
      >
        PIN zurücksetzen
      </button>
      <RouterLink :to="`/admin/view-as/${friend.id}`" class="text-xs font-medium text-indigo-600 hover:underline dark:text-indigo-400">
        Ansehen als
      </RouterLink>
    </td>
  </tr>
  <tr v-if="message || error">
    <td colspan="5" class="px-3 pb-3">
      <p v-if="message" class="text-sm text-gray-500 dark:text-gray-400">{{ message }}</p>
      <p v-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
    </td>
  </tr>
</template>
