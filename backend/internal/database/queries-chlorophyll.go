package database

import (
	"context"
	"fmt"
	"ocean-digital-twin/internal/database/models"
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
