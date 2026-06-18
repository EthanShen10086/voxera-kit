package dataparser_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/dataparser"
	_ "github.com/EthanShen10086/voxera-kit/dataparser/csv"
)

func TestNewCSVParser(t *testing.T) {
	p := dataparser.NewCSVParser()
	if p == nil {
		t.Fatal("nil parser")
	}
	formats := p.SupportedFormats()
	if len(formats) == 0 {
		t.Fatal("expected supported formats")
	}
	found := false
	for _, f := range formats {
		if f == dataparser.FormatCSV {
			found = true
		}
	}
	if !found {
		t.Fatalf("formats = %v", formats)
	}
}

func TestNewParserPanicsWhenMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown format")
		}
	}()
	dataparser.NewParser("unknown-format")
}

func TestNewParserCSV(t *testing.T) {
	p := dataparser.NewParser(dataparser.FormatCSV)
	if p == nil {
		t.Fatal("nil parser")
	}
}
