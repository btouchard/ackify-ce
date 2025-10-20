<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface TableRowProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<TableRowProps>()
</script>

<template>
  <tr :class="cn('border-b border-border/40 transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted', props.class)">
    <slot />
  </tr>
</template>
