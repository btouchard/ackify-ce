<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { documentService } from '@/services/documents'
import { extractError } from '@/services/http'
import { ArrowRight } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'

const router = useRouter()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const documentUrl = ref('')
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

const handleSubmit = async () => {
  errorMessage.value = null

  if (!documentUrl.value.trim()) {
    const homeRoute = '/'
    if (isAuthenticated.value) {
      await router.push(homeRoute)
    } else {
      await authStore.startOAuthLogin(homeRoute)
    }
    return
  }

  try {
    isSubmitting.value = true

    // Use findOrCreateDocument instead of createDocument to avoid duplicates
    // This will search for existing document by reference first, then create if not found
    const response = await documentService.findOrCreateDocument(documentUrl.value.trim())

    const homeRoute = `/?doc=${response.docId}`
    if (isAuthenticated.value) {
      await router.push(homeRoute)
    } else {
      await authStore.startOAuthLogin(homeRoute)
    }
  } catch (error) {
    errorMessage.value = extractError(error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <div
      v-if="errorMessage"
      class="w-full rounded-lg bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 p-4 text-sm text-red-800 dark:text-red-200"
    >
      {{ errorMessage }}
    </div>

    <div class="flex w-full flex-col gap-3 sm:flex-row">
      <Input
        v-model="documentUrl"
        type="text"
        data-testid="doc-form-input"
        :placeholder="$t('admin.documentForm.placeholder')"
        class="flex-1 h-11"
        :disabled="isSubmitting"
        @keyup.enter="handleSubmit"
      />
      <Button
        @click="handleSubmit"
        size="lg"
        data-testid="doc-form-submit"
        class="group whitespace-nowrap"
        :disabled="isSubmitting"
      >
        <span v-if="isSubmitting">{{ $t('admin.documentForm.submitting') }}</span>
        <span v-else>{{ $t('admin.documentForm.submit') }}</span>
        <ArrowRight v-if="!isSubmitting" :size="16" class="ml-2 transition-transform group-hover:translate-x-1" />
      </Button>
    </div>
  </div>
</template>
