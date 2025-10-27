<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { availableWebhookEvents, createWebhook, getWebhook, updateWebhook, type WebhookInput, type Webhook } from '@/services/webhooks'
import { extractError } from '@/services/http'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'
import { Loader2, Save, ArrowLeft } from 'lucide-vue-next'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.name === 'admin-webhook-new')
const id = computed(() => Number(route.params.id))

const loading = ref(false)
const saving = ref(false)
const error = ref('')

const title = ref('')
const targetUrl = ref('')
const secret = ref('')
const active = ref(true)
const events = ref<string[]>([])
const description = ref('')

async function load() {
  if (isNew.value) return
  try {
    loading.value = true
    const resp = await getWebhook(id.value)
    const wh = resp.data as Webhook
    title.value = wh.title || ''
    targetUrl.value = wh.targetUrl
    active.value = wh.active
    events.value = [...(wh.events||[])]
    description.value = wh.description || ''
  } catch (err) {
    error.value = extractError(err)
  } finally {
    loading.value = false
  }
}

function toggleEvent(key: string) {
  if (events.value.includes(key)) {
    events.value = events.value.filter(k => k !== key)
  } else {
    events.value = [...events.value, key]
  }
}

async function save() {
  error.value = ''
  if (!title.value || !targetUrl.value || (!secret.value && isNew.value) || events.value.length === 0) {
    error.value = t('admin.webhooks.form.validation')
    return
  }
  try {
    saving.value = true
    const payload: WebhookInput = {
      title: title.value.trim(),
      targetUrl: targetUrl.value.trim(),
      secret: secret.value.trim(),
      active: active.value,
      events: events.value,
      description: description.value.trim() || undefined,
    }
    if (isNew.value) {
      await createWebhook(payload)
    } else {
      // Keep existing secret when left blank during edit
      if (!payload.secret) delete (payload as any).secret
      await updateWebhook(id.value, payload)
    }
    router.push({ name: 'admin-webhooks' })
  } catch (err) {
    error.value = extractError(err)
  } finally {
    saving.value = false
  }
}

function goBack() { router.push({ name: 'admin-webhooks' }) }

onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-3xl px-4 py-10 sm:px-6 lg:px-8">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">{{ isNew ? t('admin.webhooks.new') : t('admin.webhooks.editTitle') }}</h1>
      <Button variant="outline" @click="goBack"><ArrowLeft :size="16" class="mr-2"/> {{ t('common.back') || 'Retour' }}</Button>
    </div>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Card class="clay-card">
      <CardHeader>
        <CardTitle>{{ t('admin.webhooks.form.title') }}</CardTitle>
        <CardDescription>{{ t('admin.webhooks.form.subtitle') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex items-center gap-3 py-10">
          <Loader2 :size="24" class="animate-spin" />
          <span>{{ t('admin.loading') }}</span>
        </div>
        <form v-else @submit.prevent="save" class="space-y-5">
          <div>
            <label class="block text-sm font-medium mb-2">{{ t('admin.webhooks.form.nameLabel') }}</label>
            <Input v-model="title" type="text" required :placeholder="t('admin.webhooks.form.namePlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ t('admin.webhooks.form.urlLabel') }}</label>
            <Input v-model="targetUrl" type="url" required placeholder="https://example.com/webhook" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ t('admin.webhooks.form.secretLabel') }}</label>
            <Input v-model="secret" :type="isNew ? 'text' : 'password'" :placeholder="isNew ? t('admin.webhooks.form.secretPlaceholder') : t('admin.webhooks.form.secretKeep')" :required="isNew" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ t('admin.webhooks.form.eventsLabel') }}</label>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <label v-for="e in availableWebhookEvents" :key="e.key" class="flex items-center gap-2">
                <input type="checkbox" :value="e.key" :checked="events.includes(e.key)" @change="toggleEvent(e.key)" />
                <span>{{ t(e.labelKey) }}</span>
              </label>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ t('admin.webhooks.form.descriptionLabel') }}</label>
            <Textarea v-model="description" :placeholder="t('admin.webhooks.form.descriptionPlaceholder')" />
          </div>

          <div class="pt-2">
            <Button type="submit" :disabled="saving">
              <Loader2 v-if="saving" :size="16" class="mr-2 animate-spin" />
              <Save v-else :size="16" class="mr-2" />
              {{ t('common.save') || 'Enregistrer' }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
