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

## Authenticating

To authenticate you need to supply a Bunny.net [API Key](https://dash.bunny.net/account/settings).

## Debugging

When global debug mode is enabled, this module logs additional information.
