package aliyun_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/asr"
	"github.com/EthanShen10086/voxera-kit/asr/aliyun"
)

func TestAliyunStub(t *testing.T) {
	a := aliyun.New(asr.Config{Endpoint: "https://nls.example"})
	ctx := context.Background()
	segments, err := a.Recognize(ctx, "http://audio", nil)
	if err != nil || segments != nil {
		t.Fatalf("Recognize: %v err=%v", segments, err)
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
