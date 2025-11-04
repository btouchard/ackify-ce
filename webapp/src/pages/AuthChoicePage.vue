<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'
import { Mail, LogIn, Loader2, AlertCircle, CheckCircle2 } from 'lucide-vue-next'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Button from '@/components/ui/Button.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertTitle from '@/components/ui/AlertTitle.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'

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
  <div class="min-h-full box-border flex items-center justify-center bg-background py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div class="text-center">
        <h1 class="text-3xl font-bold text-foreground">
          {{ t('auth.choice.title') }}
        </h1>
        <p class="mt-2 text-sm text-muted-foreground">
          {{ t('auth.choice.subtitle') }}
        </p>
      </div>

      <Alert v-if="errorMessage" variant="destructive">
        <AlertCircle class="h-4 w-4" />
        <AlertTitle>{{ t('common.error') }}</AlertTitle>
        <AlertDescription>{{ errorMessage }}</AlertDescription>
      </Alert>

      <Alert v-if="magicLinkSent" variant="default" class="border-green-200 bg-green-50">
        <CheckCircle2 class="h-4 w-4 text-green-600" />
        <AlertTitle class="text-green-800">{{ t('auth.magiclink.sent.title') }}</AlertTitle>
        <AlertDescription class="text-green-700">
          {{ t('auth.magiclink.sent.message') }}
          <br>
          <span class="text-xs text-green-600">
            {{ t('auth.magiclink.sent.expire') }}
          </span>
        </AlertDescription>
      </Alert>

      <!-- OAuth Login -->
      <Card v-if="oauthEnabled">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <LogIn class="h-5 w-5" />
            {{ t('auth.oauth.title') }}
          </CardTitle>
          <CardDescription>
            {{ t('auth.oauth.description') }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            @click="loginWithOAuth"
            :disabled="loading"
            class="w-full"
            size="lg"
          >
            <Loader2 v-if="loading" class="h-4 w-4 animate-spin mr-2" />
            {{ t('auth.oauth.button') }}
          </Button>
        </CardContent>
      </Card>

      <!-- Magic Link Login -->
      <Card v-if="magicLinkEnabled">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Mail class="h-5 w-5" />
            {{ t('auth.magiclink.title') }}
          </CardTitle>
          <CardDescription>
            {{ t('auth.magiclink.description') }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="requestMagicLink" class="space-y-4">
            <div>
              <label for="email" class="block text-sm font-medium text-foreground mb-1">
                {{ t('auth.magiclink.email_label') }}
              </label>
              <input
                id="email"
                v-model="email"
                type="email"
                required
                :disabled="loading"
                :placeholder="t('auth.magiclink.email_placeholder')"
                class="w-full px-3 py-2 border border-border rounded-md shadow-sm bg-input text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
              />
            </div>
            <Button
              type="submit"
              :disabled="loading"
              class="w-full"
              size="lg"
              variant="outline"
            >
              <Loader2 v-if="loading" class="h-4 w-4 animate-spin mr-2" />
              <Mail v-else class="h-4 w-4 mr-2" />
              {{ t('auth.magiclink.button') }}
            </Button>
          </form>
        </CardContent>
      </Card>

      <p class="text-center text-xs text-muted-foreground">
        {{ t('auth.choice.privacy') }}
      </p>
    </div>
  </div>
</template>
