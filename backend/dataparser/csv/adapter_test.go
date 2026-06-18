package csv_test

import (
	"context"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/dataparser"
	"github.com/EthanShen10086/voxera-kit/dataparser/csv"
)

func TestParse_EmptyCSV(t *testing.T) {
	a := csv.New()
	doc, err := a.Parse(context.Background(), strings.NewReader(""), dataparser.FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Tables) != 0 || doc.Metadata["rows"] != "0" {
		t.Fatalf("doc = %+v", doc)
	}
}

func TestParse_WithHeaderAndTypes(t *testing.T) {
	input := "name,age,active,joined\nalice,30,true,2024-01-15\nbob,,false,01/02/2023\n"
	a := csv.New()
	doc, err := a.Parse(context.Background(), strings.NewReader(input), dataparser.FormatCSV,
		dataparser.ParseOptions{HasHeader: true})
	if err != nil {
		t.Fatal(err)
	}
	table := doc.Tables[0]
	if table.RowCount != 2 || table.Headers[0] != "name" {
		t.Fatalf("table = %+v", table)
	}
	if table.Rows[0][1].Type != dataparser.CellTypeNumber {
		t.Fatalf("age type = %v", table.Rows[0][1].Type)
	}
	if table.Rows[0][2].Type != dataparser.CellTypeBoolean {
		t.Fatalf("active type = %v", table.Rows[0][2].Type)
	}
	if table.Rows[1][1].Type != dataparser.CellTypeEmpty {
		t.Fatalf("empty cell type = %v", table.Rows[1][1].Type)
	}
}

func TestParse_CustomDelimiter(t *testing.T) {
	a := csv.New()
	doc, err := a.Parse(context.Background(), strings.NewReader("a;b\n1;2\n"), dataparser.FormatCSV,
		dataparser.ParseOptions{Delimiter: ';'})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Tables[0].ColCount != 2 {
		t.Fatalf("cols = %d", doc.Tables[0].ColCount)
	}
}

func TestParse_UnsupportedFormat(t *testing.T) {
	a := csv.New()
	_, err := a.Parse(context.Background(), strings.NewReader("x"), dataparser.Format("pdf"))
	if err == nil {
		t.Fatal("expected format error")
	}
}

func TestExtractTablesAndSupportedFormats(t *testing.T) {
	a := csv.New()
	formats := a.SupportedFormats()
	if len(formats) != 1 || formats[0] != dataparser.FormatCSV {
		t.Fatalf("formats = %v", formats)
	}
	tables, err := a.ExtractTables(context.Background(), strings.NewReader("h\nv\n"), dataparser.FormatCSV)
	if err != nil || len(tables) != 1 {
		t.Fatalf("tables = %v err=%v", tables, err)
	}
}
