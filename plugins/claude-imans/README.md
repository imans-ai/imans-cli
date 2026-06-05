# Claude Imans Plugin

This Claude Code plugin packages the Imans CLI skill for Claude plugin marketplaces.

## Contents

- `.claude-plugin/plugin.json`: Claude plugin manifest.
- `skills/imans/SKILL.md`: Agent Skill that teaches Claude Code how to use `imans`.

## Test Locally

```bash
claude plugin validate ./plugins/claude-imans
claude --plugin-dir ./plugins/claude-imans
```

Inside Claude Code, invoke the skill as:

```text
/imans:imans
```

## Submit

Submit the plugin repository to the Claude community plugin review form:

- https://claude.ai/settings/plugins/submit
- https://platform.claude.com/plugins/submit

Approved community plugins are published through the `anthropics/claude-plugins-community` catalog.
