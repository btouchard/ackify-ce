<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface CardProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<CardProps>()
</script>

<template>
  <div :class="cn('bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700', props.class)">
    <slot />
  </div>
</template>
