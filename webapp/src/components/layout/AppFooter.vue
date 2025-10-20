<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Github, Mail } from 'lucide-vue-next'

const { t } = useI18n()
</script>

<template>
  <footer class="mt-auto border-t border-border/40 clay-card backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="grid grid-cols-1 gap-8 md:grid-cols-4">
        <!-- Brand -->
        <div class="space-y-4">
          <div class="flex items-center space-x-2">
            <svg class="h-8 w-8 text-primary" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span class="text-xl font-bold text-foreground">Ackify</span>
          </div>
          <p class="text-sm text-muted-foreground">
            {{ t('footer.description') }}
          </p>
        </div>

        <!-- Navigation -->
        <div>
          <h3 class="mb-4 text-sm font-semibold text-foreground">{{ t('footer.navigation.title') }}</h3>
          <ul class="space-y-3">
            <li>
              <router-link to="/" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('nav.home') }}
              </router-link>
            </li>
            <li>
              <router-link to="/signatures" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('nav.myConfirmations') }}
              </router-link>
            </li>
          </ul>
        </div>

        <!-- Resources -->
        <div>
          <h3 class="mb-4 text-sm font-semibold text-foreground">{{ t('footer.resources.title') }}</h3>
          <ul class="space-y-3">
            <li>
              <a href="https://github.com/btouchard/ackify-ce" target="_blank" rel="noopener noreferrer" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('footer.resources.documentation') }}
              </a>
            </li>
            <li>
              <a href="https://github.com/btouchard/ackify-ce" target="_blank" rel="noopener noreferrer" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('footer.resources.apiReference') }}
              </a>
            </li>
            <li>
              <a href="https://github.com/btouchard/ackify-ce/issues" target="_blank" rel="noopener noreferrer" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('footer.resources.support') }}
              </a>
            </li>
          </ul>
        </div>

        <!-- Legal & Contact -->
        <div>
          <h3 class="mb-4 text-sm font-semibold text-foreground">{{ t('footer.legal.title') }}</h3>
          <ul class="space-y-3">
            <li>
              <a href="#" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('footer.legal.terms') }}
              </a>
            </li>
            <li>
              <a href="#" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('footer.legal.privacy') }}
              </a>
            </li>
            <li>
              <a href="#" class="text-sm text-muted-foreground hover:text-primary transition-colors">
                {{ t('footer.legal.contact') }}
              </a>
            </li>
          </ul>

          <!-- Social Icons -->
          <div class="mt-6 flex space-x-4">
            <a href="https://github.com/btouchard/ackify-ce" target="_blank" rel="noopener noreferrer"
               class="text-muted-foreground hover:text-primary transition-colors"
               aria-label="GitHub">
              <Github :size="20" />
            </a>
            <a href="mailto:support@ackify.eu" class="text-muted-foreground hover:text-primary transition-colors" aria-label="Email">
              <Mail :size="20" />
            </a>
          </div>
        </div>
      </div>

      <div class="mt-12 border-t border-border/40 pt-8">
        <p class="text-center text-sm text-muted-foreground">
          &copy; {{ new Date().getFullYear() }} Ackify. {{ t('footer.copyright') }}
          <span class="text-xs">{{ t('footer.license') }}</span>
        </p>
      </div>
    </div>
  </footer>
</template>
