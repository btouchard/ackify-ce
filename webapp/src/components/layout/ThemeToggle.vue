<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Moon, Sun } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const theme = ref<'light' | 'dark'>('light')

const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
}

const applyTheme = () => {
  const root = document.documentElement

  if (theme.value === 'dark') {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }

  localStorage.setItem('theme', theme.value)
}

// Watch theme changes and apply immediately
watch(theme, () => {
  applyTheme()
}, { immediate: false })

onMounted(() => {
  const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  // Set initial theme
  theme.value = savedTheme || (prefersDark ? 'dark' : 'light')

  // Apply immediately
  applyTheme()
})
</script>

<template>
  <Button
    @click="toggleTheme"
    variant="ghost"
    size="icon"
    class="rounded-full"
    :aria-label="t('theme.toggle')"
  >
    <Sun v-if="theme === 'dark'" :size="20" class="text-muted-foreground" />
    <Moon v-else :size="20" class="text-muted-foreground" />
  </Button>
</template>
