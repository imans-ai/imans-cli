# imans-cli

[![CI](https://github.com/imans-ai/imans-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/imans-ai/imans-cli/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/imans-ai/imans-cli)](https://github.com/imans-ai/imans-cli/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/imans-ai/imans-cli)](https://github.com/imans-ai/imans-cli/blob/main/go.mod)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Homebrew](https://img.shields.io/badge/Homebrew-planned-lightgrey?logo=homebrew)](#install)
[![Scoop](https://img.shields.io/badge/Scoop-planned-lightgrey)](#install)
[![Last commit](https://img.shields.io/github/last-commit/imans-ai/imans-cli)](https://github.com/imans-ai/imans-cli/commits/main)

Official CLI for the Imans public API.

`imans-cli` is designed for two modes at the same time:

- humans working in a terminal who want readable output
- scripts and AI agents that need stable commands and clean JSON

It uses saved profiles, secure token storage, explicit flags, and resource-oriented commands that mirror the public API domains.

## Highlights

- secure token-backed profiles
- one active profile with per-command `--profile` override
- text output by default, `--json` for automation
- stable resource-oriented command names
- Linux, macOS, and Windows build targets
- release automation scaffolded with GoReleaser

## Current Command Surface

```text
imans version

imans auth add
imans auth test
imans auth remove

imans profile list
imans profile show
imans profile use

imans workspace get

imans products list
imans products get <id>

imans product-variants list

imans sales-orders list
imans sales-orders get <id>

imans sales-order-items list

imans sales-order-classifications list
imans sales-order-classifications get <id>
```

## Install

The repository is ready to build now. GitHub Releases, Homebrew, and Scoop automation are scaffolded, but if you are working from source today the most direct path is:

```bash
make build
```

That creates a local binary named `imans` in the repository root.

To install it for your current Linux user:

```bash
mkdir -p "$HOME/.local/bin"
cp ./imans "$HOME/.local/bin/imans"
chmod +x "$HOME/.local/bin/imans"
```

If `~/.local/bin` is on your `PATH`, you can then run:

```bash
imans version
```

## Build From Source

Requirements:

- Go 1.22+

Commands:

```bash
make build
make test
```

Useful development commands:

```bash
make fmt
make vet
make schema
```

## Quickstart

The default API base URL is `https://api.imans.ai/`.

### 1. Add a profile

```bash
read -rsp 'Imans token: ' IMANS_TOKEN && printf '\n'
export IMANS_TOKEN

./imans auth add --profile prod-main --token-env IMANS_TOKEN --set-active
```

You can also provide the token with:

- `--token`
- `--token-env <ENV_VAR_NAME>`
- `--token-stdin`
- interactive secure prompt fallback

### 2. Verify the profile

```bash
./imans auth test
./imans profile list
./imans workspace get
```

### 3. Add another workspace alias

```bash
./imans auth add --profile staging --base-url https://staging-api.example.com/ --token-env IMANS_TOKEN
./imans profile use staging
./imans profile show
```

Each token maps to one workspace. The CLI stores a user-defined alias for that workspace and keeps one active profile at a time.

## Output Modes

Default behavior:

- text and tables on stdout for humans
- errors and warnings on stderr

Automation mode:

```bash
./imans workspace get --json
./imans products list --all --json
./imans sales-orders list --order-date-from 2026-04-01 --json
```

Global flags:

- `--json`
- `--profile <name>`
- `--quiet`
- `--debug`
- `--no-color`

### Pagination

List endpoints support:

- `--page`
- `--page-size`
- `--all`

When `--all --json` is used, the CLI returns one combined payload instead of page-by-page JSON.

## Examples

### Profiles and auth

```bash
./imans auth add --profile acme-prod --token-env IMANS_TOKEN --set-active
./imans auth test
./imans profile list
./imans profile use acme-prod
./imans profile show
./imans auth remove acme-prod
```

### Workspace

```bash
./imans workspace get
./imans workspace get --json
```

### Products

```bash
./imans products list --search shirt
./imans products list --status enabled --page-size 100 --json
./imans products get 123
./imans product-variants list --product-id 123 --all
```

### Sales orders

```bash
./imans sales-orders list --order-date-from 2026-04-01 --order-date-to 2026-04-30
./imans sales-orders list --order-status approved,completed --json
./imans sales-orders get 456
./imans sales-order-items list --order-id 456
./imans sales-order-classifications list
./imans sales-order-classifications get 7 --json
```

## Command Reference

### `auth`

- `imans auth add`
  Stores a token securely, validates it via `workspace get`, caches workspace metadata, and optionally makes the profile active.
- `imans auth test`
  Verifies that the active or selected profile can authenticate successfully.
- `imans auth remove`
  Deletes the saved profile metadata and removes its token from secret storage.

### `profile`

- `imans profile list`
  Shows saved aliases, active marker, base URL, workspace code, and workspace name.
- `imans profile show [name]`
  Shows one saved profile or the active profile.
- `imans profile use <name>`
  Sets the active profile.

### `workspace`

- `imans workspace get`
  Reads the current workspace self endpoint.

### `products`

- `imans products list`
  Filters: `--search`, `--status`, `--category-id`, `--brand-id`, `--is-variable`
- `imans products get <id>`
  Retrieves one product with its variant list.

### `product-variants`

- `imans product-variants list`
  Filters: `--search`, `--product-id`, `--status`, `--is-bundle`

### `sales-orders`

- `imans sales-orders list`
  Filters: `--order-date-from`, `--order-date-to`, `--order-status`, `--classification-id`, `--customer-id`, `--sales-agent-id`, `--search`
- `imans sales-orders get <id>`
  Retrieves one sales order.

### `sales-order-items`

- `imans sales-order-items list`
  Filters: `--order-id`, `--product-id`

### `sales-order-classifications`

- `imans sales-order-classifications list`
- `imans sales-order-classifications get <id>`

## Profile and Secret Storage

Profile metadata is stored in your user config directory as YAML.

Examples:

- Linux: `~/.config/imans/config.yaml`
- macOS: `~/Library/Application Support/imans/config.yaml`
- Windows: `%AppData%\imans\config.yaml`

The config file stores metadata like:

- active profile name
- base URL
- cached workspace code and name
- default output preference

It does not store raw API tokens.

Token storage behavior:

- macOS: OS keychain backend
- Windows: OS credential backend
- Linux: secure keyring backend when available
- fallback: development-only insecure file backend if `IMANS_INSECURE_FILE_SECRETS=1`

Linux is intended to fail closed by default if no secure backend is available.

## Compatibility and Versioning

- `imans version` shows CLI build metadata
- the CLI can warn when the server contract version differs from the CLI schema version
- schema refresh is driven from the public schema endpoint

Refresh the local schema artifact with:

```bash
./scripts/refresh-schema.sh
```

Override the source URL if needed:

```bash
IMANS_SCHEMA_URL=http://127.0.0.1:8000/documentation/v1/schema/ ./scripts/refresh-schema.sh
```

## Completion

The Cobra completion command is available out of the box:

```bash
./imans completion bash
./imans completion zsh
./imans completion fish
./imans completion powershell
```

## Release Notes

The repository includes:

- GitHub Actions CI for test and cross-build checks
- GoReleaser config for archives and checksums
- Homebrew and Scoop release definitions

The release pipeline is scaffolded, but availability of public packages depends on published release artifacts and target repositories.

## Security Notes

- prefer `--token-env` or `--token-stdin` over putting secrets directly on the command line
- do not commit tokens or shell history containing raw secrets
- `--debug` prints request method, URL, status, and latency, but should not print the authorization header

## Status

This is the first usable foundation of the CLI: secure profiles, read-only API coverage for the main v1 domains, build/test scaffolding, and release automation scaffolding.
