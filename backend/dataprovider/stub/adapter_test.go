package stub_test

import (
	"context"
	"testing"

	dp "github.com/EthanShen10086/voxera-kit/dataprovider"
	stub "github.com/EthanShen10086/voxera-kit/dataprovider/stub"
)

func TestStubProvider(t *testing.T) {
	a := stub.New()
	ctx := context.Background()

	results, err := a.Search(ctx, "AAPL")
	if err != nil || len(results) == 0 {
		t.Fatalf("Search: %v err=%v", results, err)
	}
	quote, err := a.GetQuote(ctx, "AAPL")
	if err != nil || quote.Symbol != "AAPL" {
		t.Fatalf("GetQuote: %+v err=%v", quote, err)
	}
	if _, err := a.GetQuote(ctx, "UNKNOWN"); err == nil {
		t.Fatal("expected quote error")
	}
	fs, err := a.GetFinancials(ctx, "AAPL", dp.PeriodAnnual)
	if err != nil || fs.Symbol != "AAPL" {
		t.Fatalf("GetFinancials: %+v err=%v", fs, err)
	}
	markets := a.ListMarkets()
	if len(markets) == 0 {
		t.Fatal("expected markets")
	}
	ds, err := a.Fetch(ctx, "AAPL", dp.FetchParams{})
	if err != nil || len(ds.Rows) == 0 {
		t.Fatalf("Fetch: %+v err=%v", ds, err)
	}
	if err := a.Subscribe(ctx, "AAPL", func(_ *dp.DataSet) {}); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
