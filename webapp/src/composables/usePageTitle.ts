// SPDX-License-Identifier: AGPL-3.0-or-later
import { watch } from 'vue'
import { useI18n } from 'vue-i18n'

/**
 * Composable to manage document.title with i18n support
 * @param titleKey - i18n key for the page title
 * @param params - optional parameters for i18n interpolation
 */
export function usePageTitle(titleKey?: string, params?: Record<string, any>) {
  const { t, locale } = useI18n()

  const updateTitle = () => {
    if (titleKey) {
      const translatedTitle = params ? t(titleKey, params) : t(titleKey)
      const appName = t('app.name')
      document.title = `${translatedTitle} - ${appName}`
    } else {
      document.title = t('app.name')
    }
  }

  // Update title on mount and when locale changes
  updateTitle()

  watch(locale, () => {
    updateTitle()
  })

  return {
    updateTitle
  }
}
