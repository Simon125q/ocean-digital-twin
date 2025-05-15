-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS chlorophyll_data_raw (
    id SERIAL PRIMARY KEY,
    measurement_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    chlor_a FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- Index for spatial queries
CREATE INDEX IF NOT EXISTS chlorophyll_data_raw_location_idx ON chlorophyll_data_raw USING GIST(location);

-- Index for time-based queries
CREATE INDEX IF NOT EXISTS chlorophyll_data_raw_time_idx ON chlorophyll_data_raw(measurement_time);

-- Function to clean up old data (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_chlorophyll_data_raw() RETURNS void AS $$
BEGIN
    DELETE FROM chlorophyll_data_raw
    WHERE measurement_time < (NOW() - INTERVAL '120 days');
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the cleanup function
DROP FUNCTION IF EXISTS cleanup_old_chlorophyll_data_raw();
-- Drop the indexes
DROP INDEX IF EXISTS chlorophyll_data_raw_time_idx;
DROP INDEX IF EXISTS chlorophyll_data_raw_location_idx;
-- Drop the main table
DROP TABLE IF EXISTS chlorophyll_data_raw;

-- +goose StatementEnd
