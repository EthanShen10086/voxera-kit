// Package dataprovider defines the port interfaces for financial data retrieval,
// including quotes, financial statements, and market search.
package dataprovider

import (
	"context"
	"time"
)

// Market describes a stock exchange or trading venue.
type Market struct {
	Code     string // e.g., "SH", "SZ", "HK", "US"
	Name     string
	Currency string
	Timezone string
}

// Quote holds a real-time or delayed price snapshot for a security.
type Quote struct {
	Symbol    string
	Name      string
	Market    string
	Price     float64
	Change    float64
	ChangePct float64
	Volume    int64
	MarketCap float64
	PE        float64
	PB        float64
	Timestamp time.Time
}

// Period specifies the reporting frequency for financial statements.
type Period string

// Reporting period constants.
const (
	PeriodAnnual    Period = "annual"
	PeriodQuarterly Period = "quarterly"
	PeriodTTM       Period = "ttm"
)

// FinancialStatements aggregates income statement, balance sheet, cash flow, and key metrics.
type FinancialStatements struct {
	Symbol          string
	Period          Period
	FiscalYears     []string
	IncomeStatement []IncomeStatementItem
	BalanceSheet    []BalanceSheetItem
	CashFlow        []CashFlowItem
	KeyMetrics      []KeyMetric
}

// IncomeStatementItem represents a single line item in an income statement.
type IncomeStatementItem struct {
	Label    string
	Key      string
	Values   map[string]float64 // fiscalYear -> value
	IsGAAP   bool
	Category string // "revenue", "expense", "profit"
}

// BalanceSheetItem represents a single line item in a balance sheet.
type BalanceSheetItem struct {
	Label    string
	Key      string
	Values   map[string]float64
	Category string // "asset", "liability", "equity"
}

// CashFlowItem represents a single line item in a cash flow statement.
type CashFlowItem struct {
	Label    string
	Key      string
	Values   map[string]float64
	Category string // "operating", "investing", "financing"
}

// KeyMetric represents a derived financial ratio or metric.
type KeyMetric struct {
	Label  string
	Key    string
	Values map[string]float64
	Unit   string // "ratio", "percentage", "currency", "multiple"
}

// SearchResult represents a single match from a security search query.
type SearchResult struct {
	Symbol   string
	Name     string
	Market   string
	Type     string // "stock", "etf", "index"
	Currency string
}

// FetchParams specifies filters for a data fetch request.
type FetchParams struct {
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	Fields    []string
}

// DataSet holds columnar data returned by a fetch operation.
type DataSet struct {
	Columns []string
	Rows    []map[string]interface{}
}

// DataHandler is a callback invoked when new data arrives via a subscription.
type DataHandler func(data *DataSet)

// DataProvider is the interface for searching, fetching, and subscribing to data sources.
type DataProvider interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
	Fetch(ctx context.Context, sourceID string, params FetchParams) (*DataSet, error)
	Subscribe(ctx context.Context, sourceID string, handler DataHandler) error
	Close() error
}

// FinancialProvider extends DataProvider with financial-specific operations.
type FinancialProvider interface {
	DataProvider
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetFinancials(ctx context.Context, symbol string, period Period) (*FinancialStatements, error)
	ListMarkets() []Market
}

// ProviderConfig holds configuration for a data provider backend.
type ProviderConfig struct {
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
	CacheTTL time.Duration
}
