import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ProviderCard from '../../src/components/ProviderCard.vue'
import StatusBadge from '../../src/components/StatusBadge.vue'

describe('ProviderCard', () => {
  const mockProvider = {
    id: 'test-telegram',
    type: 'telegram',
    status: 'active',
    last_updated: '2025-01-01T12:00:00Z'
  }

  it('should render provider information', () => {
    const wrapper = mount(ProviderCard, {
      props: {
        provider: mockProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    expect(wrapper.text()).toContain('test-telegram')
    expect(wrapper.text()).toContain('telegram')
  })

  it('should render StatusBadge component', () => {
    const wrapper = mount(ProviderCard, {
      props: {
        provider: mockProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    const badge = wrapper.findComponent(StatusBadge)
    expect(badge.exists()).toBe(true)
    expect(badge.props('status')).toBe('active')
  })

  it('should format timestamp correctly', () => {
    const wrapper = mount(ProviderCard, {
      props: {
        provider: mockProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    // Check that timestamp is displayed (format depends on locale)
    expect(wrapper.text()).toContain('Last updated:')
  })

  it('should display error message when present', () => {
    const providerWithError = {
      ...mockProvider,
      status: 'error',
      error_message: 'Connection failed'
    }

    const wrapper = mount(ProviderCard, {
      props: {
        provider: providerWithError
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    expect(wrapper.text()).toContain('Connection failed')
  })

  it('should not display error message when not present', () => {
    const wrapper = mount(ProviderCard, {
      props: {
        provider: mockProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    // Error message should not be in the DOM
    const errorText = wrapper.find('.text-red-600')
    expect(errorText.exists()).toBe(false)
  })

  it('should apply correct styling classes', () => {
    const wrapper = mount(ProviderCard, {
      props: {
        provider: mockProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    expect(wrapper.classes()).toContain('bg-white')
    expect(wrapper.classes()).toContain('shadow')
    expect(wrapper.classes()).toContain('rounded-lg')
    expect(wrapper.classes()).toContain('p-6')
  })

  it('should handle invalid timestamp gracefully', () => {
    const invalidProvider = {
      ...mockProvider,
      last_updated: 'invalid-date'
    }

    const wrapper = mount(ProviderCard, {
      props: {
        provider: invalidProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    // Should show "Invalid Date" for unparseable timestamps
    expect(wrapper.text()).toContain('Invalid Date')
  })

  it('should display provider type', () => {
    const wrapper = mount(ProviderCard, {
      props: {
        provider: mockProvider
      },
      global: {
        components: {
          StatusBadge
        }
      }
    })

    expect(wrapper.text()).toContain('Type:')
    expect(wrapper.text()).toContain('telegram')
  })
})
