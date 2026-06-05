---
name: imans
description: Use the Imans CLI to query Imans workspace, catalog, and sales order data. Use when the user asks about Imans data, products, variants, orders, workspace metadata, or agent-friendly JSON exports.
homepage: https://github.com/imans-ai/imans-cli
---

# Imans CLI

Use `imans` to read Imans workspace, catalog, and sales order data from the official Imans public API.

## Setup

- Install: `curl -fsSL https://imans.ai/install | bash`
- Homebrew: `brew install imans-ai/tap/imans`
- Windows Scoop: `scoop bucket add imans https://github.com/imans-ai/scoop-bucket && scoop install imans`
- Verify: `imans version`
- Login interactively: `imans login`
- Login for automation: `imans login --token-env IMANS_TOKEN` or `imans login --token-stdin < token.txt`
- Test access: `imans auth test`

## Agent Rules

- Prefer `--json` for data you will parse, compare, summarize, or transform.
- Prefer `--all --json` when the user asks for complete lists.
- Use `--profile <name>` when the user names a workspace/profile; do not switch the active profile unless asked.
- Use `--quiet` in scripts to keep stderr noise low.
- Use `--debug` only for troubleshooting; it prints safe request diagnostics but not tokens.
- Confirm before pasting large JSON or sensitive business data into chat; summarize by default.

## Common Commands

- Workspace metadata: `imans workspace get --json`
- Saved profiles: `imans profile list`
- Active profile details: `imans profile show`
- Products: `imans products list --all --json`
- Product search: `imans products list --search "shirt" --status active --json`
- Product details: `imans products get <id> --json`
- Variants: `imans product-variants list --product-id <product-id> --all --json`
- Sales orders: `imans sales-orders list --order-date-from <yyyy-mm-dd> --order-date-to <yyyy-mm-dd> --all --json`
- Sales order details: `imans sales-orders get <id> --json`
- Sales order items: `imans sales-order-items list --order-id <order-id> --all --json`
- Sales order classifications: `imans sales-order-classifications list --all --json`

## Useful Workflows

- Understand the connected workspace: run `imans workspace get --json`, then summarize name, code, status, and relevant settings.
- Find products: run `imans products list --search "<query>" --json`; use `--all` only if the user needs exhaustive results.
- Inspect a product: run `imans products get <id> --json`; include variant details only when relevant.
- Analyze sales activity: run `imans sales-orders list` with date/status filters and `--json`; avoid unbounded order exports unless requested.
- Drill into an order: run `imans sales-orders get <id> --json`, then `imans sales-order-items list --order-id <id> --json` if line items are needed.

## Security

- Do not ask users to paste tokens into chat. Tell them to use `imans login`, `--token-stdin`, or `--token-env`.
- Avoid `imans login --token <token>` on shared machines because tokens can land in shell history.
- Tokens are stored by the CLI in the OS keychain or encrypted local fallback, not in the config file.
- Treat workspace, product, customer, and order data as business-sensitive.

## Troubleshooting

- `0`: success.
- `2`: usage error; inspect flags and required arguments.
- `3`: authentication failed; run `imans auth test` or login again.
- `4`: token lacks scope for the resource.
- `5`: resource not found.
- `6`: network problem reaching the API.
- `7`: API server error; include the CLI trace ID when reporting to support.
