<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Menu, X, ChevronDown, User, LogOut, Shield, FileSignature } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
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
  router.push({name: 'auth-choice'})
}

const logout = async () => {
  await authStore.logout()
  userMenuOpen.value = false
}

// Close mobile menu when clicking outside
const closeMobileMenu = () => {
  mobileMenuOpen.value = false
}

// Close user menu when clicking outside
const closeUserMenu = () => {
  userMenuOpen.value = false
}
</script>

<template>
  <header class="sticky top-0 z-50 w-full border-b border-border/40 clay-card backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <nav class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8" :aria-label="t('nav.mainNavigation')">
      <div class="flex h-16 items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center">
          <router-link to="/" class="focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 rounded-md">
            <AppLogo size="md" :show-version="true" />
          </router-link>
        </div>

        <!-- Desktop Navigation (only visible when authenticated) -->
        <div v-if="isAuthenticated" class="hidden md:flex md:items-center md:space-x-6">
          <router-link
            to="/"
            :class="[
              'text-sm font-medium transition-colors hover:text-primary',
              isActive('/') ? 'text-primary' : 'text-muted-foreground'
            ]"
          >
            {{ t('nav.home') }}
          </router-link>

          <router-link
            to="/signatures"
            :class="[
              'text-sm font-medium transition-colors hover:text-primary',
              isActive('/signatures') ? 'text-primary' : 'text-muted-foreground'
            ]"
          >
            {{ t('nav.myConfirmations') }}
          </router-link>

          <router-link
            v-if="isAdmin"
            to="/admin"
            :class="[
              'text-sm font-medium transition-colors hover:text-primary',
              isActive('/admin') ? 'text-primary' : 'text-muted-foreground'
            ]"
          >
            {{ t('nav.admin') }}
          </router-link>
        </div>

        <!-- Right side: Language + Theme toggle + Auth -->
        <div class="flex items-center space-x-2">
          <LanguageSelect />
          <ThemeToggle />

          <!-- Desktop Auth -->
          <div v-if="isAuthenticated" class="hidden md:block relative">
            <button
              @click="toggleUserMenu"
              class="flex items-center space-x-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
              aria-haspopup="true"
              :aria-expanded="userMenuOpen"
            >
              <User :size="18" />
              <span class="text-foreground">{{ user?.email?.split('@')[0] }}</span>
              <ChevronDown :size="16" class="text-muted-foreground" />
            </button>

            <!-- User dropdown -->
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
                class="absolute right-0 mt-2 w-56 origin-top-right clay-card rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                role="menu"
                aria-orientation="vertical"
              >
                <div class="p-2">
                  <div class="px-3 py-2 text-sm text-muted-foreground border-b border-border/40 mb-2">
                    <p class="font-medium text-foreground">{{ user?.name }}</p>
                    <p class="text-xs truncate">{{ user?.email }}</p>
                  </div>

                  <router-link
                    to="/signatures"
                    @click="userMenuOpen = false"
                    class="flex items-center space-x-2 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors"
                    role="menuitem"
                  >
                    <FileSignature :size="16" />
                    <span>{{ t('nav.myConfirmations') }}</span>
                  </router-link>

                  <router-link
                    v-if="isAdmin"
                    to="/admin"
                    @click="userMenuOpen = false"
                    class="flex items-center space-x-2 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors"
                    role="menuitem"
                  >
                    <Shield :size="16" />
                    <span>{{ t('nav.administration') }}</span>
                  </router-link>

                  <div class="border-t border-border/40 my-2"></div>

                  <button
                    @click="logout"
                    class="flex w-full items-center space-x-2 rounded-md px-3 py-2 text-sm text-destructive hover:bg-destructive/10 transition-colors"
                    role="menuitem"
                  >
                    <LogOut :size="16" />
                    <span>{{ t('nav.logout') }}</span>
                  </button>
                </div>
              </div>
            </transition>
          </div>

          <Button v-else @click="login" variant="default" size="sm" class="hidden md:inline-flex">
            {{ t('nav.login') }}
          </Button>

          <!-- Mobile menu button -->
          <button
            @click="toggleMobileMenu"
            class="md:hidden rounded-md p-2 hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
            :aria-label="t('nav.mobileMenu')"
            aria-expanded="false"
          >
            <Menu v-if="!mobileMenuOpen" :size="24" />
            <X v-else :size="24" />
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
      <div v-if="mobileMenuOpen" class="md:hidden border-t border-border/40">
        <div class="space-y-1 px-4 pb-3 pt-2">
          <!-- Navigation links (only when authenticated) -->
          <template v-if="isAuthenticated">
            <router-link
              to="/"
              @click="closeMobileMenu"
              :class="[
                'block rounded-md px-3 py-2 text-base font-medium transition-colors',
                isActive('/') ? 'bg-accent text-primary' : 'hover:bg-accent'
              ]"
            >
              {{ t('nav.home') }}
            </router-link>

            <router-link
              to="/signatures"
              @click="closeMobileMenu"
              :class="[
                'block rounded-md px-3 py-2 text-base font-medium transition-colors',
                isActive('/signatures') ? 'bg-accent text-primary' : 'hover:bg-accent'
              ]"
            >
              {{ t('nav.myConfirmations') }}
            </router-link>

            <router-link
              v-if="isAdmin"
              to="/admin"
              @click="closeMobileMenu"
              :class="[
                'block rounded-md px-3 py-2 text-base font-medium transition-colors',
                isActive('/admin') ? 'bg-accent text-primary' : 'hover:bg-accent'
              ]"
            >
              {{ t('nav.administration') }}
            </router-link>

            <div class="border-t border-border/40 pt-3 mt-3">
              <div class="px-3 py-2 text-sm text-muted-foreground mb-2">
                <p class="font-medium text-foreground">{{ user?.name }}</p>
                <p class="text-xs">{{ user?.email }}</p>
              </div>
              <button
                @click="logout"
                class="w-full text-left rounded-md px-3 py-2 text-base font-medium text-destructive hover:bg-destructive/10 transition-colors"
              >
                {{ t('nav.logout') }}
              </button>
            </div>
          </template>

          <!-- Login button (when not authenticated) -->
          <Button v-else @click="login" variant="default" class="w-full">
            {{ t('nav.login') }}
          </Button>
        </div>
      </div>
    </transition>
  </header>
</template>

<style scoped>
/* Click outside directive will be added via composable */
</style>
