package database

import (
	"context"
	"fmt"
	"ocean-digital-twin/internal/database/models"
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
