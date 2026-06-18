package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/audit"
	"github.com/EthanShen10086/voxera-kit/audit/memory"
)

func TestAuditWriteQueryCount(t *testing.T) {
	ctx := context.Background()
	a := memory.NewAdapter()
	ts := time.Now()

	entry := audit.Entry{
		TenantID: "t1", ActorID: "u1", Action: "create",
		ResourceType: "doc", ResourceID: "d1", Timestamp: ts,
	}
	if err := a.Write(ctx, entry); err != nil {
		t.Fatal(err)
	}
	if err := a.WriteBatch(ctx, []audit.Entry{entry, {
		TenantID: "t1", ActorID: "u2", Action: "delete", Timestamp: ts,
	}}); err != nil {
		t.Fatal(err)
	}

	results, err := a.Query(ctx, audit.Filter{TenantID: "t1", Action: "create", Limit: 10})
	if err != nil || len(results) != 2 {
		t.Fatalf("Query: len=%d err=%v", len(results), err)
	}
	count, err := a.Count(ctx, audit.Filter{TenantID: "t1"})
	if err != nil || count != 3 {
		t.Fatalf("Count = %d err=%v", count, err)
	}
}
