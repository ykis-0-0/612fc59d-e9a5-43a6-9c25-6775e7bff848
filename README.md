# realip-zoning
Traefik middleware plugin for IP-based zoning.

Module path `ykis.me/traefik-plugins/zoning-realip`

## What it does
- Resolves request source IPs into configured zones.
- Supports trusted proxy CIDR configuration.
- Attaches zone-specific headers when a request matches.
- Supports a null-zone header set for unmatched requests.

## Why make this?
Because Traefik doesn't support it.

## Where Zone definitions are sourced
- List of URL(s)
- Directory of files (not recursive)
- Inline list of CIDR(s)
