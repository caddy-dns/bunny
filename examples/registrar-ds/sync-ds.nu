#!/usr/bin/env nu
# Sync the DNSSEC DS record(s) for a zone to the registrar, using the CDS
# records Bunny.net publishes in the zone as the source of truth (RFC 7344).
# Nushell version of sync-ds.sh - see that file for details. Run from cron:
#
#   0 6 * * * DONDOMINIO_APIUSER=... DONDOMINIO_APIPASSWD=... nu /path/to/sync-ds.nu example.es
#
# This example targets the DonDominio/MrDomain API; adapt `api` for other
# registrars. Requires: dig.

def api [path: string, params: record] {
    http post --content-type application/x-www-form-urlencoded https://simple-api.dondominio.net/($path)/ ({
        apiuser: $env.DONDOMINIO_APIUSER
        apipasswd: $env.DONDOMINIO_APIPASSWD
    } | merge $params)
}

def main [domain: string] {
    # Ask an authoritative nameserver so the answer does not depend on caches.
    let ns = ^dig +short NS $domain | lines | first
    let want = ^dig +short CDS $domain $"@($ns)" | lines | each {|line|
        let parts = $line | split row ' '
        {
            keytag: ($parts.0 | into int)
            algorithm: ($parts.1 | into int)
            digesttype: ($parts.2 | into int)
            digest: ($parts | skip 3 | str join '' | str upcase)
        }
    }
    if ($want | is-empty) {
        print $"no CDS records published for ($domain); nothing to sync"
        return
    }

    let existing = api domain/getdnssec {domain: $domain}
        | get responseData.dnssec.entries?
        | default []
        | update digest {|entry| $entry.digest | str upcase}

    # Publish DS entries present in the zone's CDS but missing at the registrar.
    for ds in $want {
        if ($existing | where keytag == $ds.keytag and digest == $ds.digest | is-empty) {
            api domain/dnsseccreate ({domain: $domain} | merge $ds) | ignore
            print $"published DS: ($ds.keytag) ($ds.algorithm) ($ds.digesttype) ($ds.digest)"
        }
    }

    # Remove registrar DS entries no longer backed by a CDS record.
    for ds in $existing {
        if ($want | where keytag == $ds.keytag and digest == $ds.digest | is-empty) {
            api domain/dnssecdelete ({domain: $domain} | merge ($ds | select keytag algorithm digesttype digest)) | ignore
            print $"removed stale DS: ($ds.keytag) ($ds.algorithm) ($ds.digesttype) ($ds.digest)"
        }
    }
}
