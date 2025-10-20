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
      'flex h-10 w-full rounded-md clay-input px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      props.class
    )"
  />
</template>
