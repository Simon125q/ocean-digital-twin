import type { CurrentsFeatureCollection } from '@/types/currents'

const API_BASE_URL = 'http://127.0.0.1:3000'

export async function fetchCurrentsData(
  raw_data: boolean = false,
): Promise<CurrentsFeatureCollection> {
  try {
    var query = ''
    if (raw_data) {
      query = '?raw_data=true'
    }
    const response = await fetch(`${API_BASE_URL}/currents${query}`)
    if (!response.ok) {
      throw new Error(`Failed to fetch currents data: ${response.status} ${response.status}`)
    }
    console.debug('response', response)
    const data: CurrentsFeatureCollection = await response.json()
    return data
  } catch (error) {
    console.error('Error fetching currents data:', error)
    return { type: 'FeatureCollection', features: [] }
  }
}
