import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import StatusBadge from '../../src/components/StatusBadge.vue'

describe('StatusBadge', () => {
  it('should render active status with green styling', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'active'
      }
    })

    expect(wrapper.text()).toBe('Active')
    expect(wrapper.classes()).toContain('bg-green-100')
    expect(wrapper.classes()).toContain('text-green-800')
    expect(wrapper.attributes('aria-label')).toBe('Status: active')
  })

  it('should render error status with red styling', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'error'
      }
    })

    expect(wrapper.text()).toBe('Error')
    expect(wrapper.classes()).toContain('bg-red-100')
    expect(wrapper.classes()).toContain('text-red-800')
  })

  it('should render disabled status with gray styling', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'disabled'
      }
    })

    expect(wrapper.text()).toBe('Disabled')
    expect(wrapper.classes()).toContain('bg-gray-100')
    expect(wrapper.classes()).toContain('text-gray-800')
  })

  it('should render initializing status with yellow styling', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'initializing'
      }
    })

    expect(wrapper.text()).toBe('Initializing')
    expect(wrapper.classes()).toContain('bg-yellow-100')
    expect(wrapper.classes()).toContain('text-yellow-800')
  })

  it('should handle unknown status with gray styling', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'unknown'
      }
    })

    expect(wrapper.text()).toBe('Unknown')
    expect(wrapper.classes()).toContain('bg-gray-100')
    expect(wrapper.classes()).toContain('text-gray-800')
  })

  it('should have correct aria-label', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'active'
      }
    })

    expect(wrapper.attributes('aria-label')).toBe('Status: active')
    expect(wrapper.attributes('role')).toBe('status')
  })

  it('should apply Tailwind classes correctly', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'active'
      }
    })

    expect(wrapper.classes()).toContain('inline-flex')
    expect(wrapper.classes()).toContain('items-center')
    expect(wrapper.classes()).toContain('px-2.5')
    expect(wrapper.classes()).toContain('py-0.5')
    expect(wrapper.classes()).toContain('rounded-full')
    expect(wrapper.classes()).toContain('text-xs')
    expect(wrapper.classes()).toContain('font-medium')
  })

  it('should capitalize status text', () => {
    const wrapper = mount(StatusBadge, {
      props: {
        status: 'active'
      }
    })

    expect(wrapper.text()).toBe('Active')
    expect(wrapper.text()).not.toBe('active')
  })
})
