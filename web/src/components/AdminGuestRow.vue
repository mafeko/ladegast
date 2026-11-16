<script setup lang="ts">
import { ref } from 'vue'
import type { Guest } from '../lib/types'
import AdminGuestDialog from './AdminGuestDialog.vue'

defineProps<{ guest: Guest }>()
const emit = defineEmits<{
  updated: [guest: Guest]
  regenerated: [link: string]
  deleted: [guest: Guest]
}>()

const open = ref(false)
const rate = (v: number) => new Intl.NumberFormat('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(v)

// Mobile: compact card (name + vehicle on top, both prices side by side).
// md+: regular table cells.
const cell = 'px-3 md:table-cell md:py-3'
const priceCell =
  cell + ' inline-block w-1/2 text-sm before:mr-1.5 before:text-xs before:font-bold before:text-gray-400 before:uppercase before:content-[attr(data-label)] md:w-auto md:text-base md:before:content-none'
</script>

<template>
  <tr
    class="relative block cursor-pointer border-b border-gray-200 py-3 transition-colors last:border-0 hover:bg-gray-50 md:table-row md:py-0 dark:border-gray-800 dark:hover:bg-gray-950/50"
    @click="open = true"
  >
    <td :class="[cell, 'block pr-24 font-bold md:pr-3']">{{ guest.name }}</td>
    <td :class="[cell, 'block pb-1 text-sm text-gray-400 md:text-base md:text-inherit']">{{ guest.vehicleName }}</td>
    <td data-label="Sonne" :class="priceCell">{{ rate(guest.rateSolarPerKWh) }}</td>
    <td data-label="Netz" :class="priceCell">{{ rate(guest.rateGridPerKWh) }}</td>
    <td class="absolute top-2 right-2 md:static md:table-cell md:px-3 md:py-2 md:text-right">
      <div class="flex justify-end gap-2">
        <RouterLink
          :to="`/admin/view-as/${guest.id}`"
          class="btn-outline !px-3"
          :aria-label="`Ansehen als ${guest.name}`"
          :title="`Ansehen als ${guest.name}`"
          @click.stop
        >
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
        </RouterLink>
        <button type="button" class="btn-outline !px-3 md:!px-4" :aria-label="`${guest.name} bearbeiten`" @click.stop="open = true">
          <span class="md:hidden" aria-hidden="true">✎</span>
          <span class="hidden md:inline">Bearbeiten</span>
        </button>
      </div>
    </td>
  </tr>
  <Teleport to="body">
    <AdminGuestDialog
      v-if="open"
      :guest="guest"
      @close="open = false"
      @updated="emit('updated', $event)"
      @regenerated="emit('regenerated', $event)"
      @deleted="emit('deleted', $event)"
    />
  </Teleport>
</template>
