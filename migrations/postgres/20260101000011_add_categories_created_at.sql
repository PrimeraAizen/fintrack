-- +goose Up
-- +goose StatementBegin
ALTER TABLE categories ADD COLUMN created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE categories DROP COLUMN IF EXISTS created_at;
-- +goose StatementEnd
