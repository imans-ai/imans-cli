# Imans CLI Skill for NanoClaw

This is a NanoClaw-ready Claude Code skill for using the official `imans` CLI inside NanoClaw agent environments.

## Files

- `SKILL.md`: NanoClaw/Claude Code skill instructions.

## Runtime Requirement

Install and authenticate `imans` in the same container or host where NanoClaw runs commands:

```bash
curl -fsSL https://imans.ai/install | bash
imans version
imans login
imans auth test --quiet
```

For automation, prefer `imans login --token-env IMANS_TOKEN` or `imans login --token-stdin < token.txt`.

## Publish to NanoClaw

Submit this skill through the NanoClaw skills repository flow:

1. Fork the NanoClaw skills repository.
2. Copy `registry/nanoclaw/imans/SKILL.md` into the target skill directory expected by that repository.
3. Include this README or adapt it to the marketplace format.
4. Open a pull request with examples and security notes.
