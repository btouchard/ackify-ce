<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'

const { t } = useI18n()
usePageTitle('notFound.title')
</script>

<template>
  <div class="min-h-screen bg-background text-foreground flex items-center justify-center">
    <div class="text-center">
      <h1 class="text-6xl font-bold text-muted-foreground">404</h1>
      <p class="text-xl text-foreground mt-4">{{ t('notFound.title') }}</p>
      <p class="text-sm text-muted-foreground mt-2">{{ t('notFound.description') }}</p>
      <router-link
        to="/"
        class="mt-6 inline-block text-primary hover:text-primary/80 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded px-2 py-1"
      >
        {{ t('notFound.home') }}
      </router-link>
    </div>
  </div>
</template>