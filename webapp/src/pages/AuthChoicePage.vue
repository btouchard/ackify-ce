<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'
import { Mail, LogIn, Loader2, AlertCircle, CheckCircle2 } from 'lucide-vue-next'
import AppLogo from '@/components/AppLogo.vue'

const { t } = useI18n()
usePageTitle('auth.choice.title')

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const loading = ref(false)
const magicLinkSent = ref(false)
const errorMessage = ref('')

// Lire les flags d'authentification depuis les variables globales injectées dans index.html
const oauthEnabled = (window as any).ACKIFY_OAUTH_ENABLED || false
const magicLinkEnabled = (window as any).ACKIFY_MAGICLINK_ENABLED || false

const redirectTo = computed(() => {
  return (route.query.redirect as string) || '/'
})

function checkAuthMethods() {
  // Si aucune méthode disponible
  if (!oauthEnabled && !magicLinkEnabled) {
    errorMessage.value = t('auth.error.no_method_available')
    return
  }

  // Si une seule méthode disponible (OAuth), rediriger automatiquement
  const methods = [oauthEnabled, magicLinkEnabled].filter(Boolean)
  if (methods.length === 1 && oauthEnabled) {
    loginWithOAuth()
  }
  // Si seulement MagicLink, l'utilisateur doit quand même entrer son email (pas de redirection auto)
}

onMounted(async () => {
  // Si déjà connecté, rediriger
  if (!authStore.initialized) {
    await authStore.checkAuth()
  }
  if (authStore.isAuthenticated) {
    await router.push(redirectTo.value)
    return
  }

  // Vérifier les méthodes d'authentification disponibles
  checkAuthMethods()
})

async function loginWithOAuth() {
  loading.value = true
  errorMessage.value = ''
  localStorage.setItem('preferredAuthMethod', 'oauth')

  try {
    await authStore.startOAuthLogin(redirectTo.value)
  } catch (error: any) {
    errorMessage.value = error.message || t('auth.oauth.error')
  } finally {
    loading.value = false
  }
}

async function requestMagicLink() {
  if (!email.value || !isValidEmail(email.value)) {
    errorMessage.value = t('auth.magiclink.error_invalid_email')
    return
  }

  loading.value = true
  errorMessage.value = ''
  magicLinkSent.value = false
  localStorage.setItem('preferredAuthMethod', 'magiclink')

  try {
    const response = await fetch('/api/v1/auth/magic-link/request', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: email.value,
        redirectTo: redirectTo.value,
      }),
    })

    if (!response.ok) {
      throw new Error(t('auth.magiclink.error_send'))
    }

    magicLinkSent.value = true
  } catch (error: any) {
    errorMessage.value = error.message || t('auth.magiclink.error_send')
  } finally {
    loading.value = false
  }
}

function isValidEmail(email: string): boolean {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return re.test(email)
}
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex items-center justify-center px-4 sm:px-6 py-12">
    <div class="max-w-md w-full space-y-8">
      <!-- Header with logo -->
      <div class="text-center">
        <div class="flex justify-center mb-6">
          <AppLogo size="lg" :show-version="false" />
        </div>
        <h1 class="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-slate-100">
          {{ t('auth.choice.title') }}
        </h1>
        <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
          {{ t('auth.choice.subtitle') }}
        </p>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMessage" class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
        <div class="flex items-start">
          <AlertCircle :size="20" class="mr-3 mt-0.5 text-red-600 dark:text-red-400 flex-shrink-0" />
          <div class="flex-1">
            <h3 class="font-medium text-red-900 dark:text-red-200">{{ t('common.error') }}</h3>
            <p class="mt-1 text-sm text-red-700 dark:text-red-300">{{ errorMessage }}</p>
          </div>
        </div>
      </div>

      <!-- Success Alert (Magic Link Sent) -->
      <div v-if="magicLinkSent" class="bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl p-4">
        <div class="flex items-start">
          <CheckCircle2 :size="20" class="mr-3 mt-0.5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
          <div class="flex-1">
            <h3 class="font-medium text-emerald-900 dark:text-emerald-200">{{ t('auth.magiclink.sent.title') }}</h3>
            <p class="mt-1 text-sm text-emerald-700 dark:text-emerald-300">
              {{ t('auth.magiclink.sent.message') }}
            </p>
            <p class="mt-2 text-xs text-emerald-600 dark:text-emerald-400">
              {{ t('auth.magiclink.sent.expire') }}
            </p>
          </div>
        </div>
      </div>

      <!-- OAuth Login Card -->
      <div v-if="oauthEnabled" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
            <LogIn :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('auth.oauth.title') }}</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('auth.oauth.description') }}</p>
          </div>
        </div>
        <button
          @click="loginWithOAuth"
          :disabled="loading"
          class="w-full trust-gradient text-white font-medium rounded-lg px-4 py-3 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
        >
          <Loader2 v-if="loading" class="w-4 h-4 animate-spin mr-2" />
          {{ t('auth.oauth.button') }}
        </button>
      </div>

      <!-- Magic Link Login Card -->
      <div v-if="magicLinkEnabled" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
            <Mail :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('auth.magiclink.title') }}</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('auth.magiclink.description') }}</p>
          </div>
        </div>
        <form @submit.prevent="requestMagicLink" class="space-y-4">
          <div>
            <label for="email" class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              {{ t('auth.magiclink.email_label') }}
            </label>
            <input
              id="email"
              v-model="email"
              type="email"
              required
              :disabled="loading"
              :placeholder="t('auth.magiclink.email_placeholder')"
              class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-3 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
          >
            <Loader2 v-if="loading" class="w-4 h-4 animate-spin mr-2" />
            <Mail v-else class="w-4 h-4 mr-2" />
            {{ t('auth.magiclink.button') }}
          </button>
        </form>
      </div>

      <!-- Privacy note -->
      <p class="text-center text-xs text-slate-500 dark:text-slate-400">
        {{ t('auth.choice.privacy') }}
      </p>
    </div>
  </div>
</template>
