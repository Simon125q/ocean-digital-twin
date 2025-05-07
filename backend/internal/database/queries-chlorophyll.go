package database

import (
	"context"
	"fmt"
	"ocean-digital-twin/internal/database/models"
	"time"
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

func (s *service) GetChlorophyllData(ctx context.Context, startTime, endTime time.Time, minLat, minLon, maxLat, maxLon float64) ([]models.ChlorophyllData, error) {
	query := `
        SELECT 
            measurement_time,
            ST_Y(location::geography) as latitude,
            ST_X(location::geography) as longitude,
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
	rows, err := s.db.QueryContext(ctx, query, startTime, endTime, minLon, minLat, maxLon, maxLat)
	if err != nil {
		return nil, fmt.Errorf("error quering for chlor data: %w", err)
	}
	defer rows.Close()

	var result []models.ChlorophyllData
	for rows.Next() {
		var d models.ChlorophyllData
		err := rows.Scan(&d.MeasurementTime, &d.Latitude, &d.Longitude, &d.ChlorophyllA, &d.CreatedAt)
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
