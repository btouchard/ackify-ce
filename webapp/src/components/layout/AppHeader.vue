<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Menu, X, ChevronDown, LogOut, FileText, Settings, Webhook, CheckSquare } from 'lucide-vue-next'
import ThemeToggle from './ThemeToggle.vue'
import LanguageSelect from './LanguageSelect.vue'
import AppLogo from '@/components/AppLogo.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const mobileMenuOpen = ref(false)
const userMenuOpen = ref(false)

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const user = computed(() => authStore.user)
const canCreateDocuments = computed(() => authStore.canCreateDocuments)

const getLocalPart = (email: string): string => email.split('@')[0] || email
const isEmail = (str: string): boolean => str.includes('@')
const splitIntoWords = (str: string): string[] => str.split(/[.\-_\s]+/).filter(p => p.length > 0)
const capitalize = (str: string): string => str ? str.charAt(0).toUpperCase() + str.slice(1).toLowerCase() : ''

const displayName = computed(() => {
  if (!user.value?.name && !user.value?.email) return ''
  if (user.value.name && !isEmail(user.value.name)) return user.value.name

  const localPart = getLocalPart(user.value.email || user.value.name || '')
  return splitIntoWords(localPart).map(capitalize).join(' ')
})

const userInitials = computed(() => {
  if (!user.value?.name && !user.value?.email) return '?'

  if (user.value.name && !isEmail(user.value.name)) {
    const words = splitIntoWords(user.value.name)
    const first = words[0]
    const second = words[1]
    if (words.length >= 2 && first && second) {
      return (first.charAt(0) + second.charAt(0)).toUpperCase()
    }
    return user.value.name.slice(0, 2).toUpperCase()
  }

  const localPart = getLocalPart(user.value.email || user.value.name || '')
  const words = splitIntoWords(localPart)
  const first = words[0]
  const second = words[1]

  if (words.length >= 2 && first && second) {
    return (first.charAt(0) + second.charAt(0)).toUpperCase()
  }

  return localPart.charAt(0).toUpperCase()
})

const isActive = (path: string) => {
  return route.path === path
}

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

const toggleUserMenu = () => {
  userMenuOpen.value = !userMenuOpen.value
}

const login = () => {
  router.push({ name: 'auth-choice' })
}

const logout = async () => {
  await authStore.logout()
  userMenuOpen.value = false
}

const closeMobileMenu = () => {
  mobileMenuOpen.value = false
}

const closeUserMenu = () => {
  userMenuOpen.value = false
}
</script>

<template>
  <header class="sticky top-0 z-50 w-full bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700">
    <nav class="mx-auto max-w-6xl px-4 sm:px-6" :aria-label="t('nav.mainNavigation')">
      <div class="flex h-16 items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center">
          <router-link to="/" class="focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 rounded-lg">
            <AppLogo size="md" :show-version="true" />
          </router-link>
        </div>

        <!-- Desktop Navigation -->
        <div class="hidden md:flex md:items-center md:space-x-1">
          <!-- My confirmations - authenticated only -->
          <router-link
            v-if="isAuthenticated"
            to="/signatures"
            :class="[
              'px-3 py-2 text-sm font-medium rounded-lg transition-colors',
              isActive('/signatures')
                ? 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-900/30'
                : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
            ]"
          >
            {{ t('nav.myConfirmations') }}
          </router-link>

          <!-- My documents - if can create -->
          <router-link
              v-if="canCreateDocuments"
              to="/documents"
              :class="[
              'px-3 py-2 text-sm font-medium rounded-lg transition-colors',
              isActive('/documents')
                ? 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-900/30'
                : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
            ]"
          >
            {{ t('nav.myDocuments') }}
          </router-link>
        </div>

        <!-- Right side: Language + Theme + Auth -->
        <div class="flex items-center space-x-2">
          <LanguageSelect />
          <ThemeToggle />

          <!-- Desktop Auth - User dropdown -->
          <div v-if="isAuthenticated" class="hidden md:block relative">
            <button
              @click="toggleUserMenu"
              class="flex items-center space-x-2 rounded-lg px-2 py-1.5 text-sm font-medium hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              aria-haspopup="true"
              :aria-expanded="userMenuOpen"
            >
              <!-- User avatar (picture or initials fallback) -->
              <img
                v-if="user?.picture"
                :src="user.picture"
                :alt="displayName"
                class="w-8 h-8 rounded-lg object-cover"
                referrerpolicy="no-referrer"
              />
              <div
                v-else
                class="w-8 h-8 rounded-lg bg-slate-100 dark:bg-slate-700 flex items-center justify-center text-xs font-semibold text-slate-600 dark:text-slate-300"
              >
                {{ userInitials }}
              </div>
              <span class="text-slate-700 dark:text-slate-200 hidden lg:inline">{{ displayName }}</span>
              <ChevronDown :size="16" class="text-slate-400" />
            </button>

            <!-- User dropdown menu -->
            <transition
              enter-active-class="transition ease-out duration-100"
              enter-from-class="transform opacity-0 scale-95"
              enter-to-class="transform opacity-100 scale-100"
              leave-active-class="transition ease-in duration-75"
              leave-from-class="transform opacity-100 scale-100"
              leave-to-class="transform opacity-0 scale-95"
            >
              <div
                v-if="userMenuOpen"
                @click.stop
                v-click-outside="closeUserMenu"
                class="absolute right-0 mt-2 w-56 origin-top-right bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-lg focus:outline-none"
                role="menu"
                aria-orientation="vertical"
              >
                <div class="p-2">
                  <!-- User info -->
                  <div class="px-3 py-2 border-b border-slate-100 dark:border-slate-700 mb-2">
                    <p class="font-medium text-slate-900 dark:text-slate-100">{{ displayName }}</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400 truncate">{{ user?.email }}</p>
                  </div>

                  <!-- Admin section - if admin -->
                  <template v-if="isAdmin">
                    <p class="px-3 py-1 text-xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                      {{ t('nav.administration') }}
                    </p>
                    <router-link
                      to="/admin"
                      @click="userMenuOpen = false"
                      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                      role="menuitem"
                    >
                      <FileText :size="16" />
                      <span>{{ t('nav.adminMenu.allDocuments') }}</span>
                    </router-link>
                    <router-link
                      to="/admin/settings"
                      @click="userMenuOpen = false"
                      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                      role="menuitem"
                    >
                      <Settings :size="16" />
                      <span>{{ t('nav.adminMenu.settings') }}</span>
                    </router-link>
                    <router-link
                      to="/admin/webhooks"
                      @click="userMenuOpen = false"
                      class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                      role="menuitem"
                    >
                      <Webhook :size="16" />
                      <span>{{ t('nav.adminMenu.webhooks') }}</span>
                    </router-link>
                    <div class="border-t border-slate-100 dark:border-slate-700 my-2"></div>
                  </template>

                  <!-- Logout -->
                  <button
                    @click="logout"
                    class="flex w-full items-center space-x-2 rounded-lg px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                    role="menuitem"
                  >
                    <LogOut :size="16" />
                    <span>{{ t('nav.logout') }}</span>
                  </button>
                </div>
              </div>
            </transition>
          </div>

          <!-- Login button (not authenticated) -->
          <button
            v-else
            @click="login"
            class="hidden md:inline-flex trust-gradient text-white font-medium rounded-lg px-4 py-2 text-sm hover:opacity-90 transition-opacity"
          >
            {{ t('nav.login') }}
          </button>

          <!-- Mobile menu button -->
          <button
            @click="toggleMobileMenu"
            class="md:hidden rounded-lg p-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            :aria-label="t('nav.mobileMenu')"
            :aria-expanded="mobileMenuOpen"
          >
            <Menu v-if="!mobileMenuOpen" :size="24" class="text-slate-600 dark:text-slate-300" />
            <X v-else :size="24" class="text-slate-600 dark:text-slate-300" />
          </button>
        </div>
      </div>
    </nav>

    <!-- Mobile menu -->
    <transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div v-if="mobileMenuOpen" class="md:hidden border-t border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900">
        <div class="space-y-1 px-4 pb-4 pt-2">
          <!-- Navigation links (authenticated) -->
          <template v-if="isAuthenticated">
            <!-- User info -->
            <div class="px-3 py-2 mb-2">
              <p class="font-medium text-slate-900 dark:text-slate-100">{{ displayName }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ user?.email }}</p>
            </div>

            <router-link
              to="/signatures"
              @click="closeMobileMenu"
              :class="[
                'flex items-center space-x-2 rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                isActive('/signatures')
                  ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                  : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
              ]"
            >
              <CheckSquare :size="18" />
              <span>{{ t('nav.myConfirmations') }}</span>
            </router-link>

            <router-link
              v-if="canCreateDocuments"
              to="/documents"
              @click="closeMobileMenu"
              :class="[
                'flex items-center space-x-2 rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                isActive('/documents') || route.path.startsWith('/documents/')
                  ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                  : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
              ]"
            >
              <FileText :size="18" />
              <span>{{ t('nav.myDocuments') }}</span>
            </router-link>

            <!-- Admin section -->
            <template v-if="isAdmin">
              <div class="border-t border-slate-200 dark:border-slate-700 pt-3 mt-3">
                <p class="px-3 py-1 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  {{ t('nav.administration') }}
                </p>
                <router-link
                  to="/admin"
                  @click="closeMobileMenu"
                  :class="[
                    'flex items-center space-x-2 rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                    isActive('/admin') && !route.path.startsWith('/admin/')
                      ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                      : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
                  ]"
                >
                  <FileText :size="18" />
                  <span>{{ t('nav.adminMenu.allDocuments') }}</span>
                </router-link>
                <router-link
                  to="/admin/settings"
                  @click="closeMobileMenu"
                  :class="[
                    'flex items-center space-x-2 rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                    route.path.startsWith('/admin/settings')
                      ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                      : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
                  ]"
                >
                  <Settings :size="18" />
                  <span>{{ t('nav.adminMenu.settings') }}</span>
                </router-link>
                <router-link
                  to="/admin/webhooks"
                  @click="closeMobileMenu"
                  :class="[
                    'flex items-center space-x-2 rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                    route.path.startsWith('/admin/webhooks')
                      ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                      : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
                  ]"
                >
                  <Webhook :size="18" />
                  <span>{{ t('nav.adminMenu.webhooks') }}</span>
                </router-link>
              </div>
            </template>

            <!-- Logout -->
            <div class="border-t border-slate-200 dark:border-slate-700 pt-3 mt-3">
              <button
                @click="logout"
                class="flex w-full items-center space-x-2 rounded-lg px-3 py-2.5 text-base font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
              >
                <LogOut :size="18" />
                <span>{{ t('nav.logout') }}</span>
              </button>
            </div>
          </template>

          <!-- Login button (not authenticated) -->
          <button
            v-else
            @click="login"
            class="w-full trust-gradient text-white font-medium rounded-lg px-4 py-3 text-base hover:opacity-90 transition-opacity"
          >
            {{ t('nav.login') }}
          </button>
        </div>
      </div>
    </transition>
  </header>
</template>
