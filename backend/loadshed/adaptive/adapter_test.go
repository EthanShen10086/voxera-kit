package adaptive_test

import (
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/loadshed"
	"github.com/EthanShen10086/voxera-kit/loadshed/adaptive"
)

func TestAllowAndToken(t *testing.T) {
	a := adaptive.New(loadshed.Config{MaxLoad: 0.5, Window: time.Minute})

	tok, err := a.Allow()
	if err != nil {
		t.Fatal(err)
	}
	tok.Done(true)
	tok.Done(true) // idempotent

	if a.Load() != 0 {
		t.Fatalf("load after success = %v", a.Load())
	}

	tok2, err := a.Allow()
	if err != nil {
		t.Fatal(err)
	}
	tok2.Done(false)
	if a.Load() != 0.5 {
		t.Fatalf("load after failure = %v", a.Load())
	}
}

func TestOverload(t *testing.T) {
	a := adaptive.New(loadshed.Config{MaxLoad: 0.0, Window: time.Minute})
	_, err := a.Allow()
	if !errors.Is(err, loadshed.ErrOverloaded) {
		t.Fatalf("err = %v", err)
	}
}
