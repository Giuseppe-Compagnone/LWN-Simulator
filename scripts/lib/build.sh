#!/usr/bin/env bash

set -Eeuo pipefail

ensure_dependencies() {
  run_step "Install and verify workspace dependencies" \
    run_in_dir "$ROOT_DIR" yarn install --immutable
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

lwn_electron_arch() {
  case "${LWN_TARGET_ARCH:-x64}" in
    x64|arm64|armv7l)
      printf '%s\n' "${LWN_TARGET_ARCH:-x64}"
      ;;
    *)
      printf 'Unsupported Linux target architecture: %s\n' "${LWN_TARGET_ARCH:-}" >&2
      return 1
      ;;
  esac
}

lwn_go_arch() {
  case "${LWN_TARGET_ARCH:-x64}" in
    x64)
      printf 'amd64\n'
      ;;
    arm64)
      printf 'arm64\n'
      ;;
    armv7l)
      printf 'arm\n'
      ;;
    *)
      printf 'Unsupported Linux target architecture: %s\n' "${LWN_TARGET_ARCH:-}" >&2
      return 1
      ;;
  esac
}

electron_target_files() {
  local output_dir="$1"
  local target="$2"
  local -a artifacts

  shopt -s nullglob

  case "$target" in
    AppImage)
      artifacts=("$output_dir"/*.AppImage)
      ;;
    deb)
      artifacts=("$output_dir"/*.deb)
      ;;
    rpm)
      artifacts=("$output_dir"/*.rpm)
      ;;
    *)
      printf 'Unsupported Electron Linux target: %s\n' "$target" >&2
      shopt -u nullglob
      return 1
      ;;
  esac

  if (( ${#artifacts[@]} != 1 )); then
    printf 'Expected exactly one %s artifact in %s, found %d\n' \
      "$target" "$output_dir" "${#artifacts[@]}" >&2
    shopt -u nullglob
    return 1
  fi

  if ! test -s "${artifacts[0]}"; then
    shopt -u nullglob
    return 1
  fi

  shopt -u nullglob
  printf '%s\n' "${artifacts[0]}"
}

publish_electron_artifacts() {
  local output_dir="$1"
  local release_dir="$2"
  local raw_targets="${LWN_ELECTRON_TARGETS:-AppImage}"
  local -a targets
  local target source filename mode

  IFS=',' read -r -a targets <<< "$raw_targets"
  mkdir -p -- "$release_dir"
  shopt -s nullglob

  for target in "${targets[@]}"; do
    target="${target//[[:space:]]/}"
    source="$(electron_target_files "$output_dir" "$target")"
    filename="$(basename "$source")"
    mode=0644
    [[ "$target" == AppImage ]] && mode=0755

    case "$target" in
      AppImage) rm -f -- "$release_dir"/*.AppImage ;;
      deb) rm -f -- "$release_dir"/*.deb ;;
      rpm) rm -f -- "$release_dir"/*.rpm ;;
    esac

    install -m "$mode" -- "$source" "$release_dir/$filename.tmp"
    mv -- "$release_dir/$filename.tmp" "$release_dir/$filename"
  done

  shopt -u nullglob
  rm -rf -- "$output_dir"
}

verify_electron_artifacts() {
  local release_dir="$1"
  local raw_targets="${LWN_ELECTRON_TARGETS:-AppImage}"
  local -a targets
  local target

  IFS=',' read -r -a targets <<< "$raw_targets"
  for target in "${targets[@]}"; do
    target="${target//[[:space:]]/}"
    electron_target_files "$release_dir" "$target" >/dev/null
  done
}

build_backend() {
  local output_path="$1"
  local go_arch
  local -a go_env

  go_arch="$(lwn_go_arch)"
  go_env=(GOOS=linux GOARCH="$go_arch" CGO_ENABLED=0)
  if [[ "${LWN_TARGET_ARCH:-x64}" == armv7l ]]; then
    go_env+=(GOARM=7)
  fi

  run_step "Build backend ($go_arch)" run_in_dir "$ROOT_DIR/apps/backend" env \
    "${go_env[@]}" \
    go build \
    -ldflags "-X lwn-simulator-backend/version.AppVersion=$VERSION" \
    -o "$output_path" \
    cmd/server/main.go
}
