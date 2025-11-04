// SPDX-License-Identifier: AGPL-3.0-or-later
import { createI18n } from 'vue-i18n'

import en from './locales/en.json'
import fr from './locales/fr.json'
import es from './locales/es.json'
import de from './locales/de.json'
import it from './locales/it.json'

const getBrowserLocale = (): string | undefined => {
  const navigatorLocale =
    navigator.languages !== undefined
      ? navigator.languages[0]
      : navigator.language

  if (!navigatorLocale) {
    return undefined
  }

  const trimmedLocale = navigatorLocale.trim().split(/-|_/)[0]
  return trimmedLocale
}

const getInitialLocale = (): string => {
  const savedLocale = localStorage.getItem('locale')

  if (savedLocale && ['en', 'fr', 'es', 'de', 'it'].includes(savedLocale)) {
    return savedLocale
  }

  const browserLocale = getBrowserLocale()
  if (browserLocale && ['en', 'fr', 'es', 'de', 'it'].includes(browserLocale)) {
    return browserLocale
  }

  return 'en'
}

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    fr,
    es,
    de,
    it
  },
  datetimeFormats: {
    en: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    fr: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    es: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    de: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    it: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric'
      }
    }
  },
  numberFormats: {
    en: {
      currency: {
        style: 'currency',
        currency: 'USD'
      }
    },
    fr: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      }
    },
    es: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      }
    },
    de: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      }
    },
    it: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      }
    }
  }
})

export const setLocale = (locale: string) => {
  i18n.global.locale.value = locale as any
  document.documentElement.setAttribute('lang', locale)
  localStorage.setItem('locale', locale)
}

document.documentElement.setAttribute('lang', i18n.global.locale.value)
