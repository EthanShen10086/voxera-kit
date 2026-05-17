package stub

import (
	"context"
	"fmt"
	"strings"
	"time"

	dp "github.com/EthanShen10086/voxera-kit/dataprovider"
)

var mockStocks = map[string]dp.SearchResult{
	"AAPL":      {Symbol: "AAPL", Name: "Apple Inc.", Market: "US", Type: "stock", Currency: "USD"},
	"MSFT":      {Symbol: "MSFT", Name: "Microsoft Corporation", Market: "US", Type: "stock", Currency: "USD"},
	"GOOGL":     {Symbol: "GOOGL", Name: "Alphabet Inc.", Market: "US", Type: "stock", Currency: "USD"},
	"AMZN":      {Symbol: "AMZN", Name: "Amazon.com Inc.", Market: "US", Type: "stock", Currency: "USD"},
	"00700.HK":  {Symbol: "00700.HK", Name: "Tencent Holdings Ltd.", Market: "HK", Type: "stock", Currency: "HKD"},
	"09988.HK":  {Symbol: "09988.HK", Name: "Alibaba Group Holding Ltd.", Market: "HK", Type: "stock", Currency: "HKD"},
	"600519.SH": {Symbol: "600519.SH", Name: "贵州茅台", Market: "SH", Type: "stock", Currency: "CNY"},
	"000858.SZ": {Symbol: "000858.SZ", Name: "五粮液", Market: "SZ", Type: "stock", Currency: "CNY"},
}

var mockQuotes = map[string]dp.Quote{
	"AAPL": {
		Symbol: "AAPL", Name: "Apple Inc.", Market: "US",
		Price: 198.50, Change: 2.35, ChangePct: 1.20,
		Volume: 54_320_000, MarketCap: 3_080_000_000_000,
		PE: 32.5, PB: 48.2,
	},
	"MSFT": {
		Symbol: "MSFT", Name: "Microsoft Corporation", Market: "US",
		Price: 420.80, Change: -1.20, ChangePct: -0.28,
		Volume: 22_150_000, MarketCap: 3_130_000_000_000,
		PE: 36.8, PB: 12.9,
	},
	"GOOGL": {
		Symbol: "GOOGL", Name: "Alphabet Inc.", Market: "US",
		Price: 175.30, Change: 3.10, ChangePct: 1.80,
		Volume: 28_900_000, MarketCap: 2_180_000_000_000,
		PE: 25.4, PB: 7.1,
	},
	"00700.HK": {
		Symbol: "00700.HK", Name: "Tencent Holdings Ltd.", Market: "HK",
		Price: 388.60, Change: 5.80, ChangePct: 1.52,
		Volume: 15_600_000, MarketCap: 3_720_000_000_000,
		PE: 22.1, PB: 5.3,
	},
	"600519.SH": {
		Symbol: "600519.SH", Name: "贵州茅台", Market: "SH",
		Price: 1688.00, Change: -12.50, ChangePct: -0.74,
		Volume: 3_200_000, MarketCap: 2_120_000_000_000,
		PE: 28.6, PB: 8.9,
	},
}

var markets = []dp.Market{
	{Code: "US", Name: "United States", Currency: "USD", Timezone: "America/New_York"},
	{Code: "HK", Name: "Hong Kong", Currency: "HKD", Timezone: "Asia/Hong_Kong"},
	{Code: "SH", Name: "Shanghai", Currency: "CNY", Timezone: "Asia/Shanghai"},
	{Code: "SZ", Name: "Shenzhen", Currency: "CNY", Timezone: "Asia/Shanghai"},
}

type StubAdapter struct{}

func New() *StubAdapter {
	return &StubAdapter{}
}

func (s *StubAdapter) Search(_ context.Context, query string) ([]dp.SearchResult, error) {
	query = strings.ToUpper(query)
	var results []dp.SearchResult
	for symbol, sr := range mockStocks {
		if strings.Contains(strings.ToUpper(symbol), query) ||
			strings.Contains(strings.ToUpper(sr.Name), query) {
			results = append(results, sr)
		}
	}
	return results, nil
}

func (s *StubAdapter) GetQuote(_ context.Context, symbol string) (*dp.Quote, error) {
	q, ok := mockQuotes[symbol]
	if !ok {
		return nil, fmt.Errorf("stub: quote not found for %s", symbol)
	}
	q.Timestamp = time.Now()
	return &q, nil
}

func (s *StubAdapter) GetFinancials(_ context.Context, symbol string, period dp.Period) (*dp.FinancialStatements, error) {
	if _, ok := mockStocks[symbol]; !ok {
		return nil, fmt.Errorf("stub: financials not found for %s", symbol)
	}

	years := []string{"2023", "2022", "2021"}

	fs := &dp.FinancialStatements{
		Symbol:      symbol,
		Period:      period,
		FiscalYears: years,
		IncomeStatement: []dp.IncomeStatementItem{
			{
				Label: "Total Revenue", Key: "totalRevenue", Category: "revenue", IsGAAP: true,
				Values: map[string]float64{"2023": 383_285_000_000, "2022": 394_328_000_000, "2021": 365_817_000_000},
			},
			{
				Label: "Cost of Revenue", Key: "costOfRevenue", Category: "expense", IsGAAP: true,
				Values: map[string]float64{"2023": 214_137_000_000, "2022": 223_546_000_000, "2021": 212_981_000_000},
			},
			{
				Label: "Gross Profit", Key: "grossProfit", Category: "profit", IsGAAP: true,
				Values: map[string]float64{"2023": 169_148_000_000, "2022": 170_782_000_000, "2021": 152_836_000_000},
			},
			{
				Label: "Operating Expenses", Key: "operatingExpenses", Category: "expense", IsGAAP: true,
				Values: map[string]float64{"2023": 54_847_000_000, "2022": 51_334_000_000, "2021": 43_887_000_000},
			},
			{
				Label: "Operating Income", Key: "operatingIncome", Category: "profit", IsGAAP: true,
				Values: map[string]float64{"2023": 114_301_000_000, "2022": 119_437_000_000, "2021": 108_949_000_000},
			},
			{
				Label: "Net Income", Key: "netIncome", Category: "profit", IsGAAP: true,
				Values: map[string]float64{"2023": 96_995_000_000, "2022": 99_803_000_000, "2021": 94_680_000_000},
			},
		},
		BalanceSheet: []dp.BalanceSheetItem{
			{
				Label: "Total Assets", Key: "totalAssets", Category: "asset",
				Values: map[string]float64{"2023": 352_583_000_000, "2022": 352_755_000_000, "2021": 351_002_000_000},
			},
			{
				Label: "Cash & Equivalents", Key: "cashAndEquivalents", Category: "asset",
				Values: map[string]float64{"2023": 29_965_000_000, "2022": 23_646_000_000, "2021": 34_940_000_000},
			},
			{
				Label: "Total Current Assets", Key: "totalCurrentAssets", Category: "asset",
				Values: map[string]float64{"2023": 143_566_000_000, "2022": 135_405_000_000, "2021": 134_836_000_000},
			},
			{
				Label: "Total Liabilities", Key: "totalLiabilities", Category: "liability",
				Values: map[string]float64{"2023": 290_437_000_000, "2022": 302_083_000_000, "2021": 287_912_000_000},
			},
			{
				Label: "Total Current Liabilities", Key: "totalCurrentLiabilities", Category: "liability",
				Values: map[string]float64{"2023": 145_308_000_000, "2022": 153_982_000_000, "2021": 125_481_000_000},
			},
			{
				Label: "Total Equity", Key: "totalEquity", Category: "equity",
				Values: map[string]float64{"2023": 62_146_000_000, "2022": 50_672_000_000, "2021": 63_090_000_000},
			},
		},
		CashFlow: []dp.CashFlowItem{
			{
				Label: "Operating Cash Flow", Key: "operatingCashFlow", Category: "operating",
				Values: map[string]float64{"2023": 110_543_000_000, "2022": 122_151_000_000, "2021": 104_038_000_000},
			},
			{
				Label: "Capital Expenditures", Key: "capitalExpenditures", Category: "investing",
				Values: map[string]float64{"2023": -10_959_000_000, "2022": -10_708_000_000, "2021": -11_085_000_000},
			},
			{
				Label: "Free Cash Flow", Key: "freeCashFlow", Category: "operating",
				Values: map[string]float64{"2023": 99_584_000_000, "2022": 111_443_000_000, "2021": 92_953_000_000},
			},
			{
				Label: "Dividends Paid", Key: "dividendsPaid", Category: "financing",
				Values: map[string]float64{"2023": -15_025_000_000, "2022": -14_841_000_000, "2021": -14_467_000_000},
			},
			{
				Label: "Share Buybacks", Key: "shareBuybacks", Category: "financing",
				Values: map[string]float64{"2023": -77_550_000_000, "2022": -89_402_000_000, "2021": -85_500_000_000},
			},
		},
		KeyMetrics: []dp.KeyMetric{
			{
				Label: "Gross Margin", Key: "grossMargin", Unit: "percentage",
				Values: map[string]float64{"2023": 44.13, "2022": 43.31, "2021": 41.78},
			},
			{
				Label: "Operating Margin", Key: "operatingMargin", Unit: "percentage",
				Values: map[string]float64{"2023": 29.82, "2022": 30.29, "2021": 29.78},
			},
			{
				Label: "Net Margin", Key: "netMargin", Unit: "percentage",
				Values: map[string]float64{"2023": 25.31, "2022": 25.31, "2021": 25.88},
			},
			{
				Label: "Return on Equity", Key: "roe", Unit: "percentage",
				Values: map[string]float64{"2023": 156.08, "2022": 196.96, "2021": 150.07},
			},
			{
				Label: "Debt to Equity", Key: "debtToEquity", Unit: "ratio",
				Values: map[string]float64{"2023": 1.79, "2022": 2.37, "2021": 1.98},
			},
			{
				Label: "EPS", Key: "eps", Unit: "currency",
				Values: map[string]float64{"2023": 6.13, "2022": 6.11, "2021": 5.61},
			},
			{
				Label: "P/E Ratio", Key: "peRatio", Unit: "multiple",
				Values: map[string]float64{"2023": 32.5, "2022": 24.8, "2021": 30.1},
			},
		},
	}

	return fs, nil
}

func (s *StubAdapter) ListMarkets() []dp.Market {
	return markets
}

func (s *StubAdapter) Fetch(_ context.Context, sourceID string, params dp.FetchParams) (*dp.DataSet, error) {
	ds := &dp.DataSet{
		Columns: []string{"date", "open", "high", "low", "close", "volume"},
		Rows: []map[string]interface{}{
			{"date": "2024-01-02", "open": 185.0, "high": 188.5, "low": 184.2, "close": 187.3, "volume": 45_000_000},
			{"date": "2024-01-03", "open": 187.3, "high": 189.1, "low": 186.0, "close": 186.8, "volume": 38_000_000},
			{"date": "2024-01-04", "open": 186.8, "high": 187.9, "low": 183.5, "close": 184.2, "volume": 42_000_000},
		},
	}
	return ds, nil
}

func (s *StubAdapter) Subscribe(_ context.Context, _ string, _ dp.DataHandler) error {
	return nil
}

func (s *StubAdapter) Close() error {
	return nil
}

var _ dp.FinancialProvider = (*StubAdapter)(nil)
