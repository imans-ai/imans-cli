#!/usr/bin/env bash
set -euo pipefail

SCHEMA_URL="${IMANS_SCHEMA_URL:-https://api.imans.ai/documentation/v1/schema/}"
TARGET="api/openapi/schema.json"

curl -fsSL "$SCHEMA_URL" -o "$TARGET"
printf 'Wrote %s from %s\n' "$TARGET" "$SCHEMA_URL"
