#!/usr/bin/env bash
# Sync the DNSSEC DS record(s) for a zone to the registrar, using the CDS
# records Bunny.net publishes in the zone as the source of truth (RFC 7344).
# Idempotent: missing DS entries are created, stale ones (old keys after a
# rotation) are removed. Run it periodically from cron, e.g.:
#
#   0 6 * * * DOMAIN=example.es DONDOMINIO_APIUSER=... DONDOMINIO_APIPASSWD=... /path/to/sync-ds.sh
#
# This example targets the DonDominio/MrDomain API
# (https://dondominio.dev/en/api/docs/api/); adapt the `api` calls for other
# registrars. Requires: dig, curl, jq.
set -euo pipefail

: "${DOMAIN:?set DOMAIN to the zone, e.g. example.es}"
: "${DONDOMINIO_APIUSER:?set DONDOMINIO_APIUSER}"
: "${DONDOMINIO_APIPASSWD:?set DONDOMINIO_APIPASSWD}"

api() { # api <path> [curl -d args...]
	local path=$1
	shift
	curl -sf "https://simple-api.dondominio.net/$path/" \
		-d apiuser="$DONDOMINIO_APIUSER" -d apipasswd="$DONDOMINIO_APIPASSWD" \
		-d domain="$DOMAIN" "$@"
}

# Ask an authoritative nameserver for the CDS records so the answer does not
# depend on resolver caches. Lines are normalized to:
#   <keytag> <algorithm> <digesttype> <DIGEST>
ns=$(dig +short NS "$DOMAIN" | head -1)
want=$(dig +short CDS "$DOMAIN" "@$ns" |
	awk '{d=""; for (i=4; i<=NF; i++) d=d toupper($i); print $1, $2, $3, d}' | sort)
if [ -z "$want" ]; then
	echo "no CDS records published for $DOMAIN; nothing to sync" >&2
	exit 0
fi

existing=$(api domain/getdnssec |
	jq -r '.responseData.dnssec.entries // [] | .[]
	       | "\(.keytag) \(.algorithm) \(.digesttype) \(.digest | ascii_upcase)"' | sort)

# Publish DS entries present in the zone's CDS but missing at the registrar.
comm -13 <(echo "$existing") <(echo "$want") | while read -r keytag alg dtype digest; do
	[ -n "$keytag" ] || continue
	api domain/dnsseccreate -d keytag="$keytag" -d algorithm="$alg" \
		-d digesttype="$dtype" -d digest="$digest" >/dev/null
	echo "published DS: $keytag $alg $dtype $digest"
done

# Remove registrar DS entries no longer backed by a CDS record.
comm -23 <(echo "$existing") <(echo "$want") | while read -r keytag alg dtype digest; do
	[ -n "$keytag" ] || continue
	api domain/dnssecdelete -d keytag="$keytag" -d algorithm="$alg" \
		-d digesttype="$dtype" -d digest="$digest" >/dev/null
	echo "removed stale DS: $keytag $alg $dtype $digest"
done
