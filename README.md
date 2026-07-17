Bunny.net module for Caddy
===========================

This package contains a DNS provider module for [Caddy](https://github.com/caddyserver/caddy). It can be used to manage DNS records with Bunny.net accounts.

## Caddy module name

```
dns.providers.bunny
```

## Config examples

To use this module for the ACME DNS challenge, [configure the ACME issuer in your Caddy JSON](https://caddyserver.com/docs/json/apps/tls/automation/policies/issuer/acme/) like so:

```json
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "bunny",
				"access_key": "{env.BUNNY_API_KEY}",
			}
		}
	}
}
```

or with the Caddyfile:

```
tls {
	dns bunny {env.BUNNY_API_KEY}
}
```

You can replace `{env.BUNNY_API_KEY}` with the actual auth token if you prefer to put it directly in your config instead of an environment variable.

## DNSSEC

Bunny.net supports automatic DNSSEC (keys and zone signing are fully managed for you). Add the `dnssec` option and this module enables it on each zone it manages records for:

```
tls {
	dns bunny {
		access_key {env.BUNNY_API_KEY}
		dnssec
	}
}
```

or in JSON: `"dnssec": true` next to `"access_key"`.

If your domain is not registered with Bunny.net, you must publish the DS record at your registrar yourself — the module logs the exact DS record (also shown in the Bunny dashboard) and warns while it is not configured. Enabling DNSSEC is best effort: if the API call fails, a warning is logged and certificate issuance proceeds normally. Removing the option does not disable DNSSEC on the zone; disable it in the Bunny dashboard if you need to (and only after removing the DS record at the registrar).

### Publishing the DS record at your registrar

Bunny signs the zone, but DNSSEC only becomes active once the DS record is published in the parent zone, and that happens at your **registrar**, not at Bunny. There are two ways to do it:

**Manually.** Take the DS values from the module's log (or the Bunny dashboard → DNS Zone → DNSSEC) and enter them in your registrar's DNSSEC panel: key tag, algorithm, digest type and digest. Then verify:

```console
$ dig +short DS example.com        # should show your key tag / algorithm / digest
$ delv example.com                 # should report "fully validated"
```

**Automatically.** Bunny publishes `CDS`/`CDNSKEY` records (RFC 7344) in the zone. Registries that scan for them (e.g. .ch, .cz, .dk) install the DS with no action on your part. For registries that don't, a small cron job can read the zone's CDS records and sync them through your registrar's API — which also keeps the DS current across key rotations. See [`examples/registrar-ds/`](examples/registrar-ds/) for a working example against the [DonDominio/MrDomain API](https://dondominio.dev/en/api/docs/api/) in both Bash ([`sync-ds.sh`](examples/registrar-ds/sync-ds.sh)) and Nushell ([`sync-ds.nu`](examples/registrar-ds/sync-ds.nu)):

```console
$ DOMAIN=example.es DONDOMINIO_APIUSER=... DONDOMINIO_APIPASSWD=... ./sync-ds.sh
published DS: 31860 13 2 7CB2C509A1BE091D65E8949E700A69...
```

Whichever way you publish it: if you ever disable DNSSEC or move your DNS away from Bunny, **remove the DS record at the registrar first** and wait out its TTL — otherwise validating resolvers will refuse to resolve the domain.

## Authenticating

To authenticate you need to supply a Bunny.net [API Key](https://dash.bunny.net/account/settings).

## Debugging

When global debug mode is enabled, this module logs additional information.
