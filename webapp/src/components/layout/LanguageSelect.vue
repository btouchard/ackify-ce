<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronDown } from 'lucide-vue-next'
import { setLocale } from '@/i18n'
import Button from '@/components/ui/Button.vue'

const { locale, t } = useI18n()
const dropdownOpen = ref(false)

interface Language {
  code: string
  name: string
  flag: string
}

const languages: Language[] = [
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'en', name: 'English', flag: '🇬🇧' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
  { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
  { code: 'it', name: 'Italiano', flag: '🇮🇹' }
]

const currentLanguage = computed(() => {
  return languages.find(lang => lang.code === locale.value) || languages[0]
})

const toggleDropdown = () => {
  dropdownOpen.value = !dropdownOpen.value
}

const selectLanguage = (langCode: string) => {
  setLocale(langCode)
  dropdownOpen.value = false
}

const closeDropdown = () => {
  dropdownOpen.value = false
}

// Handle keyboard navigation
const handleKeydown = (event: KeyboardEvent, langCode: string) => {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    selectLanguage(langCode)
  }
}
</script>

<template>
  <div class="relative">
    <Button
      @click="toggleDropdown"
      variant="ghost"
      size="sm"
      class="rounded-md px-3 py-2 text-sm font-medium hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
      :aria-label="t('language.select')"
      aria-haspopup="true"
      :aria-expanded="dropdownOpen"
    >
      <span class="mr-1.5 text-lg leading-none" aria-hidden="true">{{ currentLanguage?.flag || '🇫🇷' }}</span>
      <span class="hidden sm:inline">{{ currentLanguage?.name || 'Français' }}</span>
      <ChevronDown :size="16" class="ml-1 text-muted-foreground" />
    </Button>

    <!-- Dropdown menu -->
    <transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
    >
      <div
        v-if="dropdownOpen"
        @click.stop
        v-click-outside="closeDropdown"
        class="absolute right-0 mt-2 w-48 origin-top-right clay-card rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none z-50"
        role="menu"
        aria-orientation="vertical"
        :aria-label="t('language.select')"
      >
        <div class="p-1">
          <button
            v-for="lang in languages"
            :key="lang.code"
            @click="selectLanguage(lang.code)"
            @keydown="(e) => handleKeydown(e, lang.code)"
            :class="[
              'flex w-full items-center space-x-3 rounded-md px-3 py-2 text-sm transition-colors',
              locale === lang.code
                ? 'bg-accent text-primary font-medium'
                : 'hover:bg-accent/50'
            ]"
            role="menuitem"
            :tabindex="0"
          >
            <span class="text-lg leading-none" aria-hidden="true">{{ lang.flag }}</span>
            <span class="flex-1 text-left">{{ lang.name }}</span>
            <span
              v-if="locale === lang.code"
              class="ml-auto h-1.5 w-1.5 rounded-full bg-primary"
              aria-label="selected"
            ></span>
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
/* Ensure emojis render consistently */
button span {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji', sans-serif;
}
</style>
