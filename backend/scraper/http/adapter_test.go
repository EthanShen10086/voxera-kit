package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EthanShen10086/voxera-kit/scraper"
	scraperhttp "github.com/EthanShen10086/voxera-kit/scraper/http"
	"github.com/EthanShen10086/voxera-kit/scraper/memory"
)

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("page"))
	}))
	defer srv.Close()

	pool := memory.NewProxyPool()
	a := scraperhttp.New(scraperhttp.Config{ProxyPool: pool})
	resp, err := a.Fetch(context.Background(), &scraper.FetchRequest{
		Method: http.MethodGet,
		URL:    srv.URL,
		Headers: map[string]string{"X-Test": "1"},
	})
	if err != nil || resp.StatusCode != http.StatusOK || string(resp.Body) != "page" {
		t.Fatalf("Fetch: %+v err=%v", resp, err)
	}
}
