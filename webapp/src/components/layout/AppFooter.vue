<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Github } from 'lucide-vue-next'
import AppLogo from '@/components/AppLogo.vue'

const { t } = useI18n()
</script>

<template>
  <footer class="mt-auto border-t border-border clay-card relative z-10">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <div class="flex flex-col items-center justify-between gap-4 sm:flex-row sm:gap-6">
        <!-- Brand & Description -->
        <div class="flex items-center gap-3 text-center sm:text-left">
          <AppLogo size="sm" :show-text="true" :show-version="true" />
          <span class="text-sm text-muted-foreground leading-tight max-w-md">{{ t('footer.description') }}</span>
        </div>

        <!-- Links & Copyright -->
        <div class="flex flex-wrap items-center justify-center gap-4 text-sm sm:flex-nowrap">
          <a
            href="https://github.com/btouchard/ackify-ce"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-1.5 text-muted-foreground hover:text-primary transition-colors"
            :aria-label="t('footer.resources.documentation')"
          >
            <Github :size="16" />
            <span>GitHub</span>
          </a>

          <span class="text-muted-foreground/50">•</span>

          <span class="text-muted-foreground whitespace-nowrap">
            &copy; {{ new Date().getFullYear() }} {{ t('footer.license') }}
          </span>
        </div>
      </div>
    </div>
  </footer>
</template>
