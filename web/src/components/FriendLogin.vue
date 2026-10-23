<script setup lang="ts">
import { ref } from 'vue'
import { friendLogin } from '../lib/api'

const props = defineProps<{ uuid: string }>()
const emit = defineEmits<{ success: [name: string] }>()

const pin = ref('')
const error = ref('')
const submitting = ref(false)

async function onSubmit() {
  error.value = ''
  submitting.value = true
  try {
    const result = await friendLogin(props.uuid, pin.value)
    emit('success', result.name)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login fehlgeschlagen'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="onSubmit">
    <label for="pin">PIN</label>
    <input id="pin" v-model="pin" type="password" inputmode="numeric" autocomplete="one-time-code" required />
    <button type="submit" :disabled="submitting">Anmelden</button>
    <p v-if="error" role="alert">{{ error }}</p>
  </form>
</template>
