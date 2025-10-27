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
        icon: 'h-5 w-5',
        text: 'text-base'
      }
    case 'lg':
      return {
        icon: 'h-10 w-10',
        text: 'text-2xl'
      }
    case 'md':
    default:
      return {
        icon: 'h-8 w-8',
        text: 'text-xl'
      }
  }
})
</script>

<template>
  <div class="flex items-center space-x-2">
    <svg
      :class="[sizeClasses.icon, 'text-primary']"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
    <div v-if="showText" class="flex flex-col">
      <span
        :class="[sizeClasses.text, textClass || 'font-bold text-foreground']"
      >
        Ackify
      </span>
      <span
        v-if="showVersion && appVersion"
        class="text-xs text-muted-foreground leading-none -mt-0.5"
      >
        {{ appVersion }}
      </span>
    </div>
  </div>
</template>
