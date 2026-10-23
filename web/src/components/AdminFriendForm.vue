<script setup lang="ts">
import { ref } from 'vue'
import { createFriend } from '../lib/api'
import type { CreatedFriend } from '../lib/types'

const emit = defineEmits<{ created: [friend: CreatedFriend] }>()

const name = ref('')
const vehicleName = ref('')
const rateSolarPerKWh = ref(0.1)
const rateGridPerKWh = ref(0.35)
const error = ref('')
const submitting = ref(false)

async function onSubmit() {
  error.value = ''
  submitting.value = true
  try {
    const created = await createFriend({
      name: name.value,
      vehicleName: vehicleName.value,
      rateSolarPerKWh: rateSolarPerKWh.value,
      rateGridPerKWh: rateGridPerKWh.value,
    })
    emit('created', created)
    name.value = ''
    vehicleName.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Anlegen fehlgeschlagen'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="onSubmit">
    <label for="name">Name</label>
    <input id="name" v-model="name" required />
    <label for="vehicleName">EVCC-Fahrzeugname</label>
    <input id="vehicleName" v-model="vehicleName" required />
    <label for="rateSolar">Preis Sonne (EUR/kWh)</label>
    <input id="rateSolar" v-model.number="rateSolarPerKWh" type="number" step="0.01" min="0" required />
    <label for="rateGrid">Preis Netz (EUR/kWh)</label>
    <input id="rateGrid" v-model.number="rateGridPerKWh" type="number" step="0.01" min="0" required />
    <button type="submit" :disabled="submitting">Freund anlegen</button>
    <p v-if="error" role="alert">{{ error }}</p>
  </form>
</template>
