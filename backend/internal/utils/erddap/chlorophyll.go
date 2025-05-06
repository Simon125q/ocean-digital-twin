package erddap

import (
	"context"
	"fmt"
	"log/slog"
	"ocean-digital-twin/internal/database/models"
	"os"
	"path/filepath"
	"time"

	"github.com/batchatco/go-native-netcdf/netcdf"
)

func (d *Downloader) DownloadChlorophyllData(ctx context.Context, startTime, endTime time.Time) ([]models.ChlorophyllData, error) {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	maxTime, err := d.GetLatestDataTime(ctx, chlorDatasetID)
	if err != nil {
		return nil, err
	}
	if endTime.After(maxTime) {
		endTime = maxTime
	}
	url := d.buildURL(startTime, endTime, chlorDatasetID)
	d.logger.Info("Downloading chlorophyll data", "url", url)

	tempFile := filepath.Join(tempDir, fmt.Sprintf("chlor_%s_%s.nc",
		startTime.Format("20060102"), endTime.Format("20060102")))

	if err := d.downloadFile(ctx, url, tempFile); err != nil {
		return nil, err
	}
	defer os.Remove(tempFile)

	data, err := d.processChlorophyllFile(tempFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Downloader) processChlorophyllFile(filePath string) ([]models.ChlorophyllData, error) {
	nc, err := netcdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening NetCDF file: %w", err)
	}
	defer nc.Close()

	timeVar, err := nc.GetVariable("time")
	if err != nil || timeVar == nil {
		return nil, fmt.Errorf("error getting time variable: %w", err)
	}
	latVar, err := nc.GetVariable("latitude")
	if err != nil || latVar == nil {
		return nil, fmt.Errorf("error getting latitude variable: %w", err)
	}
	lonVar, err := nc.GetVariable("longitude")
	if err != nil || lonVar == nil {
		return nil, fmt.Errorf("error getting longitude variable: %w", err)
	}
	chlorVar, err := nc.GetVariable("chlor_a")
	if err != nil || chlorVar == nil {
		return nil, fmt.Errorf("error getting chlor variable: %w", err)
	}
	slog.Debug("chlor_a dimensions: ", "dim", chlorVar.Dimensions)
	slog.Debug("chlor_a attributes: ", "att", chlorVar.Attributes)

	times, ok := timeVar.Values.([]float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for time values")
	}
	lats, ok := latVar.Values.([]float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for latitude values")
	}
	lons, ok := lonVar.Values.([]float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for longitude values")
	}
	chlorData, ok := chlorVar.Values.([][][][]float32)
	if !ok {
		slog.Info("chlor_a dimensions: ", "dim", chlorVar.Dimensions)
		slog.Info("chlor_a attributes: ", "att", chlorVar.Attributes)
		return nil, fmt.Errorf("unexpected type for - Chlor_a values")
	}

	var timePoints []time.Time
	for _, t := range times {
		timePoints = append(timePoints, time.Unix(int64(t), 0).UTC())
	}

	var result []models.ChlorophyllData
	for timeIdx, t := range timePoints {
		for latIdx, lat := range lats {
			for lonIdx, lon := range lons {
				chlorValue := chlorData[timeIdx][0][latIdx][lonIdx]
				if !isNaN(float64(chlorValue)) {
					result = append(result, models.ChlorophyllData{
						MeasurementTime: t,
						Latitude:        float64(lat),
						Longitude:       float64(lon),
						ChlorophyllA:    chlorValue,
					})
				}
			}
		}
	}
	d.logger.Info("Processed NetCDF file", "points", len(result))
	return result, nil
}
