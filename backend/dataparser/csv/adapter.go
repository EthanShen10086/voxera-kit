package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/EthanShen10086/voxera-kit/dataparser"
)

var datePattern = regexp.MustCompile(
	`^\d{4}[-/]\d{1,2}[-/]\d{1,2}$|^\d{1,2}[-/]\d{1,2}[-/]\d{2,4}$`,
)

type CSVAdapter struct {
	Config dataparser.ParserConfig
}

func New(cfg ...dataparser.ParserConfig) *CSVAdapter {
	a := &CSVAdapter{}
	if len(cfg) > 0 {
		a.Config = cfg[0]
	}
	return a
}

func (a *CSVAdapter) Parse(
	ctx context.Context,
	input io.Reader,
	format dataparser.Format,
	opts ...dataparser.ParseOptions,
) (*dataparser.ParsedDocument, error) {
	if format != dataparser.FormatCSV {
		return nil, fmt.Errorf("csv adapter: unsupported format %q", format)
	}

	opt := dataparser.ParseOptions{Delimiter: ',', Encoding: "utf-8"}
	if len(opts) > 0 {
		opt = mergeOptions(opt, opts[0])
	}

	reader := csv.NewReader(input)
	reader.Comma = opt.Delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv adapter: read error: %w", err)
	}

	if len(records) == 0 {
		return &dataparser.ParsedDocument{
			Format:   dataparser.FormatCSV,
			Tables:   []dataparser.Table{},
			Metadata: map[string]string{"rows": "0", "cols": "0"},
			Pages:    1,
		}, nil
	}

	table := buildTable(records, opt.HasHeader)

	return &dataparser.ParsedDocument{
		Format: dataparser.FormatCSV,
		Tables: []dataparser.Table{table},
		Metadata: map[string]string{
			"rows":      strconv.Itoa(table.RowCount),
			"cols":      strconv.Itoa(table.ColCount),
			"hasHeader": strconv.FormatBool(opt.HasHeader),
			"delimiter": string(opt.Delimiter),
		},
		Pages: 1,
	}, nil
}

func (a *CSVAdapter) ExtractTables(
	ctx context.Context,
	input io.Reader,
	format dataparser.Format,
) ([]dataparser.Table, error) {
	doc, err := a.Parse(ctx, input, format, dataparser.ParseOptions{HasHeader: true})
	if err != nil {
		return nil, err
	}
	return doc.Tables, nil
}

func (a *CSVAdapter) SupportedFormats() []dataparser.Format {
	return []dataparser.Format{dataparser.FormatCSV}
}

func buildTable(records [][]string, hasHeader bool) dataparser.Table {
	colCount := 0
	for _, row := range records {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	var headers []string
	startRow := 0

	if hasHeader && len(records) > 0 {
		headers = make([]string, colCount)
		for i, v := range records[0] {
			headers[i] = strings.TrimSpace(v)
		}
		startRow = 1
	}

	rows := make([][]dataparser.Cell, 0, len(records)-startRow)
	for ri := startRow; ri < len(records); ri++ {
		row := make([]dataparser.Cell, colCount)
		for ci := 0; ci < colCount; ci++ {
			val := ""
			if ci < len(records[ri]) {
				val = strings.TrimSpace(records[ri][ci])
			}
			row[ci] = dataparser.Cell{
				Row:  ri - startRow,
				Col:  ci,
				Value: val,
				Type: inferCellType(val),
			}
		}
		rows = append(rows, row)
	}

	return dataparser.Table{
		Name:     "Sheet1",
		Headers:  headers,
		Rows:     rows,
		RowCount: len(rows),
		ColCount: colCount,
	}
}

func inferCellType(value string) dataparser.CellType {
	if value == "" {
		return dataparser.CellTypeEmpty
	}

	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" {
		return dataparser.CellTypeBoolean
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return dataparser.CellTypeNumber
	}

	if datePattern.MatchString(value) {
		return dataparser.CellTypeDate
	}

	return dataparser.CellTypeString
}

func mergeOptions(base, override dataparser.ParseOptions) dataparser.ParseOptions {
	if override.Delimiter != 0 {
		base.Delimiter = override.Delimiter
	}
	if override.Encoding != "" {
		base.Encoding = override.Encoding
	}
	base.HasHeader = override.HasHeader
	base.Pages = override.Pages
	base.SheetNames = override.SheetNames
	return base
}
