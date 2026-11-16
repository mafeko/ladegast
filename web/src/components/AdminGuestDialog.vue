<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { deleteGuest, regenerateLink, resetPin, updateGuest } from '../lib/api'
import type { Guest } from '../lib/types'
import AdminSettlements from './AdminSettlements.vue'
import CopyButton from './CopyButton.vue'

const props = defineProps<{ guest: Guest }>()
const emit = defineEmits<{
  close: []
  updated: [guest: Guest]
  regenerated: [link: string]
  deleted: [guest: Guest]
}>()

const name = ref(props.guest.name)
const vehicleName = ref(props.guest.vehicleName)
const rateSolarPerKWh = ref(props.guest.rateSolarPerKWh)
const rateGridPerKWh = ref(props.guest.rateGridPerKWh)
const message = ref('')
const newPin = ref('')
const error = ref('')
const pinFormOpen = ref(false)
const pinInput = ref('')
const nameInput = ref<HTMLInputElement | null>(null)

// Save is only offered once something differs from the stored guest.
const dirty = computed(
  () =>
    name.value !== props.guest.name ||
    vehicleName.value !== props.guest.vehicleName ||
    rateSolarPerKWh.value !== props.guest.rateSolarPerKWh ||
    rateGridPerKWh.value !== props.guest.rateGridPerKWh,
)

watch([name, vehicleName, rateSolarPerKWh, rateGridPerKWh], () => (message.value = ''))

function resetFeedback() {
  error.value = ''
  message.value = ''
  newPin.value = ''
}

async function save() {
  resetFeedback()
  try {
    const updated = await updateGuest(props.guest.id, {
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
  if (!window.confirm(`Neuen Link für ${props.guest.name} erzeugen? Der alte Link wird sofort ungültig.`)) return
  resetFeedback()
  try {
    const result = await regenerateLink(props.guest.id)
    emit('regenerated', result.link)
    emit('close')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Fehlgeschlagen'
  }
}

// chosen === undefined asks the server for a random PIN.
async function setPin(chosen?: string) {
  if (!window.confirm(`PIN für ${props.guest.name} ändern? Die alte PIN wird sofort ungültig.`)) return
  resetFeedback()
  try {
    const result = await resetPin(props.guest.id, chosen)
    newPin.value = result.pin
    pinFormOpen.value = false
    pinInput.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Fehlgeschlagen'
  }
}

async function onDelete() {
  if (!window.confirm(`${props.guest.name} endgültig löschen? Link und PIN funktionieren danach nicht mehr.`)) return
  resetFeedback()
  try {
    await deleteGuest(props.guest.id)
    emit('deleted', props.guest)
    emit('close')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Löschen fehlgeschlagen'
  }
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

onMounted(() => {
  document.addEventListener('keydown', onKey)
  document.body.style.overflow = 'hidden'
  // Don't pop the on-screen keyboard on touch devices.
  if (window.matchMedia?.('(hover: hover)').matches) nameInput.value?.focus()
})
onUnmounted(() => {
  document.removeEventListener('keydown', onKey)
  document.body.style.overflow = ''
})
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-end justify-center sm:items-center sm:p-6" role="presentation">
    <div class="absolute inset-0 bg-gray-400/60 backdrop-blur-sm dark:bg-black/60" @click="emit('close')" />
    <div
      role="dialog"
      aria-modal="true"
      :aria-label="`${guest.name} bearbeiten`"
      class="relative flex max-h-[92dvh] w-full max-w-2xl flex-col overflow-hidden rounded-t-[20px] bg-white shadow-xl sm:max-h-[90dvh] sm:rounded-[16px] dark:bg-gray-900"
    >
      <header class="flex items-start justify-between gap-4 px-5 pt-5 pb-3 sm:px-6">
        <div class="min-w-0">
          <p class="label-caps">Gast bearbeiten</p>
          <h2 class="heading truncate">{{ guest.name }}</h2>
        </div>
        <button type="button" aria-label="Schließen" class="btn-outline !px-3" @click="emit('close')">✕</button>
      </header>

      <div class="space-y-6 overflow-y-auto px-5 pb-4 sm:px-6">
        <section aria-labelledby="sec-data">
          <h3 id="sec-data" class="label-caps mb-3">Daten</h3>
          <form id="guest-edit-form" class="grid grid-cols-2 gap-3" @submit.prevent="save">
            <div class="col-span-2 sm:col-span-1">
              <label class="field-label" for="edit-name">Name</label>
              <input id="edit-name" ref="nameInput" v-model="name" aria-label="Name" required class="field" />
            </div>
            <div class="col-span-2 sm:col-span-1">
              <label class="field-label" for="edit-vehicle">Fahrzeug</label>
              <input id="edit-vehicle" v-model="vehicleName" aria-label="Fahrzeug" required class="field" />
            </div>
            <div>
              <label class="field-label" for="edit-solar">Sonne (EUR/kWh)</label>
              <input id="edit-solar" v-model.number="rateSolarPerKWh" aria-label="Preis Sonne" type="number" step="0.01" min="0" class="field" />
            </div>
            <div>
              <label class="field-label" for="edit-grid">Netz (EUR/kWh)</label>
              <input id="edit-grid" v-model.number="rateGridPerKWh" aria-label="Preis Netz" type="number" step="0.01" min="0" class="field" />
            </div>
          </form>
        </section>

        <section aria-labelledby="sec-access" class="space-y-3">
          <h3 id="sec-access" class="label-caps">Zugang</h3>
          <div class="flex flex-wrap items-center gap-2 sm:flex-nowrap">
            <input :value="guest.link" readonly aria-label="Link" class="field min-w-0 flex-1 basis-full font-mono !text-xs sm:basis-0" @focus="($event.target as HTMLInputElement).select()" />
            <CopyButton :text="guest.link" label="Link kopieren" />
          </div>
          <div class="flex flex-wrap gap-2">
            <button type="button" class="btn-outline" @click="onRegenerateLink">Link neu</button>
            <button type="button" class="btn-outline" :aria-expanded="pinFormOpen" @click="pinFormOpen = !pinFormOpen">PIN ändern</button>
            <RouterLink :to="`/admin/view-as/${guest.id}`" class="btn-outline">Ansehen als</RouterLink>
          </div>
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
              class="field !w-56"
            />
            <button type="submit" class="btn-primary">PIN setzen</button>
            <button type="button" class="btn-outline" @click="setPin()">Zufällige PIN</button>
            <button type="button" class="btn-outline" @click="pinFormOpen = false">Abbrechen</button>
          </form>
          <div v-if="newPin" class="alert-success flex flex-wrap items-center gap-3">
            <span>Neue PIN für {{ guest.name }}:</span>
            <span class="font-mono text-xl font-semibold tracking-widest select-all">{{ newPin }}</span>
            <CopyButton :text="newPin" />
          </div>
        </section>

        <section aria-labelledby="sec-settle" class="space-y-3">
          <h3 id="sec-settle" class="label-caps">Abrechnung</h3>
          <AdminSettlements :guest-id="guest.id" :guest-name="guest.name" />
        </section>

        <section aria-labelledby="sec-danger" class="space-y-3">
          <h3 id="sec-danger" class="label-caps">Gefahrenzone</h3>
          <div class="flex flex-wrap items-center justify-between gap-3 rounded-[10px] border border-evcc-red/30 p-3">
            <p class="text-sm text-gray-500 dark:text-gray-300">Link und PIN funktionieren danach nicht mehr.</p>
            <button type="button" class="btn-danger" @click="onDelete">Löschen</button>
          </div>
        </section>

        <p v-if="error" role="alert" class="alert-error">{{ error }}</p>
      </div>

      <footer class="flex items-center justify-end gap-3 border-t border-gray-200 px-5 py-3 sm:px-6 dark:border-gray-800">
        <p v-if="message" role="status" class="mr-auto text-sm font-bold text-indigo-600 dark:text-indigo-400">✓ {{ message }}</p>
        <button type="button" class="btn-outline" @click="emit('close')">Schließen</button>
        <button type="button" class="btn-primary" :disabled="!dirty" @click="save">Speichern</button>
      </footer>
    </div>
  </div>
</template>
