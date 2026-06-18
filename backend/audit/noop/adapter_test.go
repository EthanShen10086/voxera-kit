package noop_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/audit"
	"github.com/EthanShen10086/voxera-kit/audit/noop"
)

func TestNoopAudit(t *testing.T) {
	a := noop.NewAdapter()
	ctx := context.Background()
	if err := a.Write(ctx, audit.Entry{Action: "login"}); err != nil {
		t.Fatal(err)
	}
	if err := a.WriteBatch(ctx, []audit.Entry{{Action: "logout"}}); err != nil {
		t.Fatal(err)
	}
}
