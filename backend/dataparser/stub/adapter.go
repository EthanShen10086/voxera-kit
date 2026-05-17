package stub

import (
	"context"
	"io"

	"github.com/EthanShen10086/voxera-kit/dataparser"
)

type StubAdapter struct{}

func New() *StubAdapter {
	return &StubAdapter{}
}

func (a *StubAdapter) Parse(
	ctx context.Context,
	input io.Reader,
	format dataparser.Format,
	opts ...dataparser.ParseOptions,
) (*dataparser.ParsedDocument, error) {
	return &dataparser.ParsedDocument{
		Format: format,
		Tables: financialTables(),
		Metadata: map[string]string{
			"source":   "stub",
			"company":  "Acme Corp",
			"period":   "Q4 2024",
			"currency": "USD",
		},
		Pages: 1,
	}, nil
}

func (a *StubAdapter) ExtractTables(
	ctx context.Context,
	input io.Reader,
	format dataparser.Format,
) ([]dataparser.Table, error) {
	return financialTables(), nil
}

func (a *StubAdapter) SupportedFormats() []dataparser.Format {
	return []dataparser.Format{
		dataparser.FormatPDF,
		dataparser.FormatCSV,
		dataparser.FormatXLSX,
		dataparser.FormatHTML,
		dataparser.FormatJSON,
	}
}

func financialTables() []dataparser.Table {
	return []dataparser.Table{
		{
			Name:    "Income Statement",
			Headers: []string{"Item", "Q3 2024", "Q4 2024", "YoY Change"},
			Rows: [][]dataparser.Cell{
				{
					{Row: 0, Col: 0, Value: "Revenue", Type: dataparser.CellTypeString},
					{Row: 0, Col: 1, Value: "12500000", Type: dataparser.CellTypeNumber},
					{Row: 0, Col: 2, Value: "14200000", Type: dataparser.CellTypeNumber},
					{Row: 0, Col: 3, Value: "13.6%", Type: dataparser.CellTypeString},
				},
				{
					{Row: 1, Col: 0, Value: "Cost of Goods Sold", Type: dataparser.CellTypeString},
					{Row: 1, Col: 1, Value: "7500000", Type: dataparser.CellTypeNumber},
					{Row: 1, Col: 2, Value: "8100000", Type: dataparser.CellTypeNumber},
					{Row: 1, Col: 3, Value: "8.0%", Type: dataparser.CellTypeString},
				},
				{
					{Row: 2, Col: 0, Value: "Gross Profit", Type: dataparser.CellTypeString},
					{Row: 2, Col: 1, Value: "5000000", Type: dataparser.CellTypeNumber},
					{Row: 2, Col: 2, Value: "6100000", Type: dataparser.CellTypeNumber},
					{Row: 2, Col: 3, Value: "22.0%", Type: dataparser.CellTypeString},
				},
				{
					{Row: 3, Col: 0, Value: "Operating Expenses", Type: dataparser.CellTypeString},
					{Row: 3, Col: 1, Value: "3200000", Type: dataparser.CellTypeNumber},
					{Row: 3, Col: 2, Value: "3400000", Type: dataparser.CellTypeNumber},
					{Row: 3, Col: 3, Value: "6.3%", Type: dataparser.CellTypeString},
				},
				{
					{Row: 4, Col: 0, Value: "Net Income", Type: dataparser.CellTypeString},
					{Row: 4, Col: 1, Value: "1800000", Type: dataparser.CellTypeNumber},
					{Row: 4, Col: 2, Value: "2700000", Type: dataparser.CellTypeNumber},
					{Row: 4, Col: 3, Value: "50.0%", Type: dataparser.CellTypeString},
				},
			},
			RowCount: 5,
			ColCount: 4,
		},
		{
			Name:    "Balance Sheet",
			Headers: []string{"Item", "2023-12-31", "2024-12-31"},
			Rows: [][]dataparser.Cell{
				{
					{Row: 0, Col: 0, Value: "Total Assets", Type: dataparser.CellTypeString},
					{Row: 0, Col: 1, Value: "45000000", Type: dataparser.CellTypeNumber},
					{Row: 0, Col: 2, Value: "52000000", Type: dataparser.CellTypeNumber},
				},
				{
					{Row: 1, Col: 0, Value: "Total Liabilities", Type: dataparser.CellTypeString},
					{Row: 1, Col: 1, Value: "18000000", Type: dataparser.CellTypeNumber},
					{Row: 1, Col: 2, Value: "19500000", Type: dataparser.CellTypeNumber},
				},
				{
					{Row: 2, Col: 0, Value: "Shareholders' Equity", Type: dataparser.CellTypeString},
					{Row: 2, Col: 1, Value: "27000000", Type: dataparser.CellTypeNumber},
					{Row: 2, Col: 2, Value: "32500000", Type: dataparser.CellTypeNumber},
				},
			},
			RowCount: 3,
			ColCount: 3,
		},
	}
}
