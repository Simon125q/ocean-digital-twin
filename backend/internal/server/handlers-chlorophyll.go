package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ocean-digital-twin/internal/database/models"
	"ocean-digital-twin/internal/utils/erddap"
)

func (s *Server) GetChlorophyllDataHandler(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	minLatStr := r.URL.Query().Get("min_lat")
	minLonStr := r.URL.Query().Get("min_lon")
	maxLatStr := r.URL.Query().Get("max_lat")
	maxLonStr := r.URL.Query().Get("max_lon")

	endTime := time.Now().UTC()
	startTime := endTime.Add(-7 * 24 * time.Hour)

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
	minLat := 40.83
	minLon := 1.10
	maxLat := 41.26
	maxLon := 2.53

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

	data, err := s.db.GetChlorophyllData(r.Context(), startTime, endTime, minLat, minLon, maxLat, maxLon)
	if err != nil {
		http.Error(w, "Error retrieving chlorophyll data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	geojsonData := models.ToGeoJSON(data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(geojsonData)
}

func (s *Server) SaveDataTest(w http.ResponseWriter, r *http.Request) {
	minLat := 40.83
	minLon := 1.10
	maxLat := 41.26
	maxLon := 2.53
	endTime := time.Now().UTC()
	startTime := endTime.Add(-7 * 24 * time.Hour)
	logger := slog.Default()
	downloader := erddap.NewDownloader(logger, minLat, minLon, maxLat, maxLon)
	data, err := downloader.DownloadChlorophyllData(context.Background(), startTime, endTime)
	if err != nil {
		slog.Error("error downloading chlor data", "err", err)
	}
	fmt.Println(len(data))
	s.db.SaveChlorophyllData(context.Background(), data)
}

func (s *Server) GetDataTest(w http.ResponseWriter, r *http.Request) {
	minLat := 40.83
	minLon := 1.10
	maxLat := 41.26
	maxLon := 2.53
	endTime := time.Now().UTC()
	startTime := endTime.Add(-10 * 24 * time.Hour)
	data, err := s.db.GetChlorophyllData(context.Background(), startTime, endTime, minLat, minLon, maxLat, maxLon)
	if err != nil {
		slog.Error("error getting chlor data from db", "err", err)
	}
	fmt.Printf("Len of gotten data: %v\n", len(data))
}

func (s *Server) GetDataLastTimeTest(w http.ResponseWriter, r *http.Request) {
	t, err := s.db.GetLatestChlorophyllTimestamp(context.Background())
	if err != nil {
		slog.Error("error getting chlor data from db", "err", err)
	}
	fmt.Printf("last time: %v\n", t)
}
