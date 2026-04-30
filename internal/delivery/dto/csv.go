package dto

import "github.com/diyas/fintrack/internal/service"

type CSVImportResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

func CSVImportResponseFrom(s *service.CSVImportSummary) CSVImportResponse {
	return CSVImportResponse{
		Imported: s.Imported,
		Failed:   s.Failed,
		Errors:   s.Errors,
	}
}
