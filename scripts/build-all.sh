#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="$(node -p "require('$ROOT_DIR/package.json').version")"

source "$ROOT_DIR/scripts/lib/cli.sh"
source "$ROOT_DIR/scripts/lib/build.sh"

DESKTOP_ASSETS="$ROOT_DIR/apps/desktop/assets"
DESKTOP_RELEASE_DIR="$ROOT_DIR/releases/desktop"
BUILDER_OUTPUT="$DESKTOP_RELEASE_DIR/build"
SERVER_RELEASE_DIR="$ROOT_DIR/releases/server"

prepare_releases() {
  mkdir -p "$SERVER_RELEASE_DIR"

  mkdir -p "$DESKTOP_RELEASE_DIR"
  rm -rf -- "$BUILDER_OUTPUT"
  mkdir -p "$DESKTOP_ASSETS"
}

build_launcher() {
  run_step "Build desktop launcher" run_in_dir "$ROOT_DIR/apps/launcher" env VERSION="$VERSION" yarn build
}

build_electron() {
  run_step "Package Electron AppImage" run_in_dir "$ROOT_DIR/apps/desktop" env VERSION="$VERSION" yarn build
}

publish_releases() {
  local -a artifacts

  shopt -s nullglob
  artifacts=("$BUILDER_OUTPUT"/*.AppImage)
  shopt -u nullglob

  if (( ${#artifacts[@]} != 1 )); then
    printf 'Expected exactly one AppImage in %s, found %d\n' \
      "$BUILDER_OUTPUT" "${#artifacts[@]}" >&2
    return 1
  fi

  local server_path="$SERVER_RELEASE_DIR/lwn-server"
  local server_stage="$server_path.tmp"
  local desktop_path="$DESKTOP_RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
  local desktop_stage="$desktop_path.tmp"

  install -m 0755 -- "$DESKTOP_ASSETS/lwn-server" "$server_stage"
  if ! install -m 0755 -- "${artifacts[0]}" "$desktop_stage"; then
    rm -f -- "$server_stage"
    return 1
  fi

  mv -- "$server_stage" "$server_path"
  rm -f -- "$DESKTOP_RELEASE_DIR"/*.AppImage
  mv -- "$desktop_stage" "$desktop_path"
  rm -rf -- "$BUILDER_OUTPUT"
}

verify_releases() {
  test -s "$SERVER_RELEASE_DIR/lwn-server"
  test -x "$SERVER_RELEASE_DIR/lwn-server"
  test -s "$DESKTOP_RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
  test -x "$DESKTOP_RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
}

cli_init "LWN-Simulator complete release" "$VERSION" 11
cli_header
cli_roadmap \
  "Prepare release directories" \
  "Verify workspace dependencies" \
  "Generate API contracts" \
  "Build UI components" \
  "Build SDK" \
  "Build frontend" \
  "Build backend" \
  "Build desktop launcher" \
  "Package Electron AppImage" \
  "Publish release artifacts" \
  "Verify release artifacts"

run_step "Prepare release directories" prepare_releases
ensure_dependencies
build_contracts
build_components
build_sdk
build_frontend
build_backend "$DESKTOP_ASSETS/lwn-server"
build_launcher
build_electron
run_step "Publish release artifacts" publish_releases
run_step "Verify release artifacts" verify_releases

printf '\n%sRelease artifacts%s\n' "$CLI_BOLD$CLI_CYAN" "$CLI_RESET"
cli_artifact "$SERVER_RELEASE_DIR/lwn-server"
cli_artifact "$DESKTOP_RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
cli_footer
