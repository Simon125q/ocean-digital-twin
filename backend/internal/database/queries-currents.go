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

func (s *service) SaveCurrentsData(ctx context.Context, data []models.CurrentsData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO currents_data (measurement_time, location, u_current, v_current)
        VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5)
        `)
	defer stmt.Close()
	for _, d := range data {
		_, err := stmt.ExecContext(ctx, d.MeasurementTime, d.Longitude, d.Latitude, d.UCurrent, d.VCurrent)
		if err != nil {
			return fmt.Errorf("error inserting data: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}
	return nil
}

func (s *service) SaveCurrentsDataRaw(ctx context.Context, data []models.CurrentsData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO currents_data_raw (measurement_time, location, u_current, v_current)
        VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5)
        `)
	defer stmt.Close()
	for _, d := range data {
		_, err := stmt.ExecContext(ctx, d.MeasurementTime, d.Longitude, d.Latitude, d.UCurrent, d.VCurrent)
		if err != nil {
			return fmt.Errorf("error inserting data: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}
	return nil
}

func (s *service) GetCurrentsData(ctx context.Context, startTime, endTime time.Time, minLat, minLon, maxLat, maxLon float64, rawData bool) ([]models.CurrentsData, error) {
	var query string
	if !rawData {
		query = `
            SELECT 
                id,
                measurement_time,
                ST_Y(location::geometry) as latitude,
                ST_X(location::geometry) as longitude,
                u_current,
                v_current,
                created_at
            FROM
                currents_data
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
                u_current,
                v_current,
                created_at
            FROM
                currents_data_raw
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
		return nil, fmt.Errorf("error quering for currents data: %w", err)
	}
	defer rows.Close()

	var result []models.CurrentsData
	for rows.Next() {
		var d models.CurrentsData
		err := rows.Scan(&d.ID, &d.MeasurementTime, &d.Latitude, &d.Longitude, &d.UCurrent, &d.VCurrent, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning currents data: %w", err)
		}
		result = append(result, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through currents rows: %w", err)
	}
	return result, nil
}

func (s *service) GetLatestCurrentsTimestamp(ctx context.Context) (time.Time, error) {
	query := `
        SELECT
            COALESCE(MAX(measurement_time), '1970-01-01'::timestamp)
        FROM 
            currents_data
    `
	var result time.Time
	row := s.db.QueryRowContext(ctx, query)
	if err := row.Scan(&result); err != nil {
		return time.Time{}, fmt.Errorf("error scanning row: %w", err)
	}
	return result, nil
}

func (s *service) GetAllCurrentsLocations(ctx context.Context) ([]orb.Point, error) {
	query := `
        SELECT DISTINCT ST_AsBinary(location) as geom
        FROM currents_data
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

func (s *service) GetUCurrentsDataAtLocation(ctx context.Context, point orb.Point) ([]models.UCurrentsData, error) {
	query := `
        SELECT 
            id,
            measurement_time,
            ST_Y(location::geometry) as latitude,
            ST_X(location::geometry) as longitude,
            u_current,
            created_at
        FROM
            currents_data
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
		return nil, fmt.Errorf("error finding u_currents data at point (%f, %f): %w",
			point[0], point[1], err)
	}
	defer rows.Close()

	var results []models.UCurrentsData
	for rows.Next() {
		var data models.UCurrentsData

		if err := rows.Scan(&data.ID, &data.MeasurementTime, &data.Latitude, &data.Longitude, &data.UCurrent, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning u_currents data: %w", err)
		}

		results = append(results, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (s *service) GetVCurrentsDataAtLocation(ctx context.Context, point orb.Point) ([]models.VCurrentsData, error) {
	query := `
        SELECT 
            id,
            measurement_time,
            ST_Y(location::geometry) as latitude,
            ST_X(location::geometry) as longitude,
            v_current,
            created_at
        FROM
            currents_data
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
		return nil, fmt.Errorf("error finding v_currents data at point (%f, %f): %w",
			point[0], point[1], err)
	}
	defer rows.Close()

	var results []models.VCurrentsData
	for rows.Next() {
		var data models.VCurrentsData

		if err := rows.Scan(&data.ID, &data.MeasurementTime, &data.Latitude, &data.Longitude, &data.VCurrent, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning v_currents data: %w", err)
		}

		results = append(results, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (s *service) GetAllCurrentsTimestamps(ctx context.Context) ([]time.Time, error) {
	query := `
        SELECT DISTINCT measurement_time
        FROM currents_data
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

func (s *service) GetVCurrentDataAtTimestamp(ctx context.Context, timestamp time.Time) ([][]models.VCurrentsData, error) {
	query := `
        SELECT
            id,
            measurement_time,
            ST_Y(location::geometry) as latitude,
            ST_X(location::geometry) as longitude,
            v_current,
            created_at
        FROM
            currents_data
        WHERE
            measurement_time = $1
        ORDER BY
            latitude DESC, longitude ASC -- Order by latitude (descending) and longitude (ascending)
    `
	// We're now filtering by the provided timestamp
	resultRows, err := s.db.QueryContext(ctx, query, timestamp)
	if err != nil {
		return nil, fmt.Errorf("error retrieving currents data at timestamp %s: %w",
			timestamp.Format(time.RFC3339), err)
	}
	defer resultRows.Close()

	var dataList []models.VCurrentsData
	for resultRows.Next() {
		var data models.VCurrentsData

		// Scan the data from the row
		if err := resultRows.Scan(&data.ID, &data.MeasurementTime, &data.Latitude, &data.Longitude, &data.VCurrent, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning v_current data: %w", err)
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
	vCurrentsGrid := make([][]models.VCurrentsData, rows)
	for i := range vCurrentsGrid {
		vCurrentsGrid[i] = make([]models.VCurrentsData, cols)
	}

	// Populate the grid
	for _, data := range dataList {
		latIndex, latOk := latIndexMap[data.Latitude]
		lonIndex, lonOk := lonIndexMap[data.Longitude]

		if latOk && lonOk {
			// Place the data at the calculated position in the grid
			vCurrentsGrid[latIndex][lonIndex] = data
		} else {
			fmt.Printf("Warning: Data point with unexpected lat/lon found: (%f, %f)\n", data.Latitude, data.Longitude)
		}
	}

	return vCurrentsGrid, nil
}

func (s *service) GetUCurrentDataAtTimestamp(ctx context.Context, timestamp time.Time) ([][]models.UCurrentsData, error) {
	query := `
        SELECT
            id,
            measurement_time,
            ST_Y(location::geometry) as latitude,
            ST_X(location::geometry) as longitude,
            u_current,
            created_at
        FROM
            currents_data
        WHERE
            measurement_time = $1
        ORDER BY
            latitude DESC, longitude ASC -- Order by latitude (descending) and longitude (ascending)
    `
	// We're now filtering by the provided timestamp
	resultRows, err := s.db.QueryContext(ctx, query, timestamp)
	if err != nil {
		return nil, fmt.Errorf("error retrieving currents data at timestamp %s: %w",
			timestamp.Format(time.RFC3339), err)
	}
	defer resultRows.Close()

	var dataList []models.UCurrentsData
	for resultRows.Next() {
		var data models.UCurrentsData

		// Scan the data from the row
		if err := resultRows.Scan(&data.ID, &data.MeasurementTime, &data.Latitude, &data.Longitude, &data.UCurrent, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning u_current data: %w", err)
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
	uCurrentsGrid := make([][]models.UCurrentsData, rows)
	for i := range uCurrentsGrid {
		uCurrentsGrid[i] = make([]models.UCurrentsData, cols)
	}

	// Populate the grid
	for _, data := range dataList {
		latIndex, latOk := latIndexMap[data.Latitude]
		lonIndex, lonOk := lonIndexMap[data.Longitude]

		if latOk && lonOk {
			// Place the data at the calculated position in the grid
			uCurrentsGrid[latIndex][lonIndex] = data
		} else {
			fmt.Printf("Warning: Data point with unexpected lat/lon found: (%f, %f)\n", data.Latitude, data.Longitude)
		}
	}

	return uCurrentsGrid, nil
}

func (s *service) UpdateUCurrentsData(ctx context.Context, data []models.UCurrentsData) error {
	query := `
        UPDATE currents_data
        SET u_current = $1
        WHERE id = $2
    `
	for _, d := range data {
		_, err := s.db.ExecContext(ctx, query, d.UCurrent, d.ID)
		if err != nil {
			return fmt.Errorf("error updating u_current: %w", err)
		}
	}
	return nil
}

func (s *service) UpdateVCurrentsData(ctx context.Context, data []models.VCurrentsData) error {
	query := `
        UPDATE currents_data
        SET v_current = $1
        WHERE id = $2
    `
	for _, d := range data {
		_, err := s.db.ExecContext(ctx, query, d.VCurrent, d.ID)
		if err != nil {
			return fmt.Errorf("error updating v_current: %w", err)
		}
	}
	return nil
}
