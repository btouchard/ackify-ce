<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'
import AppLogo from '@/components/AppLogo.vue'

const { t } = useI18n()
usePageTitle('notFound.title')
</script>

<template>
  <div class="min-h-full flex items-center justify-center px-4 py-16 sm:py-24">
    <div class="text-center max-w-md mx-auto">
      <!-- 404 with logo -->
      <div class="mb-8">
        <div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-slate-100 dark:bg-slate-800 mb-6">
          <span class="text-4xl font-bold text-slate-400 dark:text-slate-500">404</span>
        </div>
      </div>

      <!-- Title -->
      <h1 class="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-white mb-3">
        {{ t('notFound.title') }}
      </h1>

      <!-- Description -->
      <p class="text-slate-500 dark:text-slate-400 mb-8 text-base sm:text-lg">
        {{ t('notFound.description') }}
      </p>

      <!-- Back to home button -->
      <router-link
        to="/"
        class="inline-flex items-center gap-2 trust-gradient text-white font-medium rounded-lg px-6 py-3 hover:opacity-90 transition-opacity min-h-[48px]"
      >
        <svg class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        {{ t('notFound.home') }}
      </router-link>

      <!-- Small logo at bottom -->
      <div class="mt-12 flex justify-center">
        <AppLogo size="sm" />
      </div>
    </div>
  </div>
</template>
