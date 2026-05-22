// Package dataparser defines the port interfaces and types for parsing
// structured documents (PDF, CSV, XLSX, HTML, JSON) into tabular data.
package dataparser

import (
	"context"
	"io"
)

// Format identifies a supported document format.
type Format string

// Supported document formats.
const (
	FormatPDF  Format = "pdf"
	FormatCSV  Format = "csv"
	FormatXLSX Format = "xlsx"
	FormatHTML Format = "html"
	FormatJSON Format = "json"
)

// Cell represents a single cell in a parsed table.
type Cell struct {
	Row    int
	Col    int
	Value  string
	Type   CellType
	Merged bool
	Span   *CellSpan
}

// CellType classifies the data type of a cell value.
type CellType string

// Cell type constants.
const (
	CellTypeString  CellType = "string"
	CellTypeNumber  CellType = "number"
	CellTypeDate    CellType = "date"
	CellTypeBoolean CellType = "boolean"
	CellTypeEmpty   CellType = "empty"
)

// CellSpan describes how many rows and columns a merged cell covers.
type CellSpan struct {
	RowSpan int
	ColSpan int
}

// Table represents a parsed table with headers and rows.
type Table struct {
	Name     string
	Headers  []string
	Rows     [][]Cell
	RowCount int
	ColCount int
}

// ParsedDocument holds the result of parsing a document.
type ParsedDocument struct {
	Format   Format
	Tables   []Table
	Metadata map[string]string
	Pages    int
	Errors   []string
}

// ParseOptions configures the behavior of a parse operation.
type ParseOptions struct {
	Pages      []int    // specific pages to parse (PDF only)
	SheetNames []string // specific sheets (XLSX only)
	HasHeader  bool
	Delimiter  rune   // CSV delimiter, default ','
	Encoding   string // default "utf-8"
}

// DocumentParser is the interface for parsing structured documents into tables.
type DocumentParser interface {
	Parse(ctx context.Context, input io.Reader, format Format, opts ...ParseOptions) (*ParsedDocument, error)
	ExtractTables(ctx context.Context, input io.Reader, format Format) ([]Table, error)
	SupportedFormats() []Format
}

// ParserConfig holds configuration parameters for a document parser.
type ParserConfig struct {
	MaxFileSize int64
	TempDir     string
	Timeout     int // seconds
}
