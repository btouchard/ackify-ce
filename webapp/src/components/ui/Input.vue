<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { type InputHTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface InputProps {
  class?: InputHTMLAttributes['class']
  type?: string
  modelValue?: string | number
}

const props = withDefaults(defineProps<InputProps>(), {
  type: 'text'
})

const emits = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emits('update:modelValue', target.value)
}
</script>

<template>
  <input
    :type="type"
    :value="modelValue"
    @input="handleInput"
    :class="cn(
      'flex h-10 w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-4 py-2.5 text-sm text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 transition-colors',
      props.class
    )"
  />
</template>
