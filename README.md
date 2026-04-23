# imans-cli

Official CLI for the Imans public API.

## Goals

- secure token-backed profiles
- human-readable terminal output by default
- `--json` for scripts and AI agents
- stable command names that mirror the public API

## Current Scope

- `version`
- `auth add|test|remove`
- `profile list|show|use`
- `workspace get`
- `products list|get`
- `product-variants list`
- `sales-orders list|get`
- `sales-order-items list`
- `sales-order-classifications list|get`

## Build

```bash
make build
```

## Test

```bash
make test
```

## Auth Examples

```bash
imans auth add --profile acme-prod --token-env IMANS_TOKEN
imans auth test --profile acme-prod
imans profile use acme-prod
```

## Agent-Friendly Examples

```bash
imans products list --json
imans sales-orders list --order-date-from 2026-04-01 --json
imans workspace get --json
```

## Secret Storage

- macOS and Windows use OS-backed credential storage through the keyring backend.
- Linux fails closed when no supported secure keyring is available.
- A development-only plain file backend can be enabled with `IMANS_INSECURE_FILE_SECRETS=1`.

## Schema Refresh

```bash
./scripts/refresh-schema.sh
```
