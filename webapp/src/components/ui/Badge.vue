<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { type HTMLAttributes } from 'vue'
import { type VariantProps, cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900',
        secondary: 'border-transparent bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300',
        destructive: 'border-transparent bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400',
        outline: 'border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300',
        success: 'border-transparent bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400',
        warning: 'border-transparent bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
        info: 'border-transparent bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

interface BadgeProps {
  variant?: VariantProps<typeof badgeVariants>['variant']
  class?: HTMLAttributes['class']
}

const props = defineProps<BadgeProps>()
</script>

<template>
  <div :class="cn(badgeVariants({ variant }), props.class)">
    <slot />
  </div>
</template>
