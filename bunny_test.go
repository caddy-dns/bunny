package bunny

import (
	"fmt"
	"os"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/bunny"
)

type testCase struct {
	config          string
	setenvAccessKey string
	setenvDebug     string
	expectAccessKey string
	expectDebug     bool
}

func TestUnmarshalCaddyFile(t *testing.T) {
	tests := []testCase{
		{
			config: `bunny {
				access_key A123
				debug {env.BUNNY_DEBUG}
			}`,
			setenvDebug:     "1",
			expectAccessKey: "A123",
			expectDebug:     true,
		}, {
			config: `bunny {
				access_key {env.BUNNY_ACCESS_KEY}
				debug 1
			}`,
			setenvAccessKey: "321",
			expectAccessKey: "321",
			expectDebug:     true,
		}, {
			config: `bunny {
				access_key A123
			}`,
			expectAccessKey: "A123",
			expectDebug:     false,
		}, {
			config:          "bunny A123",
			expectAccessKey: "A123",
			expectDebug:     false,
		}, {
			config:          "bunny A123 DeBug",
			expectAccessKey: "A123",
			expectDebug:     true,
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			dispenser := caddyfile.NewTestDispenser(tc.config)
			p := Provider{Provider: &bunny.Provider{}}

			err := p.UnmarshalCaddyfile(dispenser)
			if err != nil {
				t.Errorf("UnmarshalCaddyfile failed with %v", err)
				return
			}

			os.Setenv("BUNNY_ACCESS_KEY", tc.setenvAccessKey)
			os.Setenv("BUNNY_DEBUG", tc.setenvDebug)

			err = p.Provision(caddy.Context{})
			if err != nil {
				t.Errorf("Provision failed with %v", err)
				return
			}

			if tc.expectAccessKey != p.Provider.AccessKey {
				t.Errorf("Expected AccessKey to be '%s' but got '%s'", tc.expectAccessKey, p.Provider.AccessKey)
			}

			if tc.expectDebug != p.Provider.Debug {
				t.Errorf("Expected Debug to be '%t' but got '%t'", tc.expectDebug, p.Provider.Debug)
			}
		})
	}
}
