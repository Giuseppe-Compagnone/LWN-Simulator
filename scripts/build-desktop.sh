#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="$(node -p "require('$ROOT_DIR/package.json').version")"

source "$ROOT_DIR/scripts/lib/cli.sh"
source "$ROOT_DIR/scripts/lib/build.sh"

DESKTOP_ASSETS="$ROOT_DIR/apps/desktop/assets"
RELEASE_DIR="$ROOT_DIR/releases/desktop"
BUILDER_OUTPUT="$RELEASE_DIR/build"

prepare_release() {
  mkdir -p "$RELEASE_DIR"
  rm -rf -- "$BUILDER_OUTPUT"
  rm -f -- "$RELEASE_DIR"/*.AppImage
  mkdir -p "$DESKTOP_ASSETS"
}

build_launcher() {
  run_step "Build desktop launcher" run_in_dir "$ROOT_DIR/apps/launcher" env VERSION="$VERSION" yarn build
}

build_electron() {
  run_step "Package Electron AppImage" run_in_dir "$ROOT_DIR/apps/desktop" env VERSION="$VERSION" yarn build
}

copy_release() {
  local -a artifacts=("$BUILDER_OUTPUT"/*.AppImage)

  if [[ ! -f "${artifacts[0]}" ]]; then
    echo "No AppImage was generated in $BUILDER_OUTPUT" >&2
    return 1
  fi

  cp -- "${artifacts[0]}" "$RELEASE_DIR/"
  rm -rf -- "$BUILDER_OUTPUT"
}

cli_init "LWN-Simulator desktop release" "$VERSION" 10
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
  "Publish release artifact"

run_step "Prepare release directories" prepare_release
ensure_dependencies
build_contracts
build_components
build_sdk
build_frontend
build_backend "$DESKTOP_ASSETS/lwn-server"
build_launcher
build_electron
run_step "Publish release artifact" copy_release

printf '\n%sRelease artifacts%s\n' "$CLI_BOLD$CLI_CYAN" "$CLI_RESET"
cli_artifact "$RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
cli_footer
