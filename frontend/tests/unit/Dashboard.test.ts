import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import Dashboard from '../../src/views/Dashboard.vue'
import ProviderCard from '../../src/components/ProviderCard.vue'
import * as api from '../../src/services/api'

// Mock the API module
vi.mock('../../src/services/api', () => ({
  fetchProviders: vi.fn()
}))

describe('Dashboard', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should render loading state initially', () => {
    vi.mocked(api.fetchProviders).mockImplementation(() => new Promise(() => {}))

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    expect(wrapper.text()).toContain('Loading providers')
  })

  it('should render provider list after loading', async () => {
    const mockProviders = [
      {
        id: 'test-1',
        type: 'telegram',
        status: 'active',
        last_updated: '2025-01-01T00:00:00Z'
      },
      {
        id: 'test-2',
        type: 'email',
        status: 'error',
        last_updated: '2025-01-01T00:00:00Z',
        error_message: 'Connection failed'
      }
    ]

    vi.mocked(api.fetchProviders).mockResolvedValue(mockProviders)

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    const cards = wrapper.findAllComponents(ProviderCard)
    expect(cards).toHaveLength(2)
    expect(wrapper.text()).not.toContain('Loading providers')
  })

  it('should render empty state when no providers', async () => {
    vi.mocked(api.fetchProviders).mockResolvedValue([])

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('No providers configured')
  })

  it('should render error state on fetch failure', async () => {
    vi.mocked(api.fetchProviders).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Failed to load providers')
    expect(wrapper.text()).toContain('Network error')
  })

  it('should have retry button in error state', async () => {
    vi.mocked(api.fetchProviders).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    const retryButton = wrapper.find('button')
    expect(retryButton.exists()).toBe(true)
    expect(retryButton.text()).toContain('Retry')
  })

  it('should retry loading when retry button clicked', async () => {
    vi.mocked(api.fetchProviders)
      .mockRejectedValueOnce(new Error('Network error'))
      .mockResolvedValueOnce([])

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    // Should show error initially
    expect(wrapper.text()).toContain('Failed to load providers')

    // Click retry
    const retryButton = wrapper.find('button')
    await retryButton.trigger('click')
    await flushPromises()

    // Should now show empty state
    expect(wrapper.text()).toContain('No providers configured')
    expect(wrapper.text()).not.toContain('Failed to load providers')
  })

  it('should set up auto-refresh interval', async () => {
    vi.mocked(api.fetchProviders).mockResolvedValue([])

    mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    // Initial call
    expect(api.fetchProviders).toHaveBeenCalledTimes(1)

    // Fast-forward 30 seconds
    vi.advanceTimersByTime(30000)
    await flushPromises()

    // Should have called again
    expect(api.fetchProviders).toHaveBeenCalledTimes(2)

    // Fast-forward another 30 seconds
    vi.advanceTimersByTime(30000)
    await flushPromises()

    // Should have called a third time
    expect(api.fetchProviders).toHaveBeenCalledTimes(3)
  })

  it('should clear interval on unmount', async () => {
    vi.mocked(api.fetchProviders).mockResolvedValue([])

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    // Unmount component
    wrapper.unmount()

    // Fast-forward 30 seconds
    vi.advanceTimersByTime(30000)
    await flushPromises()

    // Should still only have the initial call, no more
    expect(api.fetchProviders).toHaveBeenCalledTimes(1)
  })

  it('should display last refreshed time', async () => {
    vi.mocked(api.fetchProviders).mockResolvedValue([])

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Last refreshed:')
    expect(wrapper.text()).toContain('Auto-refresh every 30 seconds')
  })

  it('should not show last refreshed during loading', () => {
    vi.mocked(api.fetchProviders).mockImplementation(() => new Promise(() => {}))

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    expect(wrapper.text()).not.toContain('Last refreshed:')
  })

  it('should not show last refreshed in error state', async () => {
    vi.mocked(api.fetchProviders).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(Dashboard, {
      global: {
        components: {
          ProviderCard
        }
      }
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('Last refreshed:')
  })
})
