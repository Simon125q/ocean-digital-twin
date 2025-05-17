-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_cron;
SELECT cron.schedule('cleanup-currents-data', '0 3 * * *', 'SELECT cleanup_currents_data()');
SELECT cron.schedule('cleanup-currents-data-raw', '0 3 * * *', 'SELECT cleanup_currents_raw_data()');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- pause the job
SELECT cron.unschedule('cleanup-currents-data');
SELECT cron.unschedule('cleanup-currents-data-raw');
-- delete the job
DELETE FROM cron.job WHERE jobname = 'cleanup-currents-data';
DELETE FROM cron.job WHERE jobname = 'cleanup-currents-data-raw';
-- +goose StatementEnd
