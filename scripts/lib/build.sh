#!/usr/bin/env bash

set -Eeuo pipefail

ensure_dependencies() {
  if [[ -d "$ROOT_DIR/node_modules" ]]; then
    run_step "Verify workspace dependencies" run_in_dir "$ROOT_DIR" test -d node_modules
  else
    run_step "Install workspace dependencies" run_in_dir "$ROOT_DIR" yarn install --immutable
  fi
}

build_contracts() {
  run_step "Generate API contracts" run_in_dir "$ROOT_DIR/packages/contracts" yarn build
}

build_components() {
  run_step "Build UI components" run_in_dir "$ROOT_DIR/packages/ui-components" yarn build
}

build_sdk() {
  run_step "Build SDK" run_in_dir "$ROOT_DIR/packages/sdk" yarn build
}

build_frontend() {
  run_step "Build frontend" run_in_dir "$ROOT_DIR/apps/frontend" env VERSION="$VERSION" yarn build
}

build_backend() {
  local output_path="$1"

  run_step "Build backend" run_in_dir "$ROOT_DIR/apps/backend" go build \
    -ldflags "-X lwn-simulator-backend/version.AppVersion=$VERSION" \
    -o "$output_path" \
    cmd/server/main.go
}
