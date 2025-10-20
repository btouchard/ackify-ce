<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { type VariantProps, cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const alertVariants = cva(
  'relative w-full rounded-lg border p-4',
  {
    variants: {
      variant: {
        default: 'bg-background text-foreground',
        success: 'border-l-4 border-green-400 bg-green-50 text-green-900 dark:bg-green-900/20 dark:text-green-400',
        destructive: 'border-l-4 border-red-400 bg-red-50 text-red-900 dark:bg-red-900/20 dark:text-red-400',
        warning: 'border-l-4 border-yellow-400 bg-yellow-50 text-yellow-900 dark:bg-yellow-900/20 dark:text-yellow-400',
        info: 'border-l-4 border-blue-400 bg-blue-50 text-blue-900 dark:bg-blue-900/20 dark:text-blue-400',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

interface AlertProps {
  variant?: VariantProps<typeof alertVariants>['variant']
  class?: HTMLAttributes['class']
}

const props = defineProps<AlertProps>()
</script>

<template>
  <div :class="cn(alertVariants({ variant }), props.class)" role="alert">
    <slot />
  </div>
</template>
