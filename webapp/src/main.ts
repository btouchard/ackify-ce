// SPDX-License-Identifier: AGPL-3.0-or-later
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import { i18n } from './i18n'
import './style.css'
import App from './App.vue'
import { vClickOutside } from './composables/useClickOutside'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(i18n)
app.use(router)

app.directive('click-outside', vClickOutside)

app.mount('#app')
