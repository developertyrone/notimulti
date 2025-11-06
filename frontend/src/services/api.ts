// API service for communication with notification server backend

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export interface Provider {
  id: string
  type: string
  status: string
  last_updated: string
  error_message?: string
  config_checksum?: string
}

export interface ProvidersResponse {
  providers: Provider[]
  count: number
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
