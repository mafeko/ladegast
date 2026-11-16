<script setup lang="ts">
import { ref } from 'vue'
import { adminLogin } from '../lib/api'
import BrandHeader from './BrandHeader.vue'

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
  <div class="flex min-h-dvh items-center justify-center p-4">
    <form
      class="box w-full max-w-sm space-y-5 sm:p-8"
      @submit.prevent="onSubmit"
    >
      <div>
        <BrandHeader subtitle="Admin" />
        <p class="mt-3 text-sm text-gray-400">Internes Verwaltungsportal.</p>
      </div>
      <div>
        <label for="username" class="field-label">Benutzername</label>
        <input
          id="username"
          v-model="username"
          required
          class="field"
        />
      </div>
      <div>
        <label for="password" class="field-label">Passwort</label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          class="field"
        />
      </div>
      <button
        type="submit"
        :disabled="submitting"
        class="btn-primary w-full"
      >
        Anmelden
      </button>
      <p v-if="error" role="alert" class="alert-error">
        {{ error }}
      </p>
    </form>
  </div>
</template>
