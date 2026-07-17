package bunny

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/libdns/bunny"
	"github.com/libdns/libdns"
)

// TestLiveDNSSEC exercises the full DNSSEC-enable path against the real
// Bunny API. Guarded by BUNNY_LIVE_KEY and BUNNY_LIVE_ZONE so it never runs
// in CI. Do not point it at a production zone.
func TestLiveDNSSEC(t *testing.T) {
	key := os.Getenv("BUNNY_LIVE_KEY")
	zone := os.Getenv("BUNNY_LIVE_ZONE")
	if key == "" || zone == "" {
		t.Skip("BUNNY_LIVE_KEY or BUNNY_LIVE_ZONE not set")
	}

	p := Provider{Provider: &bunny.Provider{AccessKey: key}, DNSSEC: true}
	if err := p.Provision(caddy.Context{}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rec := libdns.TXT{Name: "_claude-dnssec-test", TTL: 300 * time.Second, Text: "hello"}
	set, err := p.SetRecords(ctx, zone, []libdns.Record{rec})
	if err != nil {
		t.Fatalf("SetRecords: %v", err)
	}
	t.Logf("set records: %v", set)

	// second call must hit the cache (no error, no duplicate enable)
	if _, err := p.SetRecords(ctx, zone, []libdns.Record{rec}); err != nil {
		t.Fatalf("second SetRecords: %v", err)
	}

	if _, err := p.Provider.DeleteRecords(ctx, zone, set); err != nil {
		t.Errorf("cleanup DeleteRecords: %v", err)
	}
}
