package bunny

import (
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/bunny"
)

// Provider lets Caddy read and manipulate DNS records hosted by this DNS provider.
type Provider struct{ *bunny.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.bunny",
		New: func() caddy.Module { return &Provider{new(bunny.Provider)} },
	}
}

// Provision sets up the module. Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	p.Provider.AccessKey = caddy.NewReplacer().ReplaceAll(p.Provider.AccessKey, "")
	return nil
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	bunny [<access_token>] {
//	    access_key <access_token>
//	    debug <true|false>
//	}
//
// Expansion of placeholders in the API token is left to the JSON config caddy.Provisioner (above).
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			p.Provider.AccessKey = d.Val()
		}
		if d.NextArg() {
			if strings.EqualFold(d.Val(), "debug") {
				p.Provider.Debug = true
			} else {
				return d.ArgErr()
			}
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "access_key":
				if d.NextArg() {
					p.Provider.AccessKey = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "debug":
				if d.NextArg() {
					if debug, err := strconv.ParseBool(d.Val()); err == nil {
						p.Provider.Debug = debug
					}
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	if p.Provider.AccessKey == "" {
		return d.Err("missing access key")
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)
