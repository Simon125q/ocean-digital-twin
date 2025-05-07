package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"ocean-digital-twin/internal/utils/erddap"
)

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
