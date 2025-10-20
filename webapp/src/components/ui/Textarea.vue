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
      'flex min-h-[80px] w-full rounded-md clay-input px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      props.class
    )"
  />
</template>
