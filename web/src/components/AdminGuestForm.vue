<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createGuest, listVehicles } from '../lib/api'
import type { CreatedGuest } from '../lib/types'

const emit = defineEmits<{ created: [guest: CreatedGuest]; cancel: [] }>()

const name = ref('')
const vehicleName = ref('')
const rateSolarPerKWh = ref(0.1)
const rateGridPerKWh = ref(0.35)
const pin = ref('')
const error = ref('')
const submitting = ref(false)
const vehicles = ref<string[]>([])

async function loadVehicles() {
  try {
    vehicles.value = (await listVehicles()) ?? []
  } catch {
    vehicles.value = [] // suggestions are optional; free text still works
  }
}
onMounted(loadVehicles)

async function onSubmit() {
  error.value = ''
  submitting.value = true
  try {
    const created = await createGuest({
      name: name.value,
      vehicleName: vehicleName.value,
      rateSolarPerKWh: rateSolarPerKWh.value,
      rateGridPerKWh: rateGridPerKWh.value,
      pin: pin.value || undefined,
    })
    emit('created', created)
    void loadVehicles()
    name.value = ''
    vehicleName.value = ''
    pin.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Anlegen fehlgeschlagen'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form class="space-y-4" @submit.prevent="onSubmit">
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <label for="name" class="field-label">Name</label>
        <input
          id="name"
          v-model="name"
          required
          class="field"
        />
      </div>
      <div>
        <label for="vehicleName" class="field-label">EVCC-Fahrzeugname</label>
        <input
          id="vehicleName"
          v-model="vehicleName"
          list="vehicleSuggestions"
          required
          autocomplete="off"
          class="field"
        />
        <datalist id="vehicleSuggestions">
          <option v-for="v in vehicles" :key="v" :value="v" />
        </datalist>
      </div>
      <div>
        <label for="rateSolar" class="field-label">Preis Sonne (EUR/kWh)</label>
        <input
          id="rateSolar"
          v-model.number="rateSolarPerKWh"
          type="number"
          step="0.01"
          min="0"
          required
          class="field"
        />
      </div>
      <div>
        <label for="rateGrid" class="field-label">Preis Netz (EUR/kWh)</label>
        <input
          id="rateGrid"
          v-model.number="rateGridPerKWh"
          type="number"
          step="0.01"
          min="0"
          required
          class="field"
        />
      </div>
      <div class="sm:col-span-2">
        <label for="pin" class="field-label">PIN (optional)</label>
        <input
          id="pin"
          v-model="pin"
          inputmode="numeric"
          pattern="[0-9]{4,12}"
          maxlength="12"
          autocomplete="off"
          placeholder="Leer lassen für eine zufällige 6-stellige PIN"
          title="4 bis 12 Ziffern"
          class="field"
        />
      </div>
    </div>
    <div class="flex flex-wrap gap-2">
      <button type="submit" :disabled="submitting" class="btn-primary">Gast anlegen</button>
      <button type="button" class="btn-outline" @click="emit('cancel')">Abbrechen</button>
    </div>
    <p v-if="error" role="alert" class="alert-error">
      {{ error }}
    </p>
  </form>
</template>
