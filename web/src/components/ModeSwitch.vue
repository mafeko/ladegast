<script setup lang="ts">
import { ref } from 'vue'
import { setMode } from '../lib/api'
import type { Mode } from '../lib/types'

const props = defineProps<{ uuid: string }>()
const emit = defineEmits<{ changed: [mode: Mode] }>()

const pending = ref<Mode | null>(null)
const error = ref('')

async function choose(mode: Mode) {
  error.value = ''
  pending.value = mode
  try {
    const result = await setMode(props.uuid, mode)
    emit('changed', result.mode)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Umschalten fehlgeschlagen'
  } finally {
    pending.value = null
  }
}
</script>

<template>
  <div>
    <h3 class="mb-3 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Lademodus</h3>
    <div class="flex gap-2">
      <button
        type="button"
        :disabled="pending !== null"
        class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
        @click="choose('smart')"
      >
        PV-Überschuss
      </button>
      <button
        type="button"
        :disabled="pending !== null"
        class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
        @click="choose('now')"
      >
        Schnellladen
      </button>
    </div>
    <p v-if="error" role="alert" class="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
      {{ error }}
    </p>
  </div>
</template>
