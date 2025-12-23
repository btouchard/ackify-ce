<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  size?: 'sm' | 'md' | 'lg'
  showText?: boolean
  showVersion?: boolean
  textClass?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  showText: true,
  showVersion: false,
  textClass: ''
})

const appVersion = computed(() => {
  return (window as any).ACKIFY_VERSION || ''
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return {
        logo: 'w-6 h-6',
        text: 'text-base'
      }
    case 'lg':
      return {
        logo: 'w-10 h-10',
        text: 'text-2xl'
      }
    case 'md':
    default:
      return {
        logo: 'w-8 h-8',
        text: 'text-xl'
      }
  }
})
</script>

<template>
  <div class="flex items-center gap-2">
    <!-- Logo icon -->
    <img
      src="/logo.svg"
      alt="Ackify"
      :class="[sizeClasses.logo, 'flex-shrink-0']"
    />

    <!-- Text -->
    <div v-if="showText" class="flex flex-col">
      <span
        :class="[sizeClasses.text, textClass || 'font-bold text-slate-900 dark:text-slate-50']"
      >
        Ackify
      </span>
      <span
        v-if="showVersion && appVersion"
        class="text-xs text-slate-500 dark:text-slate-400 leading-none -mt-0.5"
      >
        {{ appVersion }}
      </span>
    </div>
  </div>
</template>
