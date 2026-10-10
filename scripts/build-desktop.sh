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
  mkdir -p "$DESKTOP_ASSETS"
}

build_launcher() {
  run_step "Build desktop launcher" run_in_dir "$ROOT_DIR/apps/launcher" env VERSION="$VERSION" yarn build
}

build_electron() {
  run_step "Package Electron AppImage" run_in_dir "$ROOT_DIR/apps/desktop" env VERSION="$VERSION" yarn build
}

copy_release() {
  local -a artifacts

  shopt -s nullglob
  artifacts=("$BUILDER_OUTPUT"/*.AppImage)
  shopt -u nullglob

  if (( ${#artifacts[@]} != 1 )); then
    printf 'Expected exactly one AppImage in %s, found %d\n' \
      "$BUILDER_OUTPUT" "${#artifacts[@]}" >&2
    return 1
  fi

  local published_path="$RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
  local staged_path="$published_path.tmp"

  install -m 0755 -- "${artifacts[0]}" "$staged_path"
  rm -f -- "$RELEASE_DIR"/*.AppImage
  mv -- "$staged_path" "$published_path"
  rm -rf -- "$BUILDER_OUTPUT"
}

verify_release() {
  test -s "$RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
  test -x "$RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
}

cli_init "LWN-Simulator desktop release" "$VERSION" 11
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
  "Publish release artifact" \
  "Verify release artifact"

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
run_step "Verify release artifact" verify_release

printf '\n%sRelease artifacts%s\n' "$CLI_BOLD$CLI_CYAN" "$CLI_RESET"
cli_artifact "$RELEASE_DIR/LWN-Simulator-$VERSION.AppImage"
cli_footer
