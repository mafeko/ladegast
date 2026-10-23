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
  <tr>
    <td>{{ friend.name }}</td>
    <td>{{ friend.vehicleName }}</td>
    <td><input v-model.number="rateSolarPerKWh" type="number" step="0.01" min="0" /></td>
    <td><input v-model.number="rateGridPerKWh" type="number" step="0.01" min="0" /></td>
    <td>
      <button type="button" @click="saveRates">Speichern</button>
      <button type="button" @click="onRegenerateLink">Link neu</button>
      <button type="button" @click="onResetPin">PIN zurücksetzen</button>
      <RouterLink :to="`/admin/view-as/${friend.id}`">Ansehen als</RouterLink>
    </td>
  </tr>
  <tr v-if="message || error">
    <td colspan="5">
      <p v-if="message">{{ message }}</p>
      <p v-if="error" role="alert">{{ error }}</p>
    </td>
  </tr>
</template>
