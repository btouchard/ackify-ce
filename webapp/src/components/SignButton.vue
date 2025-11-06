<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<template>
  <div class="sign-button-container">
    <button
      v-if="!isSigned"
      @click="handleSign"
      :disabled="loading || disabled || !docId"
      :class="buttonClasses"
      type="button"
    >
      <svg
        v-if="loading"
        class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        ></circle>
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        ></path>
      </svg>
      <svg
        v-else
        class="-ml-1 mr-3 h-5 w-5"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
        />
      </svg>
      {{ loading ? $t('signButton.signing') : $t('signButton.confirmAction') }}
    </button>

    <div v-else class="signed-status">
      <div class="flex items-center justify-center space-x-2 text-green-700">
        <svg
          class="h-6 w-6"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span class="font-semibold">{{ $t('signButton.confirmed') }}</span>
      </div>
      <p v-if="signedAt" class="mt-2 text-sm text-muted-foreground text-center">
        {{ $t('signButton.on') }} {{ formatDate(signedAt) }}
      </p>
    </div>

    <div v-if="error" class="mt-4 text-red-600 text-sm text-center">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSignatureStore } from '@/stores/signatures'
import { useAuthStore } from '@/stores/auth'

interface Signature {
  userEmail: string
  signedAt: string
}

interface Props {
  docId?: string
  referer?: string
  disabled?: boolean
  signatures?: Signature[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  signed: [docId: string]
  error: [error: string]
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const signatureStore = useSignatureStore()
const loading = ref(false)
const error = ref<string | null>(null)
const isSigned = ref(false)
const signedAt = ref<string | null>(null)

// Check if current user has signed based on signatures list
async function checkIfSigned() {
  // Initialize auth store if not already done (for public pages)
  if (!authStore.initialized) {
    try {
      await authStore.checkAuth()
    } catch {
      // Ignore errors - user is not authenticated
    }
  }

  if (!props.signatures || !authStore.user?.email) {
    isSigned.value = false
    signedAt.value = null
    return
  }

  const userSignature = props.signatures.find(
    sig => sig.userEmail === authStore.user?.email
  )

  isSigned.value = !!userSignature
  signedAt.value = userSignature?.signedAt || null
}

// Watch for changes in signatures prop or auth state
watch(() => [props.signatures, authStore.user], checkIfSigned, { immediate: true, deep: true })

const buttonClasses = computed(() => {
  return [
    'inline-flex items-center justify-center px-6 py-3 border border-transparent text-base font-medium rounded-md shadow-sm text-white transition-colors',
    loading.value || props.disabled || !props.docId
      ? 'bg-indigo-400 cursor-not-allowed'
      : 'bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500',
  ]
})

async function handleSign() {
  if (!props.docId) {
    error.value = t('signButton.error.missingDocId')
    return
  }

  // Check if user is authenticated
  if (!authStore.initialized) {
    try {
      await authStore.checkAuth()
    } catch {
      // Ignore errors - user is not authenticated
    }
  }

  // If not authenticated, redirect to OAuth login
  if (!authStore.isAuthenticated) {
    try {
      await authStore.startOAuthLogin(window.location.pathname + window.location.search)
    } catch (err: any) {
      error.value = t('signButton.error.authFailed')
      emit('error', error.value)
    }
    return
  }

  loading.value = true
  error.value = null

  try {
    await signatureStore.createSignature({
      docId: props.docId,
      referer: props.referer,
    })

    isSigned.value = true
    signedAt.value = new Date().toISOString()
    emit('signed', props.docId)
  } catch (err: any) {
    const errorMessage =
      err.response?.data?.error?.message || 'Impossible de confirmer la lecture'
    error.value = errorMessage
    emit('error', errorMessage)
  } finally {
    loading.value = false
  }
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<style scoped>
.sign-button-container {
  width: 100%;
}

.signed-status {
  padding: 1rem;
  background-color: #f0fdf4;
  border: 2px solid #86efac;
  border-radius: 0.5rem;
}
</style>
