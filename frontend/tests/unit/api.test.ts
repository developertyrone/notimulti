import { describe, it, expect, beforeEach, vi } from 'vitest'
import { fetchProviders, fetchProviderDetail } from '../../src/services/api'

// Mock fetch globally
global.fetch = vi.fn()

describe('API Service', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  const baseUrl = window.location.origin

  describe('fetchProviders', () => {
    it('should fetch providers successfully', async () => {
      const mockData = {
        providers: [
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
        ],
        count: 2
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData
      })

      const result = await fetchProviders()

      expect(result).toEqual(mockData.providers)
      expect(global.fetch).toHaveBeenCalledWith(
        `${baseUrl}/api/v1/providers`,
        expect.objectContaining({
          method: 'GET',
          headers: { 'Content-Type': 'application/json' }
        })
      )
    })

    it('should throw error on HTTP error', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500
      })

      await expect(fetchProviders()).rejects.toThrow('HTTP error! status: 500')
    })

    it('should handle network error', async () => {
      ;(global.fetch as any).mockRejectedValueOnce(new Error('Network error'))

      await expect(fetchProviders()).rejects.toThrow('Network error')
    })

    it('should return empty array when no providers', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ providers: [], count: 0 })
      })

      const result = await fetchProviders()

      expect(result).toEqual([])
    })
  })

  describe('fetchProviderDetail', () => {
    it('should fetch provider detail successfully', async () => {
      const mockProvider = {
        id: 'test-provider',
        type: 'telegram',
        status: 'active',
        last_updated: '2025-01-01T00:00:00Z',
        config_checksum: 'abc123'
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockProvider
      })

      const result = await fetchProviderDetail('test-provider')

      expect(result).toEqual(mockProvider)
      expect(global.fetch).toHaveBeenCalledWith(
        `${baseUrl}/api/v1/providers/test-provider`,
        expect.objectContaining({
          method: 'GET',
          headers: { 'Content-Type': 'application/json' }
        })
      )
    })

    it('should throw error on 404', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 404
      })

      await expect(fetchProviderDetail('non-existent')).rejects.toThrow(
        'Provider not found: non-existent'
      )
    })

    it('should throw error on other HTTP errors', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500
      })

      await expect(fetchProviderDetail('test-provider')).rejects.toThrow(
        'HTTP error! status: 500'
      )
    })

    it('should handle network error', async () => {
      ;(global.fetch as any).mockRejectedValueOnce(new Error('Network error'))

      await expect(fetchProviderDetail('test-provider')).rejects.toThrow('Network error')
    })
  })
})
