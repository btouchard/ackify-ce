<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { X } from 'lucide-vue-next'
import Card from './Card.vue'
import CardHeader from './CardHeader.vue'
import CardTitle from './CardTitle.vue'
import CardContent from './CardContent.vue'
import Button from './Button.vue'
import Alert from './Alert.vue'
import AlertDescription from './AlertDescription.vue'

export interface ConfirmDialogProps {
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'default' | 'destructive' | 'warning'
  loading?: boolean
}

withDefaults(defineProps<ConfirmDialogProps>(), {
  confirmText: 'Confirmer',
  cancelText: 'Annuler',
  variant: 'default',
  loading: false,
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="emit('cancel')">
    <Card :class="['max-w-md w-full', variant === 'destructive' ? 'border-destructive' : variant === 'warning' ? 'border-orange-500' : 'border-primary']">
      <CardHeader>
        <div class="flex items-center justify-between">
          <CardTitle :class="{ 'text-destructive': variant === 'destructive', 'text-orange-600': variant === 'warning' }">
            {{ title }}
          </CardTitle>
          <Button variant="ghost" size="icon" @click="emit('cancel')" :disabled="loading">
            <X :size="20" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div class="space-y-4">
          <Alert :variant="variant === 'destructive' ? 'destructive' : 'default'"
                 :class="variant === 'warning' ? 'border-orange-500 bg-orange-50 dark:bg-orange-900/20' : variant === 'destructive' ? 'border-destructive' : ''">
            <AlertDescription :class="{ 'text-orange-800 dark:text-orange-200': variant === 'warning' }">
              {{ message }}
            </AlertDescription>
          </Alert>

          <div class="flex justify-end space-x-3 pt-2">
            <Button type="button" variant="outline" @click="emit('cancel')" :disabled="loading">
              {{ cancelText }}
            </Button>
            <Button
              @click="emit('confirm')"
              :variant="variant === 'default' ? 'default' : 'destructive'"
              :disabled="loading"
            >
              {{ loading ? 'Chargement...' : confirmText }}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
