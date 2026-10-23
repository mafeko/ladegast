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
  <div class="flex min-h-screen items-center justify-center p-4">
    <form
      class="w-full max-w-sm space-y-5 rounded-2xl border border-gray-200 bg-white p-8 shadow-sm dark:border-gray-800 dark:bg-gray-900"
      @submit.prevent="onSubmit"
    >
      <div>
        <h1 class="text-xl font-semibold tracking-tight">Ladegast Admin</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Internes Verwaltungsportal.</p>
      </div>
      <div>
        <label for="username" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">Benutzername</label>
        <input
          id="username"
          v-model="username"
          required
          class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
      <div>
        <label for="password" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">Passwort</label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:border-gray-700 dark:bg-gray-950"
        />
      </div>
      <button
        type="submit"
        :disabled="submitting"
        class="w-full rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
      >
        Anmelden
      </button>
      <p v-if="error" role="alert" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
    </form>
  </div>
</template>
