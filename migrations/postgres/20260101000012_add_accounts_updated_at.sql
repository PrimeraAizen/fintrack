-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
