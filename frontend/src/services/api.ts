// API service for communication with notification server backend

const API_BASE_URL = import.meta.env.VITE_API_URL || window.location.origin

export interface Provider {
  id: string
  type: string
  status: string
  last_updated: string
  error_message?: string
  config_checksum?: string
  last_test_at?: string
  last_test_status?: string
}

export interface ProvidersResponse {
  providers: Provider[]
  count: number
}

export interface NotificationLogEntry {
  id: number
  provider_id: string
  provider_type: string
  recipient: string
  message: string
  subject?: string
  metadata?: any
  priority: string
  status: string
  error_message?: string
  attempts: number
  created_at: string
  delivered_at?: string
  is_test: boolean
}

export interface NotificationHistoryResponse {
  notifications: NotificationLogEntry[]
  pagination: {
    page_size: number
    has_more: boolean
    next_cursor?: number
  }
}

export interface HistoryFilters {
  provider_id?: string
  provider_type?: string
  status?: string
  date_from?: string
  date_to?: string
  include_tests?: boolean
  cursor?: number
  page_size?: number
}

/**
 * Fetches the list of all providers from the backend
 * @returns Promise resolving to providers array
 * @throws Error if the request fails
 */
export async function fetchProviders(): Promise<Provider[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/providers`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data: ProvidersResponse = await response.json()
    return data.providers || []
  } catch (error) {
    console.error('Failed to fetch providers:', error)
    throw error
  }
}

/**
 * Fetches detailed information for a specific provider
 * @param id - The provider ID
 * @returns Promise resolving to provider details
 * @throws Error if the request fails or provider not found (404)
 */
export async function fetchProviderDetail(id: string): Promise<Provider> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/providers/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Provider not found: ${id}`)
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data: Provider = await response.json()
    return data
  } catch (error) {
    console.error(`Failed to fetch provider detail for ${id}:`, error)
    throw error
  }
}

/**
 * Fetches notification history with optional filters and pagination
 * @param filters - Query filters for notification history
 * @returns Promise resolving to notification history with pagination
 * @throws Error if the request fails
 */
export async function getNotificationHistory(filters?: HistoryFilters): Promise<NotificationHistoryResponse> {
  try {
    const params = new URLSearchParams()
    if (filters?.provider_id) params.append('provider_id', filters.provider_id)
    if (filters?.provider_type) params.append('provider_type', filters.provider_type)
    if (filters?.status) params.append('status', filters.status)
    if (filters?.date_from) params.append('date_from', filters.date_from)
    if (filters?.date_to) params.append('date_to', filters.date_to)
    if (filters?.include_tests !== undefined) params.append('include_tests', String(filters.include_tests))
    if (filters?.cursor) params.append('cursor', String(filters.cursor))
    if (filters?.page_size) params.append('page_size', String(filters.page_size))

    const response = await fetch(`${API_BASE_URL}/api/v1/notifications/history?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    return await response.json()
  } catch (error) {
    console.error('Failed to fetch notification history:', error)
    throw error
  }
}

/**
 * Fetches detailed information for a specific notification
 * @param id - The notification log entry ID
 * @returns Promise resolving to notification details
 * @throws Error if the request fails or notification not found (404)
 */
export async function getNotificationDetail(id: number): Promise<NotificationLogEntry> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/notifications/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Notification not found: ${id}`)
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    return await response.json()
  } catch (error) {
    console.error(`Failed to fetch notification detail for ${id}:`, error)
    throw error
  }
}

/**
 * Tests a provider configuration by sending a test notification
 * @param providerId - The provider ID to test
 * @returns Promise resolving to test result
 * @throws Error if the request fails
 */
export async function testProvider(providerId: string): Promise<any> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/providers/${providerId}/test`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    return await response.json()
  } catch (error) {
    console.error(`Failed to test provider ${providerId}:`, error)
    throw error
  }
}
