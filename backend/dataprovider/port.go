package dataprovider

import (
	"context"
	"time"
)

type Market struct {
	Code     string // e.g., "SH", "SZ", "HK", "US"
	Name     string
	Currency string
	Timezone string
}

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

type Period string

const (
	PeriodAnnual    Period = "annual"
	PeriodQuarterly Period = "quarterly"
	PeriodTTM       Period = "ttm"
)

type FinancialStatements struct {
	Symbol          string
	Period          Period
	FiscalYears     []string
	IncomeStatement []IncomeStatementItem
	BalanceSheet    []BalanceSheetItem
	CashFlow        []CashFlowItem
	KeyMetrics      []KeyMetric
}

type IncomeStatementItem struct {
	Label    string
	Key      string
	Values   map[string]float64 // fiscalYear -> value
	IsGAAP   bool
	Category string // "revenue", "expense", "profit"
}

type BalanceSheetItem struct {
	Label    string
	Key      string
	Values   map[string]float64
	Category string // "asset", "liability", "equity"
}

type CashFlowItem struct {
	Label    string
	Key      string
	Values   map[string]float64
	Category string // "operating", "investing", "financing"
}

type KeyMetric struct {
	Label  string
	Key    string
	Values map[string]float64
	Unit   string // "ratio", "percentage", "currency", "multiple"
}

type SearchResult struct {
	Symbol   string
	Name     string
	Market   string
	Type     string // "stock", "etf", "index"
	Currency string
}

type FetchParams struct {
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	Fields    []string
}

type DataSet struct {
	Columns []string
	Rows    []map[string]interface{}
}

type DataHandler func(data *DataSet)

type DataProvider interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
	Fetch(ctx context.Context, sourceID string, params FetchParams) (*DataSet, error)
	Subscribe(ctx context.Context, sourceID string, handler DataHandler) error
	Close() error
}

type FinancialProvider interface {
	DataProvider
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetFinancials(ctx context.Context, symbol string, period Period) (*FinancialStatements, error)
	ListMarkets() []Market
}

type ProviderConfig struct {
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
	CacheTTL time.Duration
}
