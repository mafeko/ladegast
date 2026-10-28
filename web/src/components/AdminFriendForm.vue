<script setup lang="ts">
import { ref } from 'vue'
import { createFriend } from '../lib/api'
import type { CreatedFriend } from '../lib/types'

const emit = defineEmits<{ created: [friend: CreatedFriend] }>()

const name = ref('')
const vehicleName = ref('')
const rateSolarPerKWh = ref(0.1)
const rateGridPerKWh = ref(0.35)
const pin = ref('')
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
      pin: pin.value || undefined,
    })
    emit('created', created)
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
        <label for="name" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">Name</label>
        <input
          id="name"
          v-model="name"
          required
          class="w-full rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-2 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
      <div>
        <label for="vehicleName" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">EVCC-Fahrzeugname</label>
        <input
          id="vehicleName"
          v-model="vehicleName"
          required
          class="w-full rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-2 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
      <div>
        <label for="rateSolar" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">Preis Sonne (EUR/kWh)</label>
        <input
          id="rateSolar"
          v-model.number="rateSolarPerKWh"
          type="number"
          step="0.01"
          min="0"
          required
          class="w-full rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-2 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
      <div>
        <label for="rateGrid" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">Preis Netz (EUR/kWh)</label>
        <input
          id="rateGrid"
          v-model.number="rateGridPerKWh"
          type="number"
          step="0.01"
          min="0"
          required
          class="w-full rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-2 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
      <div class="sm:col-span-2">
        <label for="pin" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">PIN (optional)</label>
        <input
          id="pin"
          v-model="pin"
          inputmode="numeric"
          pattern="[0-9]{4,12}"
          maxlength="12"
          autocomplete="off"
          placeholder="Leer lassen für eine zufällige 6-stellige PIN"
          title="4 bis 12 Ziffern"
          class="w-full rounded-lg border border-gray-300 bg-white min-h-11 px-3 py-2 text-base focus:outline-none sm:min-h-0 sm:text-sm focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
    </div>
    <button
      type="submit"
      :disabled="submitting"
      class="rounded-lg bg-indigo-600 min-h-11 px-4 py-2 text-sm font-medium sm:min-h-0 text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
    >
      Freund anlegen
    </button>
    <p v-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
      {{ error }}
    </p>
  </form>
</template>
