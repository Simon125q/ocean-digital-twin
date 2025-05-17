package database

import (
	"context"
	"fmt"
	"ocean-digital-twin/internal/database/models"
	"time"
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
