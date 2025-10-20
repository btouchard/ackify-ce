<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import AppShell from './components/layout/AppShell.vue'
import NotificationToast from './components/NotificationToast.vue'
import { useRoute } from 'vue-router'
import { computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const isEmbedPage = computed(() => route.meta.isEmbed === true)

const authStore = useAuthStore()

// Check authentication status on app mount
onMounted(() => {
  authStore.checkAuth()
})
</script>

<template>
  <div id="app">
    <template v-if="isEmbedPage">
      <router-view />
    </template>

    <template v-else>
      <AppShell>
        <router-view />
      </AppShell>
    </template>

    <NotificationToast />
  </div>
</template>

<style>
</style>
