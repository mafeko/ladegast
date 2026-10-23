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
    <button type="button" :disabled="pending !== null" @click="choose('smart')">PV-Überschuss</button>
    <button type="button" :disabled="pending !== null" @click="choose('now')">Schnellladen</button>
    <p v-if="error" role="alert">{{ error }}</p>
  </div>
</template>
