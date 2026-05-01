package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const csvDateLayout = "2006-01-02"

var csvHeader = []string{"date", "account", "category", "type", "amount", "currency", "note"}

type csvCategoryKey struct {
	name string
	typ  string
}

type CSVImportSummary struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

type CSV interface {
	Export(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter, w io.Writer) error
	Import(ctx context.Context, userID uuid.UUID, r io.Reader) (*CSVImportSummary, error)
}

type CSVService struct {
	transactions Transaction
	accounts     repository.Account
	categories   repository.Category
	txRepo       repository.Transaction
}

func NewCSVService(
	transactions Transaction,
	accounts repository.Account,
	categories repository.Category,
	txRepo repository.Transaction,
) *CSVService {
	return &CSVService{
		transactions: transactions,
		accounts:     accounts,
		categories:   categories,
		txRepo:       txRepo,
	}
}

func (s *CSVService) Export(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter, w io.Writer) error {
	accounts, err := s.accounts.List(ctx, userID)
	if err != nil {
		return err
	}
	accountByID := make(map[uuid.UUID]domain.Account, len(accounts))
	for _, a := range accounts {
		accountByID[a.ID] = a
	}

	categories, err := s.categories.List(ctx, userID, "")
	if err != nil {
		return err
	}
	categoryByID := make(map[uuid.UUID]domain.Category, len(categories))
	for _, c := range categories {
		categoryByID[c.ID] = c
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()
	if err := writer.Write(csvHeader); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	const pageSize = 500
	page := 1
	for {
		f := filter
		f.Page = page
		f.PerPage = pageSize
		list, _, err := s.txRepo.List(ctx, userID, f)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		for _, t := range list {
			account := accountByID[t.AccountID]
			category := categoryByID[t.CategoryID]
			row := []string{
				t.TransactionDate.Format(csvDateLayout),
				account.Name,
				category.Name,
				category.Type,
				t.Amount.String(),
				t.Currency,
				t.Note,
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("write csv row: %w", err)
			}
		}
		if len(list) < pageSize {
			break
		}
		page++
	}
	writer.Flush()
	return writer.Error()
}

func (s *CSVService) Import(ctx context.Context, userID uuid.UUID, r io.Reader) (*CSVImportSummary, error) {
	accounts, err := s.accounts.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	accountByName := make(map[string]domain.Account, len(accounts))
	for _, a := range accounts {
		accountByName[strings.ToLower(a.Name)] = a
	}

	categories, err := s.categories.List(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	categoryByKey := make(map[csvCategoryKey]domain.Category, len(categories))
	for _, c := range categories {
		categoryByKey[csvCategoryKey{name: strings.ToLower(c.Name), typ: c.Type}] = c
	}

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("%w: empty csv", domain.ErrInvalidInput)
	}
	idx, err := indexHeader(header)
	if err != nil {
		return nil, err
	}

	summary := &CSVImportSummary{}
	line := 1
	for {
		line++
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			summary.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("line %d: %s", line, err.Error()))
			continue
		}
		if err := s.importRow(ctx, userID, idx, row, accountByName, categoryByKey); err != nil {
			summary.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("line %d: %s", line, err.Error()))
			continue
		}
		summary.Imported++
	}
	return summary, nil
}

type headerIndex struct {
	date, account, category, typ, amount, currency, note int
}

func indexHeader(header []string) (headerIndex, error) {
	idx := headerIndex{date: -1, account: -1, category: -1, typ: -1, amount: -1, currency: -1, note: -1}
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "date":
			idx.date = i
		case "account":
			idx.account = i
		case "category":
			idx.category = i
		case "type":
			idx.typ = i
		case "amount":
			idx.amount = i
		case "currency":
			idx.currency = i
		case "note":
			idx.note = i
		}
	}
	if idx.date < 0 || idx.account < 0 || idx.category < 0 || idx.typ < 0 || idx.amount < 0 {
		return idx, fmt.Errorf("%w: csv must have date, account, category, type, amount columns", domain.ErrInvalidInput)
	}
	return idx, nil
}

func (s *CSVService) importRow(
	ctx context.Context,
	userID uuid.UUID,
	idx headerIndex,
	row []string,
	accountByName map[string]domain.Account,
	categoryByKey map[csvCategoryKey]domain.Category,
) error {
	get := func(i int) string {
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	dateStr := get(idx.date)
	date, err := time.Parse(csvDateLayout, dateStr)
	if err != nil {
		return fmt.Errorf("invalid date %q", dateStr)
	}

	accountName := get(idx.account)
	account, ok := accountByName[strings.ToLower(accountName)]
	if !ok {
		return fmt.Errorf("unknown account %q", accountName)
	}

	categoryName := get(idx.category)
	categoryType := strings.ToLower(get(idx.typ))
	category, ok := categoryByKey[csvCategoryKey{name: strings.ToLower(categoryName), typ: categoryType}]
	if !ok {
		return fmt.Errorf("unknown category %q (%s)", categoryName, categoryType)
	}

	amountStr := get(idx.amount)
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount %q", amountStr)
	}

	currency := strings.ToUpper(get(idx.currency))
	if currency == "" {
		currency = account.Currency
	}

	note := ""
	if idx.note >= 0 {
		note = get(idx.note)
	}

	_, err = s.transactions.Create(ctx, userID, CreateTransactionInput{
		AccountID:       account.ID,
		CategoryID:      category.ID,
		Amount:          amount,
		Currency:        currency,
		Note:            note,
		TransactionDate: &date,
	})
	return err
}
