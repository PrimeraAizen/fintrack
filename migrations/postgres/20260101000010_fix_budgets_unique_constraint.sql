-- +goose Up
-- +goose StatementBegin
ALTER TABLE budgets DROP CONSTRAINT budgets_user_id_category_id_key;
ALTER TABLE budgets ADD CONSTRAINT budgets_user_id_category_id_period_key UNIQUE (user_id, category_id, period);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE budgets DROP CONSTRAINT budgets_user_id_category_id_period_key;
ALTER TABLE budgets ADD CONSTRAINT budgets_user_id_category_id_key UNIQUE (user_id, category_id);
-- +goose StatementEnd
