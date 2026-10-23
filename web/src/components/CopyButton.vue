<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ text: string; label?: string; onColor?: boolean }>()

const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    // clipboard unavailable (insecure context / denied): the text stays selectable
  }
}
</script>

<template>
  <button
    type="button"
    class="rounded-lg border px-3 py-1.5 text-xs font-medium transition"
    :class="
      onColor
        ? 'border-white/40 text-white hover:bg-white/10'
        : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800'
    "
    @click="copy"
  >
    {{ copied ? 'Kopiert' : (label ?? 'Kopieren') }}
  </button>
</template>
