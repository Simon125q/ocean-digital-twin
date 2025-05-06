package models

import "time"

type ChlorophyllData struct {
	ID              int       `json:"id"`
	MeasurementTime time.Time `json:"measurement_time"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	ChlorophyllA    float32   `json:"chlor_a"`
	CreatedAt       time.Time `json:"created_at"`
}
