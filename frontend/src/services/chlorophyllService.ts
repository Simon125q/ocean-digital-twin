import type { ChlorophyllFeatureCollection } from '@/types/chlorophyll'

const API_BASE_URL = 'http://127.0.0.1:3000'

export async function fetchChlorophyllData(): Promise<ChlorophyllFeatureCollection> {
  try {
    const response = await fetch(`${API_BASE_URL}/chlorophyll`)
    if (!response.ok) {
      throw new Error(`Failed to fetch chlorophyll data: ${response.status} ${response.status}`)
    }
    console.debug('response', response)
    const data: ChlorophyllFeatureCollection = await response.json()
    return data
  } catch (error) {
    console.error('Error fetching chlorophyll data:', error)
    return { type: 'FeatureCollection', features: [] }
  }
}
