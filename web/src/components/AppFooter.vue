<script setup lang="ts">
// Minimal footer shown on every page (admin and guest).
const repoUrl = 'https://github.com/mafeko/ladegast'
</script>

<template>
  <footer class="px-4 pt-2 pb-6 text-center text-xs text-gray-400">
    <a :href="repoUrl" target="_blank" rel="noopener noreferrer" class="hover:text-gray-600 hover:underline dark:hover:text-gray-200">
      Ladegast on GitHub
    </a>
  </footer>
</template>
