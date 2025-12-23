<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { type TextareaHTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface TextareaProps {
  class?: TextareaHTMLAttributes['class']
  modelValue?: string
  rows?: number
}

const props = withDefaults(defineProps<TextareaProps>(), {
  rows: 4
})

const emits = defineEmits<{
  'update:modelValue': [value: string]
}>()

const handleInput = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  emits('update:modelValue', target.value)
}
</script>

<template>
  <textarea
    :value="modelValue"
    :rows="rows"
    @input="handleInput"
    :class="cn(
      'flex min-h-[80px] w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-4 py-2.5 text-sm text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 transition-colors resize-none',
      props.class
    )"
  />
</template>
