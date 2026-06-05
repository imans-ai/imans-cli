# Imans CLI Skill for ZeroClaw

This skill teaches ZeroClaw agents how to use the official `imans` CLI for read-only Imans workspace, catalog, and sales order workflows.

## Installation Requirement

The `imans` binary must be available where ZeroClaw executes shell commands.

```bash
curl -fsSL https://imans.ai/install | bash
imans version
imans login
imans auth test --quiet
```

## Permissions

- `shell_exec`: required to run the `imans` CLI.

The skill does not require write permissions to the Imans API. The current CLI command surface is read-only for workspace, catalog, and sales order resources.

## Example Prompts

- "Show me the active Imans workspace."
- "Find products matching shirt in Imans."
- "Summarize sales orders from 2026-04-01 to 2026-04-30."
- "Get the line items for sales order 456."

## Publish to ZeroClaw

Submit this folder to `zeroclaw-labs/zeroclaw-skills`:

1. Fork `https://github.com/zeroclaw-labs/zeroclaw-skills`.
2. Copy this folder to `skills/imans` in that fork.
3. Add a matching `registry.json` entry for `imans`.
4. Open a pull request.
