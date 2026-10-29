<script setup lang="ts">
import { formatEUR } from '../lib/format'
import type { FriendInfo } from '../lib/types'

defineProps<{ friend: FriendInfo }>()
</script>

<template>
  <section
    data-testid="price-info"
    class="mb-6 rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-800 dark:bg-gray-900"
  >
    <h3 class="text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Deine Preise</h3>
    <dl class="mt-2 flex flex-wrap gap-x-8 gap-y-1 text-sm">
      <div class="flex items-center gap-2">
        <dt class="flex items-center gap-1.5 text-gray-600 dark:text-gray-300">
          <span class="inline-block size-3 rounded-sm bg-amber-400"></span>Sonne
        </dt>
        <dd class="font-semibold">{{ formatEUR(friend.rateSolarPerKWh) }}<span class="text-xs font-normal text-gray-500 dark:text-gray-400"> / kWh</span></dd>
      </div>
      <div class="flex items-center gap-2">
        <dt class="flex items-center gap-1.5 text-gray-600 dark:text-gray-300">
          <span class="inline-block size-3 rounded-sm bg-slate-500 dark:bg-slate-400"></span>Netz
        </dt>
        <dd class="font-semibold">{{ formatEUR(friend.rateGridPerKWh) }}<span class="text-xs font-normal text-gray-500 dark:text-gray-400"> / kWh</span></dd>
      </div>
    </dl>
    <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">Kosten = Sonnenstrom × Sonnenpreis + Netzstrom × Netzpreis.</p>
  </section>
</template>
