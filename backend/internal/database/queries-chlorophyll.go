package database

import (
	"context"
	"fmt"
	"ocean-digital-twin/internal/database/models"
	"sort"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
)

func (s *service) SaveChlorophyllData(ctx context.Context, data []models.ChlorophyllData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO chlorophyll_data (measurement_time, location, chlor_a)
        VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4)
        `)
	defer stmt.Close()
	for _, d := range data {
		_, err := stmt.ExecContext(ctx, d.MeasurementTime, d.Longitude, d.Latitude, d.ChlorophyllA)
		if err != nil {
			return fmt.Errorf("error inserting data: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}
	return nil
}

func (s *service) SaveChlorophyllDataRaw(ctx context.Context, data []models.ChlorophyllData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO chlorophyll_data_raw (measurement_time, location, chlor_a)
        VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4)
        `)
	defer stmt.Close()
	for _, d := range data {
		_, err := stmt.ExecContext(ctx, d.MeasurementTime, d.Longitude, d.Latitude, d.ChlorophyllA)
		if err != nil {
			return fmt.Errorf("error inserting data: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}
	return nil
}

func (s *service) GetChlorophyllData(ctx context.Context, startTime, endTime time.Time, minLat, minLon, maxLat, maxLon float64, rawData bool) ([]models.ChlorophyllData, error) {
	var query string
	if !rawData {
		query = `
            SELECT 
                id,
                measurement_time,
                ST_Y(location::geometry) as latitude,
                ST_X(location::geometry) as longitude,
                chlor_a,
                created_at
            FROM
                chlorophyll_data
            WHERE
                measurement_time BETWEEN $1 AND $2
                AND ST_Intersects(
                    location::geometry,
                    ST_MakeEnvelope(
                        $3, $4, $5, $6, 4326
                    )
                )
            ORDER BY
                measurement_time
            `
	} else {
		query = `
            SELECT 
                id,
                measurement_time,
                ST_Y(location::geometry) as latitude,
                ST_X(location::geometry) as longitude,
                chlor_a,
                created_at
            FROM
                chlorophyll_data_raw
            WHERE
                measurement_time BETWEEN $1 AND $2
                AND ST_Intersects(
                    location::geometry,
                    ST_MakeEnvelope(
                        $3, $4, $5, $6, 4326
                    )
                )
            ORDER BY
                measurement_time
            `
	}
	rows, err := s.db.QueryContext(ctx, query, startTime, endTime, minLon, minLat, maxLon, maxLat)
	if err != nil {
		return nil, fmt.Errorf("error quering for chlor data: %w", err)
	}
	defer rows.Close()

	var result []models.ChlorophyllData
	for rows.Next() {
		var d models.ChlorophyllData
		err := rows.Scan(&d.ID, &d.MeasurementTime, &d.Latitude, &d.Longitude, &d.ChlorophyllA, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning chlor data: %w", err)
		}
		result = append(result, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through chlor rows: %w", err)
	}
	return result, nil
}

func (s *service) GetLatestChlorophyllTimestamp(ctx context.Context) (time.Time, error) {
	query := `
        SELECT
            COALESCE(MAX(measurement_time), '1970-01-01'::timestamp)
        FROM 
            chlorophyll_data
    `
	var result time.Time
	row := s.db.QueryRowContext(ctx, query)
	if err := row.Scan(&result); err != nil {
		return time.Time{}, fmt.Errorf("error scanning row: %w", err)
	}
	return result, nil
}

func (s *service) GetAllChlorophyllLocations(ctx context.Context) ([]orb.Point, error) {
	query := `
        SELECT DISTINCT ST_AsBinary(location) as geom
        FROM chlorophyll_data
    `
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error finding locations: %w", err)
	}
	defer rows.Close()

	var locations []orb.Point
	for rows.Next() {
		var geomBytes []byte
		if err := rows.Scan(&geomBytes); err != nil {
			return nil, fmt.Errorf("error scanning location with gap: %w", err)
		}
		geom, err := wkb.Unmarshal(geomBytes)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling location point: %w", err)
		}
		point, ok := geom.(orb.Point)
		if !ok {
			return nil, fmt.Errorf("expected geometry point, got %T", geom)
		}
		locations = append(locations, point)
	}
	return locations, nil
}

func (s *service) GetAllChlorophyllTimestamps(ctx context.Context) ([]time.Time, error) {
	query := `
        SELECT DISTINCT measurement_time
        FROM chlorophyll_data
        ORDER BY measurement_time
    `
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error finding timestamps: %w", err)
	}
	defer rows.Close()

	var timestamps []time.Time
	for rows.Next() {
		var ts time.Time
		if err := rows.Scan(&ts); err != nil {
			return nil, fmt.Errorf("failed to scan timestamp: %w", err)
		}
		timestamps = append(timestamps, ts)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return timestamps, nil
}

func (s *service) GetChlorophyllDataAtLocation(ctx context.Context, point orb.Point) ([]models.ChlorophyllData, error) {
	query := `
        SELECT 
            id,
            measurement_time,
            ST_Y(location::geometry) as latitude,
            ST_X(location::geometry) as longitude,
            chlor_a,
            created_at
        FROM
            chlorophyll_data
        WHERE
            ST_Equals(
                location::geometry,
                ST_SetSRID(
                    ST_MakePoint($1, $2),
                    4326
                )
            )
        ORDER BY
            measurement_time
    `
	rows, err := s.db.QueryContext(ctx, query, point[0], point[1])
	if err != nil {
		return nil, fmt.Errorf("error finding chlorophyll data at point (%f, %f): %w",
			point[0], point[1], err)
	}
	defer rows.Close()

	var results []models.ChlorophyllData
	for rows.Next() {
		var data models.ChlorophyllData

		if err := rows.Scan(&data.ID, &data.MeasurementTime, &data.Latitude, &data.Longitude, &data.ChlorophyllA, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning chlorophyll data: %w", err)
		}

		results = append(results, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (s *service) GetChlorophyllDataAtTimestamp(ctx context.Context, timestamp time.Time) ([][]models.ChlorophyllData, error) {
	query := `
        SELECT
            id,
            measurement_time,
            ST_Y(location::geometry) as latitude,
            ST_X(location::geometry) as longitude,
            chlor_a,
            created_at
        FROM
            chlorophyll_data
        WHERE
            measurement_time = $1
        ORDER BY
            latitude DESC, longitude ASC -- Order by latitude (descending) and longitude (ascending)
    `
	// We're now filtering by the provided timestamp
	resultRows, err := s.db.QueryContext(ctx, query, timestamp)
	if err != nil {
		return nil, fmt.Errorf("error retrieving chlorophyll data at timestamp %s: %w",
			timestamp.Format(time.RFC3339), err)
	}
	defer resultRows.Close()

	var dataList []models.ChlorophyllData
	for resultRows.Next() {
		var data models.ChlorophyllData

		// Scan the data from the row
		if err := resultRows.Scan(&data.ID, &data.MeasurementTime, &data.Latitude, &data.Longitude, &data.ChlorophyllA, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning chlorophyll data: %w", err)
		}

		dataList = append(dataList, data)
	}

	if err = resultRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// organize the flat list of data into a 2D grid based on latitude and longitude.
	// this assumes the data points form a somewhat regular grid.

	// 1. Get unique latitudes and longitudes and their sorted order
	uniqueLatitudesMap := make(map[float64]struct{})
	uniqueLongitudesMap := make(map[float64]struct{})
	for _, data := range dataList {
		uniqueLatitudesMap[data.Latitude] = struct{}{}
		uniqueLongitudesMap[data.Longitude] = struct{}{}
	}

	var uniqueLatitudes []float64
	for lat := range uniqueLatitudesMap {
		uniqueLatitudes = append(uniqueLatitudes, lat)
	}
	// Sort latitudes in descending order
	// This assumes higher latitudes are "further north" in grid
	sort.Slice(uniqueLatitudes, func(i, j int) bool {
		return uniqueLatitudes[i] > uniqueLatitudes[j]
	})

	var uniqueLongitudes []float64
	for lon := range uniqueLongitudesMap {
		uniqueLongitudes = append(uniqueLongitudes, lon)
	}
	// Sort longitudes in ascending order
	sort.Float64s(uniqueLongitudes)

	// Create a map to quickly find the index of a latitude or longitude
	latIndexMap := make(map[float64]int)
	for i, lat := range uniqueLatitudes {
		latIndexMap[lat] = i
	}

	lonIndexMap := make(map[float64]int)
	for i, lon := range uniqueLongitudes {
		lonIndexMap[lon] = i
	}

	// Initialize the 2D grid
	rows := len(uniqueLatitudes)
	cols := len(uniqueLongitudes)
	chlorophyllGrid := make([][]models.ChlorophyllData, rows)
	for i := range chlorophyllGrid {
		chlorophyllGrid[i] = make([]models.ChlorophyllData, cols)
	}

	// Populate the grid
	for _, data := range dataList {
		latIndex, latOk := latIndexMap[data.Latitude]
		lonIndex, lonOk := lonIndexMap[data.Longitude]

		if latOk && lonOk {
			// Place the data at the calculated position in the grid
			chlorophyllGrid[latIndex][lonIndex] = data
		} else {
			fmt.Printf("Warning: Data point with unexpected lat/lon found: (%f, %f)\n", data.Latitude, data.Longitude)
		}
	}

	return chlorophyllGrid, nil
}

func (s *service) UpdateChlorophyllData(ctx context.Context, data []models.ChlorophyllData) error {
	query := `
        UPDATE chlorophyll_data
        SET chlor_a = $1
        WHERE id = $2
    `
	for _, d := range data {
		_, err := s.db.ExecContext(ctx, query, d.ChlorophyllA, d.ID)
		if err != nil {
			return fmt.Errorf("error updating chlor_a: %w", err)
		}
	}
	return nil
}
