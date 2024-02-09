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
				"access_key": "{env.BUNNY_API_KEY}"
			}
		}
	}
}
```

or with the Caddyfile:

```
tls {
	acme_dns bunny {env.BUNNY_API_KEY}
}
```

You can replace `{env.BUNNY_API_KEY}` with the actual auth token if you prefer to put it directly in your config instead of an environment variable.

## Authenticating

To authenticate you need to supply a Bunny.net [API Key](https://dash.bunny.net/account/settings).