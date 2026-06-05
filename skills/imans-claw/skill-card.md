## Description: <br>
Use the Imans CLI from OpenClaw agents to query Imans workspace, catalog, and sales order data as JSON. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[imans-ai](https://clawhub.ai/user/imans-ai) <br>

### License/Terms of Use: <br>
Apache-2.0 for the Imans CLI repository. <br>

## Use Case: <br>
Developers, operators, and personal AI agents use this skill to set up and run `imans` commands for read-only Imans workspace, catalog, product, product variant, sales order, order item, and classification workflows from an OpenClaw agent session. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Imans API tokens and workspace access can expose sensitive business data, including catalog, customer-related order, and sales information. <br>
Mitigation: Use `imans login`, `--token-stdin`, or `--token-env`; avoid pasting tokens into chat or passing tokens directly on the command line. <br>
Risk: Broad list commands with `--all --json` may expose large datasets in chat logs or group messaging channels. <br>
Mitigation: Prefer narrow filters, summarize results by default, and confirm before exporting or displaying large raw JSON payloads. <br>
Risk: OpenClaw sandboxed agents may not have the `imans` binary or saved profile available inside the execution environment. <br>
Mitigation: Install `imans` and authenticate in the same host or sandbox where commands are executed. <br>

## Reference(s): <br>
- [Imans CLI repository](https://github.com/imans-ai/imans-cli) <br>
- [Imans CLI releases](https://github.com/imans-ai/imans-cli/releases) <br>
- [OpenClaw skills documentation](https://docs.openclaw.ai/tools/skills) <br>

## Skill Output: <br>
**Output Type(s):** [guidance, shell commands, JSON workflow instructions] <br>
**Output Format:** [Markdown with inline shell commands] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Emphasizes `--json`, profile-aware command usage, secure token setup, and compact summaries for chat surfaces.] <br>

## Skill Version(s): <br>
0.1.0 (source: Imans CLI first public release) <br>

## Ethical Considerations: <br>
Users should verify that they are authorized to access the target Imans workspace, avoid exposing sensitive business data in shared chats or logs, review generated commands before running them, and follow their organization's security and compliance requirements. <br>
