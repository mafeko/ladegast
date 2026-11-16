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
    class="btn"
    :class="
      onColor
        ? 'border-white/40 text-white hover:bg-white/10'
        : 'btn-outline'
    "
    @click="copy"
  >
    {{ copied ? 'Kopiert' : (label ?? 'Kopieren') }}
  </button>
</template>
