<script setup lang="ts">
import { ref } from 'vue'
import { adminLogin } from '../lib/api'

const emit = defineEmits<{ success: [] }>()

const username = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function onSubmit() {
  error.value = ''
  submitting.value = true
  try {
    await adminLogin(username.value, password.value)
    emit('success')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login fehlgeschlagen'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="onSubmit">
    <label for="username">Benutzername</label>
    <input id="username" v-model="username" required />
    <label for="password">Passwort</label>
    <input id="password" v-model="password" type="password" required />
    <button type="submit" :disabled="submitting">Anmelden</button>
    <p v-if="error" role="alert">{{ error }}</p>
  </form>
</template>
