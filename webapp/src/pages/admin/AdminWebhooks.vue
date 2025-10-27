<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { listWebhooks, deleteWebhook, toggleWebhook, type Webhook } from '@/services/webhooks'
import { extractError } from '@/services/http'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import Button from '@/components/ui/Button.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'
import { Loader2, Plus, Pencil, Trash2, ToggleLeft, ToggleRight, BadgeCheck } from 'lucide-vue-next'

const router = useRouter()
const { t } = useI18n()

const loading = ref(true)
const error = ref('')
const items = ref<Webhook[]>([])
const deleting = ref<number | null>(null)
const toggling = ref<number | null>(null)

async function load() {
  try {
    loading.value = true
    error.value = ''
    const resp = await listWebhooks()
    items.value = resp.data
  } catch (err) {
    error.value = extractError(err)
  } finally {
    loading.value = false
  }
}

function gotoNew() { router.push({ name: 'admin-webhook-new' }) }
function gotoEdit(id: number) { router.push({ name: 'admin-webhook-edit', params: { id } }) }

async function onDelete(id: number) {
  if (!confirm(t('admin.webhooks.confirmDelete'))) return
  try {
    deleting.value = id
    await deleteWebhook(id)
    await load()
  } catch (err) {
    error.value = extractError(err)
  } finally {
    deleting.value = null
  }
}

async function onToggle(id: number, enable: boolean) {
  try {
    toggling.value = id
    await toggleWebhook(id, enable)
    await load()
  } catch (err) {
    error.value = extractError(err)
  } finally {
    toggling.value = null
  }
}

function formatEvents(evts: string[]): string[] {
  return evts.map(e => t(`admin.webhooks.eventsMap.${e}`, e))
}

onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">{{ t('admin.webhooks.title') }}</h1>
        <p class="text-muted-foreground">{{ t('admin.webhooks.subtitle') }}</p>
      </div>
      <Button @click="gotoNew">
        <Plus :size="16" class="mr-2" />
        {{ t('admin.webhooks.new') }}
      </Button>
    </div>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Card class="clay-card">
      <CardHeader>
        <CardTitle>{{ t('admin.webhooks.listTitle') }}</CardTitle>
        <CardDescription>{{ t('admin.webhooks.listSubtitle') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex items-center gap-3 py-10">
          <Loader2 :size="24" class="animate-spin" />
          <span>{{ t('admin.loading') }}</span>
        </div>

        <div v-else>
          <div v-if="items.length > 0" class="rounded-md border border-border/40 overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{{ t('admin.webhooks.columns.title') }}</TableHead>
                  <TableHead>{{ t('admin.webhooks.columns.url') }}</TableHead>
                  <TableHead>{{ t('admin.webhooks.columns.events') }}</TableHead>
                  <TableHead>{{ t('admin.webhooks.columns.status') }}</TableHead>
                  <TableHead class="text-right">{{ t('admin.webhooks.columns.actions') }}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="wh in items" :key="wh.id">
                  <TableCell>
                    <div class="font-medium">{{ wh.title || '-' }}</div>
                    <div v-if="wh.description" class="text-xs text-muted-foreground">{{ wh.description }}</div>
                  </TableCell>
                  <TableCell>
                    <a :href="wh.targetUrl" target="_blank" class="text-primary hover:underline">{{ wh.targetUrl }}</a>
                  </TableCell>
                  <TableCell>
                    <div class="flex flex-wrap gap-1">
                      <span v-for="e in formatEvents(wh.events)" :key="e" class="px-2 py-0.5 text-xs rounded bg-muted">{{ e }}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span v-if="wh.active" class="inline-flex items-center text-green-600"><BadgeCheck :size="16" class="mr-1"/>{{ t('admin.webhooks.status.enabled') }}</span>
                    <span v-else class="inline-flex items-center text-muted-foreground">{{ t('admin.webhooks.status.disabled') }}</span>
                  </TableCell>
                  <TableCell class="text-right">
                    <div class="flex items-center justify-end gap-2">
                      <Button variant="outline" size="sm" @click="gotoEdit(wh.id)">
                        <Pencil :size="14" class="mr-1" /> {{ t('admin.webhooks.edit') }}
                      </Button>
                      <Button variant="outline" size="sm" @click="onToggle(wh.id, !wh.active)" :disabled="toggling===wh.id">
                        <Loader2 v-if="toggling===wh.id" :size="14" class="mr-1 animate-spin" />
                        <ToggleRight v-else-if="!wh.active" :size="14" class="mr-1" />
                        <ToggleLeft v-else :size="14" class="mr-1" />
                        {{ wh.active ? t('admin.webhooks.disable') : t('admin.webhooks.enable') }}
                      </Button>
                      <Button variant="destructive" size="sm" @click="onDelete(wh.id)" :disabled="deleting===wh.id">
                        <Loader2 v-if="deleting===wh.id" :size="14" class="mr-1 animate-spin" />
                        <Trash2 v-else :size="14" class="mr-1" />
                        {{ t('admin.webhooks.delete') }}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <div v-else class="py-10 text-center text-muted-foreground">{{ t('admin.webhooks.empty') }}</div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

