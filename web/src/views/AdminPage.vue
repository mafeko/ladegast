<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { adminLogout, listFriends } from '../lib/api'
import type { CreatedFriend, Friend } from '../lib/types'
import AdminLogin from '../components/AdminLogin.vue'
import AdminFriendForm from '../components/AdminFriendForm.vue'
import AdminFriendRow from '../components/AdminFriendRow.vue'
import CopyButton from '../components/CopyButton.vue'

const checking = ref(true)
const authenticated = ref(false)
const friends = ref<Friend[]>([])
const error = ref('')
const lastCreated = ref<CreatedFriend | null>(null)
const notice = ref('')
const noticeLink = ref('')

function friendName(id: string): string {
  return friends.value.find((f) => f.id === id)?.name ?? 'neuen Freund'
}

async function loadFriends() {
  error.value = ''
  try {
    friends.value = await listFriends()
    authenticated.value = true
  } catch {
    authenticated.value = false
  } finally {
    checking.value = false
  }
}

function onLoginSuccess() {
  authenticated.value = true
  loadFriends()
}

function clearNotice() {
  notice.value = ''
  noticeLink.value = ''
}

function onCreated(friend: CreatedFriend) {
  lastCreated.value = friend
  clearNotice()
  loadFriends()
}

// The friend's id is the link slug, so a regenerated link changes the id the
// row was keyed on - reload the list to pick up the new one.
function onRegenerated(link: string) {
  lastCreated.value = null
  notice.value = 'Neuer Link:'
  noticeLink.value = link
  loadFriends()
}

function onDeleted(friend: Friend) {
  friends.value = friends.value.filter((f) => f.id !== friend.id)
  if (lastCreated.value?.id === friend.id) lastCreated.value = null
  notice.value = `${friend.name} wurde gelöscht.`
  noticeLink.value = ''
}

async function logout() {
  try {
    await adminLogout()
  } finally {
    authenticated.value = false
    friends.value = []
    lastCreated.value = null
    clearNotice()
  }
}

function onUpdated(updated: Friend) {
  const i = friends.value.findIndex((f) => f.id === updated.id)
  if (i !== -1) friends.value[i] = updated
}

onMounted(loadFriends)
</script>

<template>
  <main>
    <div v-if="checking" class="flex min-h-dvh items-center justify-center text-sm text-gray-500 dark:text-gray-400">Lade...</div>
    <AdminLogin v-else-if="!authenticated" @success="onLoginSuccess" />
    <div v-else class="mx-auto max-w-5xl px-4 py-6 sm:py-10">
      <div class="mb-6 flex items-center justify-between gap-4">
        <h1 class="text-xl font-semibold sm:text-2xl tracking-tight">Freunde verwalten</h1>
        <button
          type="button"
          class="rounded-lg border border-gray-300 bg-white min-h-11 px-4 py-2 text-sm font-medium sm:min-h-0 text-gray-700 transition hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
          @click="logout"
        >
          Abmelden
        </button>
      </div>
      <div class="mb-6 rounded-2xl border border-gray-200 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-800 dark:bg-gray-900">
        <h3 class="mb-4 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:text-gray-400">Neuen Freund anlegen</h3>
        <AdminFriendForm @created="onCreated" />
      </div>
      <div v-if="lastCreated" class="mb-6 space-y-3 rounded-lg bg-indigo-600 px-4 py-4 text-sm text-white">
        <p class="font-medium">Zugang für {{ friendName(lastCreated.id) }} - die PIN wird nur jetzt angezeigt.</p>
        <div class="flex flex-wrap items-center gap-3">
          <a :href="lastCreated.link" target="_blank" rel="noopener" class="break-all underline">{{ lastCreated.link }}</a>
          <CopyButton :text="lastCreated.link" label="Link kopieren" on-color />
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <span>PIN:</span>
          <span class="font-mono text-xl font-semibold tracking-widest select-all">{{ lastCreated.pin }}</span>
          <CopyButton :text="lastCreated.pin" label="PIN kopieren" on-color />
        </div>
      </div>
      <div v-if="notice" class="mb-6 flex flex-wrap items-center gap-3 rounded-lg bg-indigo-600 px-4 py-3 text-sm text-white">
        <span>{{ notice }}</span>
        <template v-if="noticeLink">
          <span class="break-all select-all">{{ noticeLink }}</span>
          <CopyButton :text="noticeLink" label="Link kopieren" on-color />
        </template>
      </div>
      <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-800 dark:bg-gray-900">
        <table class="block w-full text-left text-sm md:table">
          <thead class="hidden md:table-header-group">
            <tr class="border-b border-gray-200 text-xs font-semibold tracking-wide text-gray-500 uppercase dark:border-gray-800 dark:text-gray-400">
              <th class="px-3 py-2">Name</th>
              <th class="px-3 py-2">Fahrzeug</th>
              <th class="px-3 py-2">EUR/kWh Sonne</th>
              <th class="px-3 py-2">EUR/kWh Netz</th>
              <th class="px-3 py-2">Aktionen</th>
            </tr>
          </thead>
          <tbody class="block md:table-row-group">
            <tr v-if="friends.length === 0" class="block md:table-row">
              <td colspan="5" class="block px-3 py-6 text-center text-gray-500 md:table-cell dark:text-gray-400">
                Noch keine Freunde angelegt.
              </td>
            </tr>
            <AdminFriendRow v-for="f in friends" :key="f.id" :friend="f" @updated="onUpdated" @regenerated="onRegenerated" @deleted="onDeleted" />
          </tbody>
        </table>
      </div>
      <p v-if="error" role="alert" class="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
        {{ error }}
      </p>
    </div>
  </main>
</template>
