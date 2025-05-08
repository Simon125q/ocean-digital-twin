package erddap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	ChlorDatasetID     = "noaacwNPPVIIRSchlaDaily"
	erddapBaseURL      = "https://coastwatch.noaa.gov/erddap/griddap/"
	erddapInfoBaseURL  = "https://coastwatch.noaa.gov/erddap/info/"
	fileType           = "nc"
	tempDir            = "tmp/erddap"
	defaultHTTPTimeout = 1500 * time.Second
)

type Downloader struct {
	logger     *slog.Logger
	minLat     float64
	maxLat     float64
	minLon     float64
	maxLon     float64
	httpClient *http.Client
}

func NewDownloader(logger *slog.Logger, minLat, minLon, maxLat, maxLon float64) *Downloader {
	return &Downloader{
		logger: logger,
		minLat: minLat,
		maxLat: maxLat,
		minLon: minLon,
		maxLon: maxLon,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

func (d *Downloader) buildURL(startTime, endTime time.Time, datasetID string) string {
	startStr := startTime.Format("2006-01-02T15:04:05Z")
	endStr := endTime.Format("2006-01-02T15:04:05Z")

	query := fmt.Sprintf("chlor_a[(%s):1:(%s)][(0.0):1:(0.0)][(%.5f):1:(%.5f)][(%.5f):1:(%.5f)]",
		startStr, endStr, d.minLat, d.maxLat, d.minLon, d.maxLon)

	return fmt.Sprintf("%s/%s.%s?%s", erddapBaseURL, datasetID, fileType, query)
}

func (d *Downloader) downloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			return fmt.Errorf("error downloading file from %s: request timed out after %s: %w", url, d.httpClient.Timeout, err)
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("error downloading file from %s: context deadline exceeded: %w", url, err)
		}
		return fmt.Errorf("error downloading file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		d.logger.Error("ERDDAP request returned non-OK status",
			"status_code", resp.StatusCode,
			"response_body", string(bodyBytes))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating a file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error coping response to file: %w", err)
	}

	d.logger.Info("File downloaded successfully", "path", destPath)
	return nil
}

func isNaN(f float64) bool {
	return f != f
}
