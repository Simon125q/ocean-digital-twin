package erddap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DataInfo struct {
	Table struct {
		Rows [][]interface{} `json:"rows"`
	} `json:"table"`
}

func (d *Downloader) GetLatestDataTime(ctx context.Context, datasetID string) (time.Time, error) {
	infoURL := fmt.Sprintf("%s/%s/index.json", erddapInfoBaseURL, datasetID)
	d.logger.Info("Fetching ERDDAP info for latest time", "url", infoURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("error creating info request: %w", err)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("error fetching info from %s: %w", infoURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		d.logger.Error("ERDDAP info request returned non-OK status",
			"url", infoURL,
			"status_code", resp.StatusCode,
			"response_body", string(bodyBytes))
		return time.Time{}, fmt.Errorf("unexpected status code %d from info endpoint %s. Response: %s", resp.StatusCode, infoURL, string(bodyBytes))
	}

	var dataInfo DataInfo
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&dataInfo); err != nil {
		return time.Time{}, fmt.Errorf("error decoding info response from %s: %w", infoURL, err)
	}

	var actualRangeString string
	found := false
	for _, row := range dataInfo.Table.Rows {
		// Expected row structure: ["attribute", "time", "actual_range", "double", "min_time, max_time"]
		if len(row) >= 5 {
			rowType, ok0 := row[0].(string)
			varName, ok1 := row[1].(string)
			attrName, ok2 := row[2].(string)
			if ok0 && ok1 && ok2 && rowType == "attribute" && varName == "time" && attrName == "actual_range" {
				// Found the actual_range row for the time variable
				// The value is at index 4
				if val, ok := row[4].(string); ok {
					actualRangeString = val
					found = true
					break
				} else {
					d.logger.Error("Unexpected type or structure for 'time' actual_range value",
						"file", infoURL, "actual_value_type", fmt.Sprintf("%T", row[4]))
					return time.Time{}, fmt.Errorf("unexpected type for 'time' actual_range value in ERDDAP response from %s", infoURL)
				}
			}
		}
	}
	if !found {
		return time.Time{}, fmt.Errorf("could not find 'time' actual_range attribute in ERDDAP response from %s", infoURL)
	}

	// The max time is the second element of the actual_range array
	rangeValues := strings.Split(actualRangeString, ",")
	if len(rangeValues) != 2 {
		return time.Time{}, fmt.Errorf("unexpected format for 'time' actual_range string '%s' from %s: expected 'min,max'", actualRangeString, infoURL)
	}
	maxTimeValue, err := strconv.ParseFloat(strings.TrimSpace(rangeValues[1]), 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse max time value '%s' from %s: %w", strings.TrimSpace(rangeValues[1]), infoURL, err)
	}
	latestTime := time.Unix(int64(maxTimeValue), 0).UTC()

	d.logger.Info("Successfully retrieved latest data time", "latest_time", latestTime)
	return latestTime, nil
}
