#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_DIR="$ROOT_DIR/apps/frontend/out"
TARGET_DIR="$ROOT_DIR/apps/backend/internal/frontend/web"

source "$ROOT_DIR/scripts/lib/cli.sh"

copy_frontend() {
  rm -rf -- "$TARGET_DIR"
  mkdir -p "$TARGET_DIR"
  cp -a "$SOURCE_DIR"/. "$TARGET_DIR/"
  find "$TARGET_DIR" -type d -empty -delete
}

cli_init "Embed frontend assets" "local" 1
cli_header
cli_roadmap "Copy static frontend into backend"
run_step "Copy static frontend into backend" copy_frontend

printf '\n%sEmbedded frontend%s\n' "$CLI_BOLD$CLI_CYAN" "$CLI_RESET"
cli_artifact "$TARGET_DIR"
cli_footer
