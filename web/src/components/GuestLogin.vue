<script setup lang="ts">
import { ref } from 'vue'
import { guestLogin } from '../lib/api'
import BrandHeader from './BrandHeader.vue'

const props = defineProps<{ uuid: string }>()
const emit = defineEmits<{ success: [name: string] }>()

const pin = ref('')
const error = ref('')
const submitting = ref(false)

async function onSubmit() {
  error.value = ''
  submitting.value = true
  try {
    const result = await guestLogin(props.uuid, pin.value)
    emit('success', result.name)
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
        <BrandHeader />
        <p class="mt-3 text-sm text-gray-400">Gib deine PIN ein, um fortzufahren.</p>
      </div>
      <div>
        <label for="pin" class="field-label">PIN</label>
        <input
          id="pin"
          v-model="pin"
          type="password"
          inputmode="numeric"
          autocomplete="one-time-code"
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
