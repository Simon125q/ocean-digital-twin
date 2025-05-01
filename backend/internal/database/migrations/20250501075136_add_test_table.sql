-- +goose Up
-- +goose StatementBegin
CREATE TABLE test (
    id SERIAL PRIMARY KEY,
    count INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE test
-- +goose StatementEnd
