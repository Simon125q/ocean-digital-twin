-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_cron;
SELECT cron.schedule('cleanup-chlorophyll-data', '0 3 * * *', 'SELECT cleanup_chlorophyll_data()');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- pause the job
SELECT cron.unschedule('cleanup-chlorophyll-data');
-- delete the job
DELETE FROM cron.job WHERE jobname = 'cleanup-chlorophyll-data';
-- +goose StatementEnd
