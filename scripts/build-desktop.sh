#!/bin/bash

set -e


ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

VERSION=$(node -p "require('$ROOT_DIR/package.json').version")

echo "Building version: $VERSION"

DESKTOP_ASSETS="$ROOT_DIR/apps/desktop/assets"
RELEASE_DIR="$ROOT_DIR/releases/desktop"
BUILDER_OUTPUT="$RELEASE_DIR/build"
SERVER_DIR="$ROOT_DIR/apps/backend/cmd/server"


echo "Preparing desktop release"

mkdir -p "$RELEASE_DIR"
rm -rf "$BUILDER_OUTPUT"
rm -f "$RELEASE_DIR"/*.AppImage
mkdir -p "$DESKTOP_ASSETS"

if [[ ! -d "$ROOT_DIR/node_modules" ]]; then
  echo "Installing dependencies"
  cd "$ROOT_DIR"
  yarn install --immutable
fi

echo "Building contracts"

cd "$ROOT_DIR/packages/contracts"

yarn build


echo "Building components"

cd "$ROOT_DIR/packages/ui-components"

yarn build


echo "Building SDK"

cd "$ROOT_DIR/packages/sdk"

yarn build


echo "Building frontend"

cd "$ROOT_DIR/apps/frontend"

VERSION=$VERSION yarn build


echo "Building backend"

cd "$ROOT_DIR/apps/backend"

go build \
-ldflags "-X lwn-simulator-backend/version.AppVersion=$VERSION" \
-o "$DESKTOP_ASSETS/lwn-server" \
cmd/server/main.go


echo "Building launcher"

cd "$ROOT_DIR/apps/launcher"

VERSION=$VERSION yarn build


echo "Building Electron"

cd "$ROOT_DIR/apps/desktop"

VERSION=$VERSION yarn build


echo "Copying release"

cp "$BUILDER_OUTPUT"/*.AppImage "$RELEASE_DIR/"
rm -rf "$BUILDER_OUTPUT"


echo "Desktop release created"

ls -lah "$RELEASE_DIR"