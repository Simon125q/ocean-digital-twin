package models

import (
	"math"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type CurrentsData struct {
	ID              int       `json:"id"`
	MeasurementTime time.Time `json:"measurement_time"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	//surface geostrophic eastward sea water velocity in m/s
	UCurrent float32 `json:"u_current"`
	//surface geostrophic northward sea water velocity in m/s
	VCurrent  float32   `json:"v_current"`
	CreatedAt time.Time `json:"created_at"`
}

type VCurrentsData struct {
	ID              int       `json:"id"`
	MeasurementTime time.Time `json:"measurement_time"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	VCurrent        float32   `json:"v_current"`
	CreatedAt       time.Time `json:"created_at"`
}

type UCurrentsData struct {
	ID              int       `json:"id"`
	MeasurementTime time.Time `json:"measurement_time"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	UCurrent        float32   `json:"u_current"`
	CreatedAt       time.Time `json:"created_at"`
}

func (c *UCurrentsData) Value() float32 {
	return c.UCurrent
}

func (c *UCurrentsData) SetValue(val float32) {
	c.UCurrent = val
}

func (c *VCurrentsData) Value() float32 {
	return c.VCurrent
}

func (c *VCurrentsData) SetValue(val float32) {
	c.VCurrent = val
}

func CurrentsDataToGeoJSON(data []CurrentsData) *geojson.FeatureCollection {
	fc := geojson.NewFeatureCollection()

	for _, d := range data {
		point := orb.Point{d.Longitude, d.Latitude}
		feature := geojson.NewFeature(point)
		if !math.IsNaN(float64(d.UCurrent)) && (!math.IsNaN(float64(d.VCurrent))) {
			feature.Properties = map[string]interface{}{
				"id":               d.ID,
				"measurement_time": d.MeasurementTime,
				"u_current":        d.UCurrent,
				"v_current":        d.VCurrent,
				"current_angle":    calculateCurrentAngle(d.UCurrent, d.VCurrent),
				"magnitude":        calculateMagnitude(d.UCurrent, d.VCurrent),
			}
			fc.Append(feature)
		}
	}
	return fc
}

func calculateCurrentAngle(u, v float32) float32 {
	if u == 0 && v == 0 {
		return 0.0
	}

	if u == 0 {
		if v > 0 {
			return 0.0 // North
		}
		return 180.0 // South
	}

	if v == 0 {
		if u > 0 {
			return 90.0 // East
		}
		return 270.0 // West
	}

	radians := math.Atan2(float64(v), float64(u))

	angleFromEastDegrees := radians * (180.0 / math.Pi)

	// Adjust to Mapbox's system where 0° is North and angles increase clockwise.
	mapboxAngle := 90.0 - angleFromEastDegrees

	normalizedAngle := math.Mod(mapboxAngle, 360.0)

	if normalizedAngle < 0 {
		normalizedAngle += 360.0
	}

	return float32(normalizedAngle)
}

func calculateMagnitude(u, v float32) float32 {
	return float32(math.Sqrt(math.Pow(float64(u), 2) + math.Pow(float64(v), 2)))
}
