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
  rm -rf -- "$SERVER_RELEASE_DIR"
  mkdir -p "$SERVER_RELEASE_DIR"

  mkdir -p "$DESKTOP_RELEASE_DIR"
  rm -rf -- "$BUILDER_OUTPUT"
  rm -f -- "$DESKTOP_RELEASE_DIR"/*.AppImage
  mkdir -p "$DESKTOP_ASSETS"
}

publish_server() {
  cp -- "$DESKTOP_ASSETS/lwn-server" "$SERVER_RELEASE_DIR/lwn-server"
}

build_launcher() {
  run_step "Build desktop launcher" run_in_dir "$ROOT_DIR/apps/launcher" env VERSION="$VERSION" yarn build
}

build_electron() {
  run_step "Package Electron AppImage" run_in_dir "$ROOT_DIR/apps/desktop" env VERSION="$VERSION" yarn build
}

publish_desktop() {
  local -a artifacts=("$BUILDER_OUTPUT"/*.AppImage)

  if [[ ! -f "${artifacts[0]}" ]]; then
    echo "No AppImage was generated in $BUILDER_OUTPUT" >&2
    return 1
  fi

  cp -- "${artifacts[0]}" "$DESKTOP_RELEASE_DIR/"
  rm -rf -- "$BUILDER_OUTPUT"
}

verify_releases() {
  test -x "$SERVER_RELEASE_DIR/lwn-server"
  test -f "$DESKTOP_RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
}

cli_init "LWN-Simulator complete release" "$VERSION" 12
cli_header
cli_roadmap \
  "Prepare release directories" \
  "Verify workspace dependencies" \
  "Generate API contracts" \
  "Build UI components" \
  "Build SDK" \
  "Build frontend" \
  "Build backend" \
  "Publish server binary" \
  "Build desktop launcher" \
  "Package Electron AppImage" \
  "Publish AppImage" \
  "Verify release artifacts"

run_step "Prepare release directories" prepare_releases
ensure_dependencies
build_contracts
build_components
build_sdk
build_frontend
build_backend "$DESKTOP_ASSETS/lwn-server"
run_step "Publish server binary" publish_server
build_launcher
build_electron
run_step "Publish AppImage" publish_desktop
run_step "Verify release artifacts" verify_releases

printf '\n%sRelease artifacts%s\n' "$CLI_BOLD$CLI_CYAN" "$CLI_RESET"
cli_artifact "$SERVER_RELEASE_DIR/lwn-server"
cli_artifact "$DESKTOP_RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
cli_footer
