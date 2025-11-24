// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import NotificationToast from '@/components/NotificationToast.vue'
import { useUIStore } from '@/stores/ui'

describe('NotificationToast Component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  const mountComponent = () => {
    return mount(NotificationToast, {
      global: {
        stubs: {
          'transition-group': false
        }
      }
    })
  }

  describe('Business logic - Notification display by type', () => {
    it('should render success notification with correct styling', () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('Operation successful', 'The operation completed')

      const wrapper = mountComponent()

      expect(wrapper.text()).toContain('Operation successful')
      expect(wrapper.text()).toContain('The operation completed')
      expect(wrapper.find('.text-green-500').exists()).toBe(true)
    })

    it('should render error notification with correct styling', () => {
      const uiStore = useUIStore()
      uiStore.showError('Operation failed', 'An error occurred')

      const wrapper = mountComponent()

      expect(wrapper.text()).toContain('Operation failed')
      expect(wrapper.text()).toContain('An error occurred')
      expect(wrapper.find('.text-destructive').exists()).toBe(true)
    })

    it('should render warning notification with correct styling', () => {
      const uiStore = useUIStore()
      uiStore.showWarning('Warning', 'Please be careful')

      const wrapper = mountComponent()

      expect(wrapper.text()).toContain('Warning')
      expect(wrapper.text()).toContain('Please be careful')
      expect(wrapper.find('.text-yellow-500').exists()).toBe(true)
    })

    it('should render info notification with correct styling', () => {
      const uiStore = useUIStore()
      uiStore.showInfo('Information', 'Here is some info')

      const wrapper = mountComponent()

      expect(wrapper.text()).toContain('Information')
      expect(wrapper.text()).toContain('Here is some info')
      expect(wrapper.find('.text-primary').exists()).toBe(true)
    })
  })

  describe('Business logic - Multiple notifications', () => {
    it('should display multiple notifications at once', () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('First notification')
      uiStore.showError('Second notification')
      uiStore.showWarning('Third notification')

      const wrapper = mountComponent()

      expect(wrapper.findAll('.bg-card')).toHaveLength(3)
      expect(wrapper.text()).toContain('First notification')
      expect(wrapper.text()).toContain('Second notification')
      expect(wrapper.text()).toContain('Third notification')
    })

    it('should allow notifications with only title (no message)', () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('Title only')

      const wrapper = mountComponent()

      const notifications = wrapper.findAll('.bg-card')
      expect(notifications).toHaveLength(1)
      expect(wrapper.text()).toContain('Title only')
    })
  })

  describe('Business logic - Notification removal', () => {
    it('should close notification when close button is clicked', async () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('Test notification')

      const wrapper = mountComponent()

      expect(wrapper.findAll('.bg-card')).toHaveLength(1)

      const closeButton = wrapper.find('button')
      await closeButton.trigger('click')

      expect(uiStore.notifications).toHaveLength(0)
    })
  })

  describe('Accessibility', () => {
    it('should have accessible close button with aria-label', () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('Test notification')

      const wrapper = mountComponent()

      const closeButton = wrapper.find('button')
      expect(closeButton.attributes('aria-label')).toBe('Fermer la notification')
    })

    it('should have sr-only text for screen readers', () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('Test notification')

      const wrapper = mountComponent()

      expect(wrapper.find('.sr-only').text()).toBe('Close')
    })
  })

  describe('Visual consistency', () => {
    it('should apply consistent card styling across notification types', () => {
      const uiStore = useUIStore()
      uiStore.showSuccess('Success')
      uiStore.showError('Error')
      uiStore.showWarning('Warning')
      uiStore.showInfo('Info')

      const wrapper = mountComponent()

      const cards = wrapper.findAll('.bg-card')
      expect(cards).toHaveLength(4)

      // All cards should have same base classes
      cards.forEach(card => {
        expect(card.classes()).toContain('bg-card')
        expect(card.classes()).toContain('text-card-foreground')
        expect(card.classes()).toContain('shadow-lg')
        expect(card.classes()).toContain('rounded-lg')
      })
    })
  })
})
