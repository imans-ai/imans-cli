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

imans login

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

### Quick install (macOS / Linux)

```bash
curl -fsSL https://imans.ai/install | bash
```

Downloads the right build for your OS/architecture from GitHub Releases,
verifies its SHA-256 checksum, and installs the `imans` binary to
`/usr/local/bin` (or `~/.local/bin` if that is not writable). Override with
`IMANS_VERSION`, `IMANS_INSTALL_DIR`, or `IMANS_RELEASE_BASE` (mirror).

### Homebrew (macOS / Linux)

```bash
brew install imans-ai/tap/imans
```

### Scoop (Windows)

```powershell
scoop bucket add imans https://github.com/imans-ai/scoop-bucket
scoop install imans
```

### Direct download

Grab a prebuilt archive for your platform from the
[Releases page](https://github.com/imans-ai/imans-cli/releases), verify it
against `checksums.txt`, extract `imans`, and put it on your `PATH`.

### From source

```bash
make build   # produces ./imans in the repo root
```

Then verify:

```bash
imans version
```

## Agent Skills

This repository includes skills that teach AI agents how to use the `imans` CLI
for workspace, catalog, and sales order workflows.

Install the general skill with the skills CLI:

```bash
npx skills add imans-ai/imans-cli
```

Install the OpenClaw-specific skill after it is published to ClawHub:

```bash
openclaw skills install imans-claw
```

For local testing from a checkout, install the skill directory directly:

```bash
openclaw skills install ./skills/imans-claw --as imans-claw
```

The skills live in:

- `skills/imans/SKILL.md`
- `skills/imans-claw/SKILL.md`
- `skills/imans-claw/skill-card.md`

Maintainers can publish the OpenClaw skill to ClawHub under the `imans-ai`
owner:

```bash
clawhub login
clawhub sync --dry-run --owner imans-ai
clawhub sync --all --owner imans-ai
```

### Other Agent Platforms

The repo also includes submission-ready packages for agent ecosystems that need
platform-specific artifacts:

| Platform | Status | Path / action |
|---|---|---|
| Claude Code / Claude Cowork | package-ready | `plugins/claude-imans`; validate with `claude plugin validate ./plugins/claude-imans`, then submit through Claude's community plugin form. |
| ZeroClaw | PR-ready | `registry/zeroclaw/imans` plus `registry/zeroclaw/registry-entry.json`; copy into `zeroclaw-labs/zeroclaw-skills` and open a PR. |
| NanoClaw | PR-ready | `registry/nanoclaw/imans` plus `registry/nanoclaw/marketplace-entry.json`; copy into the NanoClaw skills marketplace format and open a PR. |
| Hermes Agent | covered | Uses Agent Skills, GitHub taps, `skills.sh`, and ClawHub-compatible sources; use `skills/imans/SKILL.md` or the published `skills.sh` repo. |
| PicoClaw | covered | Supports ClawHub/GitHub-style skills; use the published `imans-claw` ClawHub skill or this repo's `skills/` folder. |
| Manus | manual import | Supports Agent Skills, but no public third-party publishing flow was found; import/adapt `skills/imans/SKILL.md` in a Manus workspace/team. |
| Perplexity Computer | no hub | No public third-party skill/plugin publishing surface was found. |
| TrustClaw | no hub | No TrustClaw-native skill registry was found; integrations appear to go through Composio toolkits instead. |

Claude community plugin submission:

```bash
claude plugin validate ./plugins/claude-imans
claude --plugin-dir ./plugins/claude-imans
```

Then submit through one of:

- `https://claude.ai/settings/plugins/submit`
- `https://platform.claude.com/plugins/submit`

ZeroClaw submission:

```bash
# In a fork of zeroclaw-labs/zeroclaw-skills
mkdir -p skills/imans
cp -R /path/to/imans-cli/registry/zeroclaw/imans/* skills/imans/
# Add registry/zeroclaw/registry-entry.json to that repo's registry.json shape.
```

NanoClaw submission:

```bash
# In a fork of the NanoClaw skills marketplace
cp -R /path/to/imans-cli/registry/nanoclaw/imans ./imans
# Adapt registry/nanoclaw/marketplace-entry.json to the target marketplace schema.
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

### Fastest path: `imans login`

For most users this is the only onboarding step:

```bash
imans login
```

It prompts you to paste your API token, validates it against the workspace
endpoint, stores it securely, and makes the resulting profile active. On
success you see a confirmation and can start using resource commands
immediately:

```text
✓ Connected to Acme Co (ACME-01)
active profile  acme-01
base url        https://api.imans.ai/
next            imans products list
```

The profile is named automatically from your workspace. Run `imans login`
again with a different workspace token to add another workspace — each is
saved as its own profile and the most recent login becomes active. Switch
between them with `imans profile use <name>`. Pass `--profile <name>` to
choose the name yourself.

You can also supply the token non-interactively with `--token-stdin`,
`--token-env <ENV_VAR_NAME>`, `--token`, or the `IMANS_TOKEN` environment
variable — handy for scripts and CI.

### Advanced: `imans auth add`

`auth add` is the lower-level command for explicitly named profiles. It
requires `--profile` and does not set the profile active unless you pass
`--set-active`:

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
- Linux: Secret Service keyring backend when available and unlocked
- Linux without a keyring (headless servers, containers, or a WSL setup with no
  configured keyring): automatic
  fallback to an encrypted file under the config directory (`secrets.enc`)
- explicit override: development-only plaintext file backend if
  `IMANS_INSECURE_FILE_SECRETS=1`

The Linux encrypted-file fallback keeps onboarding working out of the box where
no OS keyring exists. The file is AES-256-GCM encrypted with a key derived from
the machine and user identity, so it is never stored in plaintext and is not
portable to other machines or users. It is not a substitute for an OS keyring:
anything able to run as your user on your machine can recompute the key. Where a
keyring is available it is always preferred.

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
