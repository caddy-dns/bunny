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
	input  string
	setenv string
	expect string
}

func TestUnmarshalCaddyFile(t *testing.T) {
	tests := []testCase{
		{
			input: `bunny {
				access_key A123
			}`,
			expect: "A123",
		}, {
			input: `bunny {
				access_key {env.BUNNY_ACCESS_KEY}
			}`,
			setenv: "321",
			expect: "321",
		}, {
			input:  `bunny 123`,
			expect: "123",
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			dispenser := caddyfile.NewTestDispenser(tc.input)
			p := Provider{Provider: &bunny.Provider{}}

			err := p.UnmarshalCaddyfile(dispenser)
			if err != nil {
				t.Errorf("UnmarshalCaddyfile failed with %v", err)
				return
			}

			os.Setenv("BUNNY_ACCESS_KEY", tc.setenv)

			err = p.Provision(caddy.Context{})
			if err != nil {
				t.Errorf("Provision failed with %v", err)
				return
			}

			if tc.expect != p.Provider.AccessKey {
				t.Errorf("Expected AccessKey to be '%s' but got '%s'", tc.expect, p.Provider.AccessKey)
			}
		})
	}
}
