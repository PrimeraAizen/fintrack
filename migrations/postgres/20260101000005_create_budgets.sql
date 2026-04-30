-- +goose Up
-- +goose StatementBegin
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    spending_limit NUMERIC(15,2) NOT NULL,
    period VARCHAR(10) NOT NULL CHECK (period IN ('weekly', 'monthly')),
    spent NUMERIC(15,2) NOT NULL DEFAULT 0,
    period_start DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, category_id)
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS budgets;
-- +goose StatementEnd
