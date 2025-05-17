package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"ocean-digital-twin/internal/database/models"
	"strconv"
	"time"
)

func (s *Server) GetCurrentsDataHandler(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	minLatStr := r.URL.Query().Get("min_lat")
	minLonStr := r.URL.Query().Get("min_lon")
	maxLatStr := r.URL.Query().Get("max_lat")
	maxLonStr := r.URL.Query().Get("max_lon")
	rawDataStr := r.URL.Query().Get("raw_data")

	endTime := time.Now().UTC()
	startTime := endTime.Add(-14 * 24 * time.Hour)
	rawData := false

	if startTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			startTime = parsedTime
		} else {
			slog.Error("Error parsing time", "time", startTimeStr, "err", err)
		}
	}
	if endTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			endTime = parsedTime
		} else {
			slog.Error("Error parsing time", "time", endTimeStr, "err", err)
		}
	}
	minLat := 40.50
	minLon := 1.10
	maxLat := 41.46
	maxLon := 2.83

	if minLatStr != "" {
		if val, err := strconv.ParseFloat(minLatStr, 64); err == nil {
			minLat = val
		}
	}
	if minLonStr != "" {
		if val, err := strconv.ParseFloat(minLonStr, 64); err == nil {
			minLon = val
		}
	}
	if maxLatStr != "" {
		if val, err := strconv.ParseFloat(maxLatStr, 64); err == nil {
			maxLat = val
		}
	}
	if maxLonStr != "" {
		if val, err := strconv.ParseFloat(maxLonStr, 64); err == nil {
			maxLon = val
		}
	}

	if rawDataStr != "" {
		val, err := strconv.ParseBool(rawDataStr)
		if err != nil {
			http.Error(w, "Error parsing raw data parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawData = val
	}

	data, err := s.db.GetCurrentsData(r.Context(), startTime, endTime, minLat, minLon, maxLat, maxLon, rawData)
	if err != nil {
		http.Error(w, "Error retrieving currents data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	geojsonData := models.CurrentsDataToGeoJSON(data)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(geojsonData)
	if err != nil {
		http.Error(w, "Error transforming chlorophyll data: "+err.Error(), http.StatusInternalServerError)
	}
}
