// SPDX-License-Identifier: AGPL-3.0-or-later
import { onMounted, onUnmounted, type Ref } from 'vue'

export function useClickOutside(elementRef: Ref<HTMLElement | null>, callback: () => void) {
  const handleClickOutside = (event: MouseEvent) => {
    if (elementRef.value && !elementRef.value.contains(event.target as Node)) {
      callback()
    }
  }

  onMounted(() => {
    document.addEventListener('click', handleClickOutside)
  })

  onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside)
  })
}

// Extended HTMLElement type for directive
interface HTMLElementWithClickOutside extends HTMLElement {
  clickOutsideEvent?: (event: Event) => void
}

// Directive version
export const vClickOutside = {
  mounted(el: HTMLElement, binding: { value: () => void }) {
    const element = el as HTMLElementWithClickOutside

    element.clickOutsideEvent = (event: Event) => {
      if (!(el === event.target || el.contains(event.target as Node))) {
        binding.value()
      }
    }

    // Add a small delay to avoid immediate trigger from the click that opened the element
    setTimeout(() => {
      document.addEventListener('click', element.clickOutsideEvent!)
    }, 50)
  },
  unmounted(el: HTMLElement) {
    const element = el as HTMLElementWithClickOutside
    if (element.clickOutsideEvent) {
      document.removeEventListener('click', element.clickOutsideEvent)
    }
  }
}
