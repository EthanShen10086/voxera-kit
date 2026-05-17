package dataparser

import (
	"context"
	"io"
)

type Format string

const (
	FormatPDF  Format = "pdf"
	FormatCSV  Format = "csv"
	FormatXLSX Format = "xlsx"
	FormatHTML Format = "html"
	FormatJSON Format = "json"
)

type Cell struct {
	Row    int
	Col    int
	Value  string
	Type   CellType
	Merged bool
	Span   *CellSpan
}

type CellType string

const (
	CellTypeString  CellType = "string"
	CellTypeNumber  CellType = "number"
	CellTypeDate    CellType = "date"
	CellTypeBoolean CellType = "boolean"
	CellTypeEmpty   CellType = "empty"
)

type CellSpan struct {
	RowSpan int
	ColSpan int
}

type Table struct {
	Name     string
	Headers  []string
	Rows     [][]Cell
	RowCount int
	ColCount int
}

type ParsedDocument struct {
	Format   Format
	Tables   []Table
	Metadata map[string]string
	Pages    int
	Errors   []string
}

type ParseOptions struct {
	Pages      []int    // specific pages to parse (PDF only)
	SheetNames []string // specific sheets (XLSX only)
	HasHeader  bool
	Delimiter  rune   // CSV delimiter, default ','
	Encoding   string // default "utf-8"
}

type DocumentParser interface {
	Parse(ctx context.Context, input io.Reader, format Format, opts ...ParseOptions) (*ParsedDocument, error)
	ExtractTables(ctx context.Context, input io.Reader, format Format) ([]Table, error)
	SupportedFormats() []Format
}

type ParserConfig struct {
	MaxFileSize int64
	TempDir     string
	Timeout     int // seconds
}
