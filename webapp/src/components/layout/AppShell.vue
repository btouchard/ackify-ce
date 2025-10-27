<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import AppFooter from './AppFooter.vue'
import SkipToContent from '../accessibility/SkipToContent.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col bg-background relative">
    <SkipToContent />
    <AppHeader class="flex-shrink-0" />

    <main id="main-content" class="flex-grow w-full">
      <slot />
    </main>

    <AppFooter class="flex-shrink-0" />
  </div>
</template>
