package bunny

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/bunny"
)

type testCase struct {
	Config    string
	AccessKey string
	Debug     bool
}

func TestUnmarshalCaddyFile(t *testing.T) {
	tests := []testCase{
		{
			Config: `bunny {
				access_key A123
				debug true
			}`,
			AccessKey: "A123",
			Debug:     true,
		}, {
			Config: `bunny {
				access_key 321
				debug 1
			}`,
			AccessKey: "321",
			Debug:     true,
		}, {
			Config: `bunny {
				access_key A123
			}`,
			AccessKey: "A123",
			Debug:     false,
		}, {
			Config:    "bunny A123",
			AccessKey: "A123",
			Debug:     false,
		}, {
			Config:    "bunny A123 DEBUG",
			AccessKey: "A123",
			Debug:     true,
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			dispenser := caddyfile.NewTestDispenser(tc.Config)
			p := Provider{&bunny.Provider{}}
			err := p.UnmarshalCaddyfile(dispenser)

			if err != nil {
				t.Errorf("UnmarshalCaddyfile failed with %v", err)
				return
			}

			if tc.AccessKey != p.Provider.AccessKey {
				t.Errorf("Expected AccessKey to be '%s' but got '%s'", tc.AccessKey, p.Provider.AccessKey)
			}
			if tc.Debug != p.Provider.Debug {
				t.Errorf("Expected Debug to be '%t' but got '%t'", tc.Debug, p.Provider.Debug)
			}
		})
	}
}
