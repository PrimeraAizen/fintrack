package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diyas/fintrack/config"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

const (
	seedEmail    = "demo@fintrack.local"
	seedPassword = "demo1234"
	seedCurrency = "KZT"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pg, err := postgres.New(ctx, &cfg.PG)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pg.Close()

	if err := seed(ctx, pg.Pool); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	log.Printf("seed complete — user %q password %q", seedEmail, seedPassword)
}

func seed(ctx context.Context, pool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE email = $1`, seedEmail); err != nil {
		return fmt.Errorf("clear seed user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	var userID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, base_currency)
		VALUES ($1, $2, $3) RETURNING id`,
		seedEmail, string(hash), seedCurrency,
	).Scan(&userID); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	categoryIDs, err := seedCategories(ctx, tx, userID)
	if err != nil {
		return err
	}

	accountIDs, err := seedAccounts(ctx, tx, userID)
	if err != nil {
		return err
	}

	if err := seedTransactions(ctx, tx, accountIDs, categoryIDs); err != nil {
		return err
	}

	if err := seedBudget(ctx, tx, userID, categoryIDs["Food"]); err != nil {
		return err
	}

	if err := seedTransfer(ctx, tx, accountIDs["Cash"], accountIDs["Card"]); err != nil {
		return err
	}

	if err := seedSavingGoal(ctx, tx, userID); err != nil {
		return err
	}

	if err := seedExchangeRates(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func seedCategories(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (map[string]uuid.UUID, error) {
	categories := []struct {
		Name string
		Type string
		Icon string
	}{
		{"Salary", "income", "briefcase"},
		{"Freelance", "income", "laptop"},
		{"Investment", "income", "trending-up"},
		{"Gift", "income", "gift"},
		{"Food", "expense", "utensils"},
		{"Transport", "expense", "car"},
		{"Housing", "expense", "home"},
		{"Entertainment", "expense", "film"},
		{"Health", "expense", "heart"},
		{"Education", "expense", "book"},
		{"Shopping", "expense", "shopping-bag"},
		{"Other", "expense", "more-horizontal"},
	}
	out := make(map[string]uuid.UUID, len(categories))
	for _, c := range categories {
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO categories (user_id, name, type, icon)
			VALUES ($1, $2, $3, $4) RETURNING id`,
			userID, c.Name, c.Type, c.Icon,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert category %s: %w", c.Name, err)
		}
		out[c.Name] = id
	}
	return out, nil
}

func seedAccounts(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (map[string]uuid.UUID, error) {
	accounts := []struct {
		Name     string
		Type     string
		Currency string
		Balance  decimal.Decimal
	}{
		{"Cash", "cash", "KZT", decimal.NewFromInt(50000)},
		{"Card", "card", "KZT", decimal.NewFromInt(250000)},
		{"Savings", "savings", "USD", decimal.NewFromInt(1500)},
	}
	out := make(map[string]uuid.UUID, len(accounts))
	for _, a := range accounts {
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO accounts (user_id, name, type, currency, balance)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			userID, a.Name, a.Type, a.Currency, a.Balance,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert account %s: %w", a.Name, err)
		}
		out[a.Name] = id
	}
	return out, nil
}

type seedTx struct {
	Account  string
	Category string
	Type     string
	Amount   int64
	DaysAgo  int
	Note     string
}

func seedTransactions(ctx context.Context, tx pgx.Tx, accounts, categories map[string]uuid.UUID) error {
	rng := rand.New(rand.NewSource(42))
	now := time.Now().UTC()
	plan := []seedTx{
		{"Card", "Salary", "income", 800000, 30, "March payroll"},
		{"Card", "Salary", "income", 800000, 1, "April payroll"},
		{"Cash", "Freelance", "income", 120000, 15, "Logo design"},
		{"Card", "Investment", "income", 25000, 10, "Dividends"},

		{"Card", "Housing", "expense", 180000, 28, "Rent"},
		{"Card", "Housing", "expense", 18000, 25, "Utilities"},
		{"Cash", "Food", "expense", 8500, 27, "Groceries"},
		{"Card", "Food", "expense", 12500, 22, "Groceries"},
		{"Cash", "Food", "expense", 4200, 18, "Lunch"},
		{"Card", "Food", "expense", 9800, 12, "Dinner out"},
		{"Card", "Food", "expense", 14000, 6, "Groceries"},
		{"Cash", "Food", "expense", 3500, 2, "Coffee shop"},
		{"Card", "Transport", "expense", 22000, 26, "Fuel"},
		{"Cash", "Transport", "expense", 1500, 20, "Bus pass"},
		{"Card", "Transport", "expense", 8500, 9, "Taxi"},
		{"Card", "Entertainment", "expense", 15000, 21, "Concert tickets"},
		{"Cash", "Entertainment", "expense", 4500, 14, "Cinema"},
		{"Card", "Health", "expense", 35000, 17, "Pharmacy"},
		{"Card", "Education", "expense", 60000, 11, "Online course"},
		{"Card", "Shopping", "expense", 28000, 8, "Clothing"},
		{"Card", "Shopping", "expense", 11500, 4, "Books"},
		{"Cash", "Other", "expense", 5000, 3, "Misc"},
	}

	balanceDelta := make(map[uuid.UUID]decimal.Decimal)
	for _, p := range plan {
		amount := decimal.NewFromInt(p.Amount + int64(rng.Intn(500)))
		date := now.AddDate(0, 0, -p.DaysAgo)
		accountID := accounts[p.Account]
		categoryID := categories[p.Category]
		if accountID == uuid.Nil || categoryID == uuid.Nil {
			return fmt.Errorf("missing seed reference for %s/%s", p.Account, p.Category)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO transactions (account_id, category_id, amount, currency, converted_amount, note, transaction_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			accountID, categoryID, amount, "KZT", amount, p.Note, date,
		)
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
		delta := amount
		if p.Type == "expense" {
			delta = amount.Neg()
		}
		balanceDelta[accountID] = balanceDelta[accountID].Add(delta)
	}

	for accountID, delta := range balanceDelta {
		if _, err := tx.Exec(ctx,
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
			delta, accountID,
		); err != nil {
			return fmt.Errorf("update account balance: %w", err)
		}
	}
	return nil
}

func seedBudget(ctx context.Context, tx pgx.Tx, userID, foodCategoryID uuid.UUID) error {
	if foodCategoryID == uuid.Nil {
		return nil
	}
	periodStart := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	_, err := tx.Exec(ctx, `
		INSERT INTO budgets (user_id, category_id, spending_limit, period, spent, period_start)
		VALUES ($1, $2, $3, 'monthly', 0, $4)`,
		userID, foodCategoryID, decimal.NewFromInt(80000), periodStart,
	)
	if err != nil {
		return fmt.Errorf("insert budget: %w", err)
	}
	return nil
}

func seedTransfer(ctx context.Context, tx pgx.Tx, fromAccount, toAccount uuid.UUID) error {
	if fromAccount == uuid.Nil || toAccount == uuid.Nil {
		return nil
	}
	amount := decimal.NewFromInt(20000)
	if _, err := tx.Exec(ctx, `
		INSERT INTO transfers (from_account_id, to_account_id, amount, from_currency, to_currency, exchange_rate)
		VALUES ($1, $2, $3, 'KZT', 'KZT', 1)`,
		fromAccount, toAccount, amount,
	); err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromAccount); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toAccount); err != nil {
		return err
	}
	return nil
}

func seedSavingGoal(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	deadline := time.Now().UTC().AddDate(1, 0, 0)
	_, err := tx.Exec(ctx, `
		INSERT INTO saving_goals (user_id, name, target_amount, current_amount, deadline)
		VALUES ($1, 'Emergency Fund', $2, $3, $4)`,
		userID, decimal.NewFromInt(2000000), decimal.NewFromInt(450000), deadline,
	)
	if err != nil {
		return fmt.Errorf("insert saving goal: %w", err)
	}
	return nil
}

func seedExchangeRates(ctx context.Context, tx pgx.Tx) error {
	rates := []struct {
		Base, Target string
		Rate         string
	}{
		{"USD", "KZT", "470.250000"},
		{"EUR", "KZT", "510.800000"},
		{"RUB", "KZT", "5.150000"},
		{"KZT", "USD", "0.002126"},
		{"KZT", "EUR", "0.001958"},
	}
	for _, r := range rates {
		rate, err := decimal.NewFromString(r.Rate)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO exchange_rates (base_currency, target_currency, rate, cached_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (base_currency, target_currency)
			DO UPDATE SET rate = EXCLUDED.rate, cached_at = NOW()`,
			r.Base, r.Target, rate,
		); err != nil {
			return fmt.Errorf("insert exchange rate %s/%s: %w", r.Base, r.Target, err)
		}
	}
	return nil
}
