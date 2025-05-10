package models

import (
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type ChlorophyllData struct {
	ID              int       `json:"id"`
	MeasurementTime time.Time `json:"measurement_time"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	ChlorophyllA    float32   `json:"chlor_a"`
	CreatedAt       time.Time `json:"created_at"`
}

func (c *ChlorophyllData) Value() float32 {
	return c.ChlorophyllA
}

func (c *ChlorophyllData) SetValue(val float32) {
	c.ChlorophyllA = val
}

func ToGeoJSON(data []ChlorophyllData) *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()

	for _, d := range data {
		point := orb.Point{d.Longitude, d.Latitude}
		feature := geojson.NewFeature(point)
		feature.Properties = map[string]interface{}{
			"id":               d.ID,
			"measurement_time": d.MeasurementTime,
			"chlor_a":          d.ChlorophyllA,
		}
		fc.Append(feature)
	}
	return fc
}
