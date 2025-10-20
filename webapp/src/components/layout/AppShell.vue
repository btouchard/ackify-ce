<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import AppFooter from './AppFooter.vue'
import SkipToContent from '../accessibility/SkipToContent.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col bg-background">
    <SkipToContent />
    <AppHeader />

    <main id="main-content" class="flex-1 min-h-[60vh]">
      <slot />
    </main>

    <AppFooter />
  </div>
</template>
