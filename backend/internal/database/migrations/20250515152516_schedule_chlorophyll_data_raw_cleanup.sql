-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_cron;
SELECT cron.schedule('cleanup-chlorophyll-data-raw', '0 3 * * *', 'SELECT cleanup_chlorophyll_data_raw()');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- pause the job
SELECT cron.unschedule('cleanup-chlorophyll-data-raw');
-- delete the job
DELETE FROM cron.job WHERE jobname = 'cleanup-chlorophyll-data-raw';
-- +goose StatementEnd
