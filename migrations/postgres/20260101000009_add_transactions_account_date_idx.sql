-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_transactions_account_id_date ON transactions(account_id, transaction_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_account_id_date;
-- +goose StatementEnd
