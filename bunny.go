package bunny

import (
	"context"
	"strings"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/bunny"
	"github.com/libdns/libdns"
	"go.uber.org/zap"
)

// Provider lets Caddy read and manipulate DNS records hosted by this DNS provider.
type Provider struct {
	*bunny.Provider

	// DNSSEC enables DNSSEC on each zone this provider writes records to.
	// Bunny.net manages keys and signing automatically; if the domain's
	// registrar is not Bunny.net, the DS record logged on enablement must
	// be published at the registrar manually.
	DNSSEC bool `json:"dnssec,omitempty"`

	logger *zap.Logger
	// pointer so the zero Provider stays copyable (Caddy copies modules by value)
	dnssec *dnssecState
}

type dnssecState struct {
	mu sync.Mutex
	on map[string]bool
}

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.bunny",
		New: func() caddy.Module { return &Provider{Provider: new(bunny.Provider)} },
	}
}

// Provision sets up the module. Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	p.Provider.AccessKey = caddy.NewReplacer().ReplaceAll(p.Provider.AccessKey, "")
	p.logger = ctx.Logger(p)
	p.dnssec = &dnssecState{on: make(map[string]bool)}

	// Set up the logger. This will automatically enable debug in the provider.
	p.Logger = func(msg string, records []libdns.Record) {
		if len(records) > 1 {
			ctx.Logger(p).Debug(msg, zap.Any("records", records))
		} else if len(records) > 0 {
			ctx.Logger(p).Debug(msg, zap.Any("record", records[0]))
		} else {
			ctx.Logger(p).Debug(msg)
		}
	}
	return nil
}

// AppendRecords adds records to the zone, first ensuring DNSSEC is enabled
// on it if configured. Implements libdns.RecordAppender.
func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.ensureDNSSEC(ctx, zone)
	return p.Provider.AppendRecords(ctx, zone, recs)
}

// SetRecords sets records in the zone, first ensuring DNSSEC is enabled
// on it if configured. Implements libdns.RecordSetter.
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.ensureDNSSEC(ctx, zone)
	return p.Provider.SetRecords(ctx, zone, recs)
}

// ensureDNSSEC enables DNSSEC on the zone if it isn't already, delegating to
// the libdns provider. It is best effort: failures are logged but never block
// the record operation, so certificate issuance keeps working even if DNSSEC
// can't be enabled.
func (p *Provider) ensureDNSSEC(ctx context.Context, zone string) {
	if !p.DNSSEC || p.dnssec == nil {
		return
	}
	zone = strings.ToLower(strings.TrimSuffix(zone, "."))

	p.dnssec.mu.Lock()
	defer p.dnssec.mu.Unlock()
	if p.dnssec.on[zone] {
		return
	}
	info, err := p.Provider.EnableDNSSEC(ctx, zone)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("could not enable DNSSEC; continuing without it",
				zap.String("zone", zone), zap.Error(err))
		}
		return
	}
	p.dnssec.on[zone] = true

	if p.logger == nil {
		return
	}
	p.logger.Info("DNSSEC enabled", zap.String("zone", zone))
	if info.DSRecord != "" && !info.DSConfigured {
		p.logger.Warn("DS record is not configured at the registrar; DNSSEC validation will not work until you publish it there",
			zap.String("zone", zone), zap.String("ds_record", info.DSRecord))
	}
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	bunny [<access_token>] {
//	    access_key <access_token>
//	    dnssec
//	}
//
// Expansion of placeholders in the API token is left to the JSON config caddy.Provisioner (above).
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			p.Provider.AccessKey = d.Val()
		}
		if d.NextArg() {
			return d.ArgErr()
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
			case "dnssec":
				if d.NextArg() {
					return d.ArgErr()
				}
				p.DNSSEC = true
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
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
)
