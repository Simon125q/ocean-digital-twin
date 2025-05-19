import { GeoJSON } from 'geojson'

export interface CurrentsFeatureProperties {
  id: number
  measurement_time: string
  v_current: number
  u_current: number
}

export type CurrentsGeoJSONFeature = GeoJSON.Feature<GeoJSON.Point, CurrentsFeatureProperties>

export type CurrentsFeatureCollection = GeoJSON.FeatureCollection<
  GeoJSON.Point,
  CurrentsFeatureProperties
>
