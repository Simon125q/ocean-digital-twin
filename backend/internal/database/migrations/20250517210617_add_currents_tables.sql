-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS currents_data (
    id SERIAL PRIMARY KEY,
    measurement_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    u_current FLOAT,
    v_current FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS currents_data_raw (
    id SERIAL PRIMARY KEY,
    measurement_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    u_current FLOAT,
    v_current FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- Index for spatial queries
CREATE INDEX IF NOT EXISTS currents_data_location_idx ON currents_data USING GIST(location);
CREATE INDEX IF NOT EXISTS currents_data_raw_location_idx ON currents_data_raw USING GIST(location);

-- Index for time-based queries
CREATE INDEX IF NOT EXISTS currents_data_time_idx ON currents_data(measurement_time);
CREATE INDEX IF NOT EXISTS currents_data_raw_time_idx ON currents_data_raw(measurement_time);

-- Function to clean up old data (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_currents_data() RETURNS void AS $$
BEGIN
    DELETE FROM currents_data 
    WHERE measurement_time < (NOW() - INTERVAL '120 days');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_old_currents_data_raw() RETURNS void AS $$
BEGIN
    DELETE FROM currents_data_raw 
    WHERE measurement_time < (NOW() - INTERVAL '120 days');
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the cleanup function
DROP FUNCTION IF EXISTS cleanup_old_currents_data();
DROP FUNCTION IF EXISTS cleanup_old_currents_data_raw();
-- Drop the indexes
DROP INDEX IF EXISTS currents_data_time_idx;
DROP INDEX IF EXISTS currents_data_raw_time_idx;
DROP INDEX IF EXISTS currents_data_location_idx;
DROP INDEX IF EXISTS currents_data_raw_location_idx;
-- Drop the main table
DROP TABLE IF EXISTS currents_data;
DROP TABLE IF EXISTS currents_data_raw;

-- +goose StatementEnd
