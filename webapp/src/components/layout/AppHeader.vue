<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Menu, X, ChevronDown, LogOut, Shield } from 'lucide-vue-next'
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

// User initials for avatar
const userInitials = computed(() => {
  if (!user.value?.name && !user.value?.email) return '?'
  const name = user.value.name || user.value.email || ''
  const parts = name.split(/[\s@]+/).filter(p => p.length > 0)
  if (parts.length >= 2) {
    const first = parts[0] ?? ''
    const second = parts[1] ?? ''
    if (first.length > 0 && second.length > 0) {
      return (first.charAt(0) + second.charAt(0)).toUpperCase()
    }
  }
  return name.slice(0, 2).toUpperCase()
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
        <div v-if="isAuthenticated" class="hidden md:flex md:items-center md:space-x-1">
          <router-link
            to="/"
            :class="[
              'px-3 py-2 text-sm font-medium rounded-lg transition-colors',
              isActive('/')
                ? 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-900/30'
                : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
            ]"
          >
            {{ t('nav.home') }}
          </router-link>

          <router-link
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
              <!-- User avatar with initials -->
              <div class="w-8 h-8 rounded-lg bg-slate-100 dark:bg-slate-700 flex items-center justify-center text-xs font-semibold text-slate-600 dark:text-slate-300">
                {{ userInitials }}
              </div>
              <span class="text-slate-700 dark:text-slate-200 hidden lg:inline">{{ user?.name || user?.email?.split('@')[0] }}</span>
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
                    <p class="font-medium text-slate-900 dark:text-slate-100">{{ user?.name }}</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400 truncate">{{ user?.email }}</p>
                  </div>

                  <!-- Menu items -->
                  <router-link
                    v-if="isAdmin"
                    to="/admin"
                    @click="userMenuOpen = false"
                    class="flex items-center space-x-2 rounded-lg px-3 py-2 text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                    role="menuitem"
                  >
                    <Shield :size="16" />
                    <span>{{ t('nav.administration') }}</span>
                  </router-link>

                  <div v-if="isAdmin" class="border-t border-slate-100 dark:border-slate-700 my-2"></div>

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
            <router-link
              to="/"
              @click="closeMobileMenu"
              :class="[
                'block rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                isActive('/')
                  ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                  : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
              ]"
            >
              {{ t('nav.home') }}
            </router-link>

            <router-link
              to="/signatures"
              @click="closeMobileMenu"
              :class="[
                'block rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                isActive('/signatures')
                  ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                  : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
              ]"
            >
              {{ t('nav.myConfirmations') }}
            </router-link>

            <router-link
              v-if="isAdmin"
              to="/admin"
              @click="closeMobileMenu"
              :class="[
                'block rounded-lg px-3 py-2.5 text-base font-medium transition-colors',
                isActive('/admin') || route.path.startsWith('/admin')
                  ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                  : 'text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
              ]"
            >
              {{ t('nav.administration') }}
            </router-link>

            <!-- User section -->
            <div class="border-t border-slate-200 dark:border-slate-700 pt-3 mt-3">
              <div class="px-3 py-2 mb-2">
                <p class="font-medium text-slate-900 dark:text-slate-100">{{ user?.name }}</p>
                <p class="text-xs text-slate-500 dark:text-slate-400">{{ user?.email }}</p>
              </div>
              <button
                @click="logout"
                class="w-full text-left rounded-lg px-3 py-2.5 text-base font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
              >
                {{ t('nav.logout') }}
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
