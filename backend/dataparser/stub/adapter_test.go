package stub_test

import (
	"context"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/dataparser"
	"github.com/EthanShen10086/voxera-kit/dataparser/stub"
)

func TestStubParser(t *testing.T) {
	a := stub.New()
	ctx := context.Background()

	doc, err := a.Parse(ctx, strings.NewReader(""), dataparser.FormatPDF)
	if err != nil || len(doc.Tables) != 2 || doc.Metadata["source"] != "stub" {
		t.Fatalf("Parse: %+v err=%v", doc, err)
	}

	tables, err := a.ExtractTables(ctx, nil, dataparser.FormatCSV)
	if err != nil || len(tables) != 2 {
		t.Fatalf("ExtractTables: %v err=%v", tables, err)
	}

	formats := a.SupportedFormats()
	if len(formats) < 4 {
		t.Fatalf("formats = %v", formats)
	}
}
