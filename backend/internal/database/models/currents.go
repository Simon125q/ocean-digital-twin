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
			}
			fc.Append(feature)
		}
	}
	return fc
}
