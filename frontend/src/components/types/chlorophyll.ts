import { GeoJSON } from 'geojson'

export interface ChlorophyllFeatureProperties {
  id: number
  measurement_time: string
  chlor_a: number
}

export type ChlorophyllGeoJSONFeature = GeoJSON.Feature<GeoJSON.Point, ChlorophyllFeatureProperties>

export type ChlorophyllFeatureCollection = GeoJSON.FeatureCollection<
  GeoJSON.Point,
  ChlorophyllFeatureProperties
>
