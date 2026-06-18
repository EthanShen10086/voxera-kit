package azure_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/asr"
	"github.com/EthanShen10086/voxera-kit/asr/azure"
)

func TestAzureStub(t *testing.T) {
	a := azure.New(asr.Config{APIKey: "key"})
	ctx := context.Background()
	if _, err := a.Recognize(ctx, "http://audio", nil); err != nil {
		t.Fatal(err)
	}
	ch, err := a.RecognizeStream(ctx, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
