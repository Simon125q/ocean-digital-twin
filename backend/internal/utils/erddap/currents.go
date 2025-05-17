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

func (d *Downloader) DownloadCurrentsData(ctx context.Context, startTime, endTime time.Time) ([]models.CurrentsData, error) {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	maxTime, err := d.GetLatestDataTime(ctx, CurrentsDatasetID)
	if err != nil {
		return nil, err
	}
	if endTime.After(maxTime) {
		endTime = maxTime
	}
	vars := []string{"u_current", "v_current"}
	url := d.buildURLWithVars(startTime, endTime, CurrentsDatasetID, vars)
	d.logger.Info("Downloading currents data", "url", url)

	tempFile := filepath.Join(tempDir, fmt.Sprintf("chlor_%s_%s.nc",
		startTime.Format("20060102"), endTime.Format("20060102")))

	if err := d.downloadFile(ctx, url, tempFile); err != nil {
		return nil, err
	}
	defer os.Remove(tempFile)

	data, err := d.processCurrentsFile(tempFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Downloader) processCurrentsFile(filePath string) ([]models.CurrentsData, error) {
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
	uCurrentVar, err := nc.GetVariable("u_current")
	if err != nil || uCurrentVar == nil {
		return nil, fmt.Errorf("error getting u_current variable: %w", err)
	}
	vCurrentVar, err := nc.GetVariable("v_current")
	if err != nil || vCurrentVar == nil {
		return nil, fmt.Errorf("error getting u_current variable: %w", err)
	}

	times, ok := timeVar.Values.([]float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for time values")
	}
	lats, ok := latVar.Values.([]float32)
	if !ok {
		fmt.Printf("lat type: %T", latVar.Values)
		return nil, fmt.Errorf("unexpected type for latitude values")
	}
	lons, ok := lonVar.Values.([]float32)
	if !ok {
		fmt.Printf("lon type: %T", lonVar.Values)
		return nil, fmt.Errorf("unexpected type for longitude values")
	}
	uCurrentData, ok := uCurrentVar.Values.([][][]float64)
	if !ok {
		fmt.Printf("u_current type: %T\n", uCurrentVar.Values)
		slog.Info("u_current dimensions: ", "dim", uCurrentVar.Dimensions)
		slog.Info("u_current attributes: ", "att", uCurrentVar.Attributes)
		return nil, fmt.Errorf("unexpected type for u_current values")
	}
	vCurrentData, ok := vCurrentVar.Values.([][][]float64)
	if !ok {
		fmt.Printf("v_current type: %T\n", vCurrentVar.Values)
		slog.Info("v_current dimensions: ", "dim", uCurrentVar.Dimensions)
		slog.Info("v_current attributes: ", "att", uCurrentVar.Attributes)
		return nil, fmt.Errorf("unexpected type for v_current values")
	}

	var timePoints []time.Time
	for _, t := range times {
		timePoints = append(timePoints, time.Unix(int64(t), 0).UTC())
	}

	var result []models.CurrentsData
	for timeIdx, t := range timePoints {
		for latIdx, lat := range lats {
			for lonIdx, lon := range lons {
				uCurrentValue := uCurrentData[timeIdx][latIdx][lonIdx]
				vCurrentValue := vCurrentData[timeIdx][latIdx][lonIdx]
				result = append(result, models.CurrentsData{
					MeasurementTime: t,
					Latitude:        float64(lat),
					Longitude:       float64(lon),
					UCurrent:        float32(uCurrentValue),
					VCurrent:        float32(vCurrentValue),
				})
			}
		}
	}
	d.logger.Info("Processed NetCDF file", "points", len(result))
	return result, nil
}
