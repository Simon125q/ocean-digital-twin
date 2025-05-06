#!/bin/bash
set -e

# Wait for PostgreSQL to start
until pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
  echo "Waiting for PostgreSQL to start..."
  sleep 1
done

# Create extensions
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS postgis;
    CREATE EXTENSION IF NOT EXISTS postgis_topology;
    CREATE EXTENSION IF NOT EXISTS pg_cron;
    
    -- Grant permissions for pg_cron if needed
    GRANT USAGE ON SCHEMA cron TO "$POSTGRES_USER";
EOSQL

echo "PostgreSQL initialized with PostGIS and pg_cron extensions"

