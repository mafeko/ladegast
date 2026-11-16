<script setup lang="ts">
// "Ladegast" wordmark shown on admin and public (guest) pages.
// Styled after the evcc header: bold uppercase title next to a green icon.
defineProps<{ subtitle?: string }>()
</script>

<template>
  <div class="flex items-center gap-3">
    <svg class="h-8 w-8 shrink-0 text-indigo-600 dark:text-indigo-400" viewBox="0 0 32 32" fill="none" aria-hidden="true">
      <rect width="32" height="32" rx="8" fill="currentColor" />
      <path d="M17.5 5 9 18h6l-1.5 9L23 13h-6.2z" fill="#fff" class="dark:fill-gray-950" />
    </svg>
    <div class="leading-tight">
      <span class="block text-xl font-bold tracking-tight uppercase">Ladegast</span>
      <span v-if="subtitle" class="block text-xs font-bold tracking-wide text-gray-400 uppercase">{{ subtitle }}</span>
    </div>
  </div>
</template>
